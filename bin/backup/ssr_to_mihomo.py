#!/usr/bin/env python3
import base64
import json
import pathlib
import re
import sys
import urllib.parse
import urllib.request


def b64decode_urlsafe(s: str) -> str:
    s = s.strip()
    s = urllib.parse.unquote(s)
    s = s.replace('-', '+').replace('_', '/')
    s += '=' * (-len(s) % 4)
    return base64.b64decode(s).decode('utf-8', errors='ignore')


def yaml_quote(value) -> str:
    return json.dumps('' if value is None else str(value), ensure_ascii=False)


def fetch_subscription(url: str) -> str:
    req = urllib.request.Request(
        url,
        headers={'User-Agent': 'ClashforWindows/0.20 MihomoBox/0.1'},
    )
    with urllib.request.urlopen(req, timeout=30) as r:
        raw = r.read().decode('utf-8', errors='ignore')

    if 'ssr://' in raw:
        return raw

    try:
        decoded = b64decode_urlsafe(raw)
        if 'ssr://' in decoded:
            return decoded
    except Exception:
        pass

    return raw


def parse_ssr(uri: str):
    if not uri.startswith('ssr://'):
        return None

    try:
        decoded = b64decode_urlsafe(uri[6:])
    except Exception:
        return None

    main, _, query = decoded.partition('/?')
    parts = main.split(':')
    if len(parts) < 6:
        return None

    server, port, protocol, method, obfs, password_b64 = parts[:6]
    params = urllib.parse.parse_qs(query)

    def get_param(name: str) -> str:
        value = params.get(name, [''])[0]
        if not value:
            return ''
        try:
            return b64decode_urlsafe(value)
        except Exception:
            return urllib.parse.unquote(value)

    try:
        password = b64decode_urlsafe(password_b64)
    except Exception:
        password = password_b64

    remarks = get_param('remarks') or f'{server}:{port}'
    obfs_param = get_param('obfsparam')
    protocol_param = get_param('protoparam')

    node = {
        'name': remarks,
        'server': server,
        'port': int(port),
        'cipher': method,
        'password': password,
        'protocol': protocol,
        'obfs': obfs,
        'udp': True,
    }
    if obfs_param:
        node['obfs-param'] = obfs_param
    if protocol_param:
        node['protocol-param'] = protocol_param
    return node


def unique_names(nodes):
    seen = {}
    for n in nodes:
        base = n['name']
        if base not in seen:
            seen[base] = 1
        else:
            seen[base] += 1
            n['name'] = f'{base} #{seen[base]}'
    return nodes


def render_config(nodes):
    names = [n['name'] for n in nodes]
    lines = [
        'mixed-port: 7890',
        'allow-lan: false',
        'bind-address: 127.0.0.1',
        'mode: rule',
        'log-level: info',
        'ipv6: false',
        'external-controller: 127.0.0.1:9090',
        'secret: ""',
        '',
        'profile:',
        '  store-selected: true',
        '  store-fake-ip: true',
        '',
        'dns:',
        '  enable: true',
        '  listen: 127.0.0.1:1053',
        '  ipv6: false',
        '  enhanced-mode: fake-ip',
        '  fake-ip-range: 198.18.0.1/16',
        '  nameserver:',
        '    - 1.1.1.1',
        '    - 8.8.8.8',
        '',
        'proxies:',
    ]

    for n in nodes:
        lines += [
            f'  - name: {yaml_quote(n["name"])}',
            '    type: ssr',
            f'    server: {yaml_quote(n["server"])}',
            f'    port: {n["port"]}',
            f'    cipher: {yaml_quote(n["cipher"])}',
            f'    password: {yaml_quote(n["password"])}',
            f'    protocol: {yaml_quote(n["protocol"])}',
            f'    obfs: {yaml_quote(n["obfs"])}',
            '    udp: true',
        ]
        if n.get('obfs-param'):
            lines.append(f'    obfs-param: {yaml_quote(n["obfs-param"])}')
        if n.get('protocol-param'):
            lines.append(f'    protocol-param: {yaml_quote(n["protocol-param"])}')

    lines += [
        '',
        'proxy-groups:',
        '  - name: AUTO',
        '    type: url-test',
        '    url: http://www.gstatic.com/generate_204',
        '    interval: 300',
        '    tolerance: 80',
        '    proxies:',
    ]
    for name in names:
        lines.append(f'      - {yaml_quote(name)}')

    lines += [
        '',
        '  - name: PROXY',
        '    type: select',
        '    proxies:',
        '      - AUTO',
    ]
    for name in names:
        lines.append(f'      - {yaml_quote(name)}')

    lines += [
        '      - DIRECT',
        '',
        'rules:',
        '  - DOMAIN-SUFFIX,local,DIRECT',
        '  - IP-CIDR,127.0.0.0/8,DIRECT,no-resolve',
        '  - IP-CIDR,10.0.0.0/8,DIRECT,no-resolve',
        '  - IP-CIDR,172.16.0.0/12,DIRECT,no-resolve',
        '  - IP-CIDR,192.168.0.0/16,DIRECT,no-resolve',
        '  - IP-CIDR,100.64.0.0/10,DIRECT,no-resolve',
        '  - MATCH,PROXY',
        '',
    ]
    return '\n'.join(lines)


def main():
    if len(sys.argv) != 3:
        print('Usage: ssr_to_mihomo.py <subscription_url> <output_config.yaml>')
        sys.exit(1)

    sub_url = sys.argv[1]
    out_file = pathlib.Path(sys.argv[2]).expanduser()
    text = fetch_subscription(sub_url)
    uris = re.findall(r'ssr://[A-Za-z0-9_\-+/=]+', text)
    nodes = [parse_ssr(u) for u in uris]
    nodes = unique_names([n for n in nodes if n])

    if not nodes:
        print('没有解析到 SSR 节点。请确认订阅地址有效，并且返回 ssr:// 节点。')
        sys.exit(2)

    out_file.parent.mkdir(parents=True, exist_ok=True)
    out_file.write_text(render_config(nodes), encoding='utf-8')
    print(f'已生成配置: {out_file}')
    print(f'节点数量: {len(nodes)}')


if __name__ == '__main__':
    main()
