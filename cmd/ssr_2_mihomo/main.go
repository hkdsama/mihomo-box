package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type SSRNode struct {
	Name          string
	Server        string
	Port          int
	Cipher        string
	Password      string
	Protocol      string
	Obfs          string
	ObfsParam     string
	ProtocolParam string
}

func main() {
	// merge subcommand: merge config_base.yaml + proxies.yaml → config.yaml
	if len(os.Args) >= 2 && os.Args[1] == "merge" {
		mergeCLI(os.Args[2:])
		return
	}

	var subURL string
	var inputFile string
	var proxiesOut string
	var baseFile string
	var mergedOut string
	var groupName string
	var listOnly bool

	flag.StringVar(&subURL, "url", "", "SSR subscription URL")
	flag.StringVar(&inputFile, "input", "", "local subscription text file, optional")
	flag.StringVar(&proxiesOut, "proxies-out", "./config/proxies.yaml", "output file for proxies+proxy-groups section")
	flag.StringVar(&baseFile, "base", "./config/config_base.yaml", "base config file (merged with proxies to produce final config)")
	flag.StringVar(&mergedOut, "output", "./config/config.yaml", "final merged config output")
	flag.StringVar(&groupName, "group", "PROXY", "mihomo select proxy group name")
	flag.BoolVar(&listOnly, "list", false, "only list parsed node names, do not write config")
	flag.Parse()

	if subURL == "" && inputFile == "" {
		fatal("必须提供 --url 或 --input")
	}

	raw, err := loadSubscription(subURL, inputFile)
	if err != nil {
		fatal("读取订阅失败: %v", err)
	}

	nodes, err := parseSubscription(raw)
	if err != nil {
		fatal("解析订阅失败: %v", err)
	}
	if len(nodes) == 0 {
		fatal("没有解析到 SSR 节点。请确认订阅内容里有 ssr:// 节点")
	}

	nodes = uniqueNames(nodes)

	if listOnly {
		for i, n := range nodes {
			fmt.Printf("%3d. %s\n", i+1, n.Name)
		}
		return
	}

	proxiesContent := renderProxiesSection(nodes, groupName)
	if err := os.MkdirAll(filepath.Dir(proxiesOut), 0755); err != nil {
		fatal("创建输出目录失败: %v", err)
	}
	if err := os.WriteFile(proxiesOut, []byte(proxiesContent), 0644); err != nil {
		fatal("写入 proxies 配置失败: %v", err)
	}
	fmt.Printf("已生成 proxies 配置: %s\n", proxiesOut)
	fmt.Printf("节点数量: %d\n", len(nodes))

	// auto-merge if base file exists
	baseContent, err := os.ReadFile(baseFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("提示: 未找到 %s，跳过合并。\n", baseFile)
			fmt.Printf("      请创建该文件后执行: ./bin/ssr_2_mihomo merge\n")
			return
		}
		fatal("读取基础配置失败: %v", err)
	}

	merged := mergeConfigs(string(baseContent), proxiesContent)
	if err := os.MkdirAll(filepath.Dir(mergedOut), 0755); err != nil {
		fatal("创建输出目录失败: %v", err)
	}
	if err := os.WriteFile(mergedOut, []byte(merged), 0644); err != nil {
		fatal("写入最终配置失败: %v", err)
	}
	fmt.Printf("已合并配置: %s\n", mergedOut)
}

func mergeCLI(args []string) {
	fs := flag.NewFlagSet("merge", flag.ExitOnError)
	baseFile := fs.String("base", "./config/config_base.yaml", "base config file")
	proxiesFile := fs.String("proxies", "./config/proxies.yaml", "proxies section file")
	output := fs.String("output", "./config/config.yaml", "merged output file")
	fs.Parse(args)

	baseContent, err := os.ReadFile(*baseFile)
	if err != nil {
		fatal("读取基础配置失败: %v", err)
	}
	proxiesContent, err := os.ReadFile(*proxiesFile)
	if err != nil {
		fatal("读取 proxies 配置失败: %v", err)
	}

	merged := mergeConfigs(string(baseContent), string(proxiesContent))
	if err := os.MkdirAll(filepath.Dir(*output), 0755); err != nil {
		fatal("创建输出目录失败: %v", err)
	}
	if err := os.WriteFile(*output, []byte(merged), 0644); err != nil {
		fatal("写入最终配置失败: %v", err)
	}
	fmt.Printf("已合并配置: %s\n", *output)
}

// mergeConfigs concatenates base config and proxies section.
// The two halves cover disjoint top-level YAML keys so plain
// concatenation produces a valid single-document YAML file.
func mergeConfigs(base, proxies string) string {
	base = strings.TrimRight(base, "\n")
	proxies = strings.TrimLeft(proxies, "\n")
	return base + "\n\n" + proxies
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "错误: "+format+"\n", args...)
	os.Exit(1)
}

func loadSubscription(subURL, inputFile string) (string, error) {
	if inputFile != "" {
		b, err := os.ReadFile(inputFile)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	if strings.Contains(subURL, "ssr://") && !strings.HasPrefix(subURL, "http://") && !strings.HasPrefix(subURL, "https://") {
		return subURL, nil
	}

	urls := []string{subURL}
	if strings.HasPrefix(subURL, "http://") {
		urls = append(urls, "https://"+strings.TrimPrefix(subURL, "http://"))
	}
	userAgents := []string{
		"ClashforWindows/0.20 MihomoBox/0.1",
		"ClashforWindows/0.20",
		"ClashMetaForAndroid/2.10.1",
		"mihomo-box/0.1",
	}

	var lastErr error
	for _, candidateURL := range urls {
		for _, ua := range userAgents {
			text, err := fetchSubscription(candidateURL, ua)
			if err != nil {
				lastErr = err
				continue
			}
			if strings.TrimSpace(text) == "" {
				lastErr = fmt.Errorf("订阅返回空内容: %s", candidateURL)
				continue
			}
			return text, nil
		}
	}
	if lastErr == nil {
		lastErr = errors.New("订阅返回空内容")
	}
	return "", lastErr
}

func fetchSubscription(subURL, userAgent string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, subURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Cache-Control", "no-cache")

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy: nil,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP 状态码 %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if len(body) >= 2 && body[0] == 0x1f && body[1] == 0x8b {
		zr, err := gzip.NewReader(bytes.NewReader(body))
		if err == nil {
			defer zr.Close()
			decoded, err := io.ReadAll(zr)
			if err == nil {
				body = decoded
			}
		}
	}
	return string(body), nil
}

func parseSubscription(raw string) ([]SSRNode, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil, errors.New("订阅内容为空")
	}

	// 有些订阅直接返回多行 ssr://，有些订阅整体再包一层 base64。
	if !strings.Contains(text, "ssr://") {
		decoded, err := decodeBase64Flexible(text)
		if err == nil && strings.Contains(decoded, "ssr://") {
			text = decoded
		}
	}

	uris := extractSSRURIs(text)
	var nodes []SSRNode
	for _, uri := range uris {
		n, err := parseSSRURI(uri)
		if err == nil {
			nodes = append(nodes, n)
		}
	}
	return nodes, nil
}

func extractSSRURIs(text string) []string {
	// SSR 的主体通常是 base64/base64-url 字符。
	re := regexp.MustCompile(`ssr://[A-Za-z0-9_\-+/=]+`)
	matches := re.FindAllString(text, -1)

	seen := map[string]bool{}
	var uris []string
	for _, m := range matches {
		m = strings.TrimSpace(m)
		if m == "" || seen[m] {
			continue
		}
		seen[m] = true
		uris = append(uris, m)
	}
	return uris
}

func parseSSRURI(uri string) (SSRNode, error) {
	if !strings.HasPrefix(uri, "ssr://") {
		return SSRNode{}, errors.New("not ssr uri")
	}

	encoded := strings.TrimPrefix(uri, "ssr://")
	decoded, err := decodeBase64Flexible(encoded)
	if err != nil {
		return SSRNode{}, err
	}

	mainPart, queryPart, _ := strings.Cut(decoded, "/?")
	parts := strings.Split(mainPart, ":")
	if len(parts) < 6 {
		return SSRNode{}, fmt.Errorf("SSR main part 字段不足: %s", decoded)
	}

	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return SSRNode{}, fmt.Errorf("端口无效: %s", parts[1])
	}

	password, err := decodeBase64Flexible(parts[5])
	if err != nil {
		// 少数订阅可能密码不是标准 base64，保留原始值。
		password = parts[5]
	}

	params, _ := url.ParseQuery(queryPart)
	getParam := func(key string) string {
		v := params.Get(key)
		if v == "" {
			return ""
		}
		decoded, err := decodeBase64Flexible(v)
		if err == nil {
			return decoded
		}
		uv, err := url.QueryUnescape(v)
		if err == nil {
			return uv
		}
		return v
	}

	name := getParam("remarks")
	if name == "" {
		name = fmt.Sprintf("%s:%d", parts[0], port)
	}

	return SSRNode{
		Name:          name,
		Server:        parts[0],
		Port:          port,
		Protocol:      parts[2],
		Cipher:        parts[3],
		Obfs:          parts[4],
		Password:      password,
		ObfsParam:     getParam("obfsparam"),
		ProtocolParam: getParam("protoparam"),
	}, nil
}

func decodeBase64Flexible(s string) (string, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, " ", "")

	if unescaped, err := url.QueryUnescape(s); err == nil {
		s = unescaped
	}

	candidates := []string{s}
	if strings.ContainsAny(s, "-_") {
		candidates = append(candidates, strings.NewReplacer("-", "+", "_", "/").Replace(s))
	}

	var lastErr error
	for _, c := range candidates {
		for _, enc := range []*base64.Encoding{
			base64.RawURLEncoding,
			base64.URLEncoding,
			base64.RawStdEncoding,
			base64.StdEncoding,
		} {
			out, err := enc.DecodeString(c)
			if err == nil {
				return string(out), nil
			}
			lastErr = err
		}

		padded := c + strings.Repeat("=", (4-len(c)%4)%4)
		for _, enc := range []*base64.Encoding{
			base64.URLEncoding,
			base64.StdEncoding,
		} {
			out, err := enc.DecodeString(padded)
			if err == nil {
				return string(out), nil
			}
			lastErr = err
		}
	}
	return "", lastErr
}

func uniqueNames(nodes []SSRNode) []SSRNode {
	seen := map[string]int{}
	for i := range nodes {
		base := strings.TrimSpace(nodes[i].Name)
		if base == "" {
			base = fmt.Sprintf("%s:%d", nodes[i].Server, nodes[i].Port)
		}
		seen[base]++
		if seen[base] == 1 {
			nodes[i].Name = base
		} else {
			nodes[i].Name = fmt.Sprintf("%s #%d", base, seen[base])
		}
	}
	return nodes
}

func yq(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// renderProxiesSection generates only the proxies + proxy-groups sections.
// These are auto-generated from the SSR subscription; everything else
// (ports, DNS, rules) lives in config_base.yaml and is not touched here.
func renderProxiesSection(nodes []SSRNode, groupName string) string {
	var b strings.Builder

	sort.SliceStable(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})

	b.WriteString("proxies:\n")
	for _, n := range nodes {
		fmt.Fprintf(&b, "  - name: %s\n", yq(n.Name))
		b.WriteString("    type: ssr\n")
		fmt.Fprintf(&b, "    server: %s\n", yq(n.Server))
		fmt.Fprintf(&b, "    port: %d\n", n.Port)
		fmt.Fprintf(&b, "    cipher: %s\n", yq(n.Cipher))
		fmt.Fprintf(&b, "    password: %s\n", yq(n.Password))
		fmt.Fprintf(&b, "    protocol: %s\n", yq(n.Protocol))
		fmt.Fprintf(&b, "    obfs: %s\n", yq(n.Obfs))
		b.WriteString("    udp: true\n")
		if n.ObfsParam != "" {
			fmt.Fprintf(&b, "    obfs-param: %s\n", yq(n.ObfsParam))
		}
		if n.ProtocolParam != "" {
			fmt.Fprintf(&b, "    protocol-param: %s\n", yq(n.ProtocolParam))
		}
	}
	b.WriteString("\n")

	b.WriteString("proxy-groups:\n")
	b.WriteString("  - name: AUTO\n")
	b.WriteString("    type: url-test\n")
	b.WriteString("    url: http://www.gstatic.com/generate_204\n")
	b.WriteString("    interval: 300\n")
	b.WriteString("    tolerance: 80\n")
	b.WriteString("    proxies:\n")
	for _, n := range nodes {
		fmt.Fprintf(&b, "      - %s\n", yq(n.Name))
	}
	b.WriteString("\n")

	fmt.Fprintf(&b, "  - name: %s\n", yq(groupName))
	b.WriteString("    type: select\n")
	b.WriteString("    proxies:\n")
	b.WriteString("      - AUTO\n")
	for _, n := range nodes {
		fmt.Fprintf(&b, "      - %s\n", yq(n.Name))
	}
	b.WriteString("      - DIRECT\n")

	return b.String()
}
