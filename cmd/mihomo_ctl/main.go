package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
)

type proxyGroup struct {
	Now string   `json:"now"`
	All []string `json:"all"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "list":
		err = list(os.Args[2:])
	case "pick":
		err = pick(os.Args[2:])
	case "payload":
		err = payload(os.Args[2:])
	case "now":
		err = now(os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		err = fmt.Errorf("unknown command: %s", os.Args[1])
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  mihomo_ctl list <proxy-group-json>")
	fmt.Fprintln(os.Stderr, "  mihomo_ctl pick <proxy-group-json> <index>")
	fmt.Fprintln(os.Stderr, "  mihomo_ctl payload <proxy-name>")
	fmt.Fprintln(os.Stderr, "  mihomo_ctl now <proxy-group-json>")
}

func list(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("list requires <proxy-group-json>")
	}
	g, err := readProxyGroup(args[0])
	if err != nil {
		return err
	}
	fmt.Println("当前节点:", g.Now)
	fmt.Println("------------------------------------------------------------")
	for i, name := range g.All {
		fmt.Printf("%2d. %s\n", i+1, name)
	}
	fmt.Println("------------------------------------------------------------")
	fmt.Println(" 0. 返回")
	return nil
}

func pick(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("pick requires <proxy-group-json> <index>")
	}
	g, err := readProxyGroup(args[0])
	if err != nil {
		return err
	}
	idx, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("编号无效: %s", args[1])
	}
	if idx < 1 || idx > len(g.All) {
		return fmt.Errorf("编号超出范围: %d", idx)
	}
	fmt.Println(g.All[idx-1])
	return nil
}

func payload(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("payload requires <proxy-name>")
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]string{"name": args[0]})
}

func now(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("now requires <proxy-group-json>")
	}
	g, err := readProxyGroup(args[0])
	if err != nil {
		return err
	}
	fmt.Println(g.Now)
	return nil
}

func readProxyGroup(path string) (proxyGroup, error) {
	var r io.Reader
	if path == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(path)
		if err != nil {
			return proxyGroup{}, err
		}
		defer f.Close()
		r = f
	}

	var g proxyGroup
	if err := json.NewDecoder(r).Decode(&g); err != nil {
		return proxyGroup{}, err
	}
	return g, nil
}
