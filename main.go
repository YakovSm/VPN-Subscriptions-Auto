package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang"
	"gopkg.in/yaml.v3"
)

// Производительность регулируется этими константами.
const (
	fetchWorkers  = 12
	tcpWorkers    = 512
	tcpTimeout    = 500 * time.Millisecond
	dnsTimeout    = 1500 * time.Millisecond
	httpTimeout   = 10 * time.Second
	maxSourceSize = 64 << 20 // 64 MiB на один источник
)

var (
	geoDb *geoip2.Reader

	httpClient = &http.Client{
		Timeout: httpTimeout,
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          256,
			MaxIdleConnsPerHost:   8,
			MaxConnsPerHost:       8,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   7 * time.Second,
			ResponseHeaderTimeout: 12 * time.Second,
			ExpectContinueTimeout: time.Second,
		},
	}

	invalidFilenameRE = regexp.MustCompile(`[<>:"/\\|?*]`)
	profileTitleRE    = regexp.MustCompile(`(?mi)^#\s*profile-title:\s*(.+)$`)
	dnsCache          sync.Map // map[string]*dnsRecord; DNS выполняется один раз на hostname
)

var sources = []string{
	"https://raw.githubusercontent.com/Xumuk0/White-List-VPN/refs/heads/main/%D0%A0%D0%B5%D1%86%D0%B5%D0%BF%D1%82-%D0%9F%D0%B8%D1%80%D0%BE%D0%B3%D0%BE%D0%B2.txt",
	"https://raw.githubusercontent.com/fifthidea/sub/refs/heads/main/sub/sub-full-plaintxt.txt",
	"https://raw.githubusercontent.com/igareck/vpn-configs-for-russia/refs/heads/main/WHITE-CIDR-RU-all.txt",
	"https://raw.githubusercontent.com/solovyov-jenya2004/all_subs/refs/heads/main/final_sorted",
	"https://raw.githubusercontent.com/svinakraft-maker/FlareFeed/refs/heads/main/public/podpiska.txt",
	"https://raw.githubusercontent.com/HalyavusVPNUS/vless-subscription/refs/heads/main/hysteria.txt",
	"https://raw.githubusercontent.com/cbusifabcap/daily_free_vpn/refs/heads/main/sub/URI.yml",
	"https://raw.githubusercontent.com/nick-core-web/reality-subscriptions/refs/heads/main/all.txt",
	"https://raw.githubusercontent.com/RnRLover/rjsxrd-karing-subscription/refs/heads/main/subscription.json",
	"https://raw.githubusercontent.com/alireza786786/Free-V2ray-Configs/refs/heads/main/all_servers.txt",
	"https://raw.githubusercontent.com/F0rc3Run/F0rc3Run/refs/heads/main/Best-Results/proxies.txt",
	"https://raw.githubusercontent.com/ShadowException/VPN/refs/heads/main/configs/VPN-cat",
	"https://raw.githubusercontent.com/4n0nymou3/multi-proxy-config-fetcher/refs/heads/main/configs/proxy_configs.txt",
	"https://raw.githubusercontent.com/awesome-vpn/awesome-vpn/refs/heads/master/all",
	"https://raw.githubusercontent.com/wuqb2i4f/xray-config-toolkit/refs/heads/main/output/base64/mix-uri",
	"https://raw.githubusercontent.com/Ruk1ng001/freeSub/refs/heads/main/v2ray",
	"https://raw.githubusercontent.com/teknovpnhub/v2ray-subscription/refs/heads/main/servers.txt",
	"https://raw.githubusercontent.com/ripaojiedian/freenode/refs/heads/main/sub",
	"https://gitverse.ru/api/repos/ru-wbl/wl/raw/branch/master/KvRuVPN%2FKvRuVPN.txt",
	"https://gitverse.ru/api/repos/Pizduk/PizdukVPN/raw/branch/master/WlSubPiz.txt",
	"https://gitverse.ru/api/repos/flaafix/AetrisVPN_white_list_lite/raw/branch/master/AetrisVPN.txt",
	"https://www.xrayvip.com/free.txt",
	"https://raw.githubusercontent.com/Mosifree/-FREE2CONFIG/refs/heads/main/Clash_Movaghat",
	"https://raw.githubusercontent.com/iboxz/free-v2ray-collector/refs/heads/main/main/mix.txt",
	"https://raw.githubusercontent.com/Barabama/FreeNodes/refs/heads/main/nodes/merged.txt",
	"https://raw.githubusercontent.com/hamedp-71/Sub_Checker_Creator/refs/heads/main/final.txt",
	"https://raw.githubusercontent.com/rango-cfs/NewCollector/refs/heads/main/v2ray_links.txt",
	"https://raw.githubusercontent.com/mohamadfg-dev/telegram-v2ray-configs-collector/refs/heads/main/category/vless.txt",
	"https://raw.githubusercontent.com/mohamadfg-dev/telegram-v2ray-configs-collector/refs/heads/main/category/trojan.txt",
	"https://raw.githubusercontent.com/LalatinaHub/Mineral/refs/heads/master/result/nodes",
	"https://raw.githubusercontent.com/free18/v2ray/refs/heads/main/v.txt",
	"https://raw.githubusercontent.com/shabane/kamaji/refs/heads/master/hub/best.txt",
	"https://raw.githubusercontent.com/Danialsamadi/v2go/refs/heads/main/AllConfigsSub.txt",
	"https://raw.githubusercontent.com/Firmfox/Proxify/refs/heads/main/v2ray_configs/separated_by_protocol/vless.txt",
	"https://raw.githubusercontent.com/Firmfox/Proxify/refs/heads/main/v2ray_configs/separated_by_protocol/trojan.txt",
	"https://raw.githubusercontent.com/Firmfox/Proxify/refs/heads/main/v2ray_configs/separated_by_protocol/other.txt",
	"https://raw.githubusercontent.com/MatinGhanbari/v2ray-configs/refs/heads/main/subscriptions/v2ray/all_sub.txt",
	"https://raw.githubusercontent.com/itsyebekhe/PSG/refs/heads/main/config.txt",
	"https://raw.githubusercontent.com/barry-far/V2ray-Config/refs/heads/main/All_Configs_Sub.txt",
	"https://raw.githubusercontent.com/mahdibland/V2RayAggregator/refs/heads/master/sub/sub_merge.txt",
	"https://raw.githubusercontent.com/Surfboardv2ray/TGParse/refs/heads/main/splitted/mixed",
	"https://raw.githubusercontent.com/Pawdroid/Free-servers/refs/heads/main/sub",
	"https://openproxylist.com/v2ray/rawlist/subscribe",
	"https://raw.githubusercontent.com/SoliSpirit/v2ray-configs/refs/heads/main/all_configs.txt",
	"https://raw.githubusercontent.com/lagzian/SS-Collector/refs/heads/main/configs/subscription-all.txt",
	"https://raw.githubusercontent.com/mahsanet/MahsaFreeConfig/refs/heads/main/mtn/sub_1.txt",
	"https://raw.githubusercontent.com/roosterkid/openproxylist/refs/heads/main/V2RAY.txt",
	"https://raw.githubusercontent.com/MhdiTaheri/V2rayCollector/refs/heads/main/sub/mix",
	"https://raw.githubusercontent.com/Created-By/Telegram-Eag1e_YT/refs/heads/main/%40Eag1e_YT",
	"https://github.com/Argh94/Proxy-List/raw/refs/heads/main/All_Config.txt",
	"https://raw.githubusercontent.com/histeenn/VLESS-PO-GRIBI/refs/heads/main/deploy/sub.txt",
	"https://raw.githubusercontent.com/nikita29a/FreeProxyList/refs/heads/main/mirror/1.txt",
	"https://raw.githubusercontent.com/sakha1370/OpenRay/refs/heads/main/output/all_valid_proxies.txt",
	"https://raw.githubusercontent.com/VOID-Anonymity/V.O.I.D-VPN_Bypass/refs/heads/main/url_work.txt",
	"https://raw.githubusercontent.com/EEvanescence/4Diana/refs/heads/main/AllConfigsSub.txt",
	"https://raw.githubusercontent.com/iampedii/whitedns-sub/refs/heads/main/base64.txt",
	"https://raw.githubusercontent.com/patterniha/Free-Configs/refs/heads/main/configs.txt",
	"https://robin.victoriacross.ir/",
	"https://raw.githubusercontent.com/aneek0/goida-vpn-configs/refs/heads/main/githubmirror/26.txt",
}

type XrayConfig struct {
	Remarks   string `json:"remarks"`
	Outbounds []struct {
		Protocol string `json:"protocol"`
		Tag      string `json:"tag"`
		Settings struct {
			Vnext []struct {
				Address string `json:"address"`
				Port    int    `json:"port"`
				Users   []struct {
					ID         string `json:"id"`
					Flow       string `json:"flow"`
					Encryption string `json:"encryption"`
				} `json:"users"`
			} `json:"vnext"`
			Servers []struct {
				Address  string `json:"address"`
				Port     int    `json:"port"`
				Password string `json:"password"`
				Auth     string `json:"auth"`
				Sni      string `json:"sni"`
			} `json:"servers"`
		} `json:"settings"`
		StreamSettings struct {
			Network         string `json:"network"`
			Security        string `json:"security"`
			RealitySettings struct {
				ServerName  string `json:"serverName"`
				Fingerprint string `json:"fingerprint"`
				PublicKey   string `json:"publicKey"`
				ShortId     string `json:"shortId"`
			} `json:"realitySettings"`
			TlsSettings struct {
				ServerName  string `json:"serverName"`
				Fingerprint string `json:"fingerprint"`
			} `json:"tlsSettings"`
			WsSettings struct {
				Path string `json:"path"`
				Host string `json:"host"`
			} `json:"wsSettings"`
			GrpcSettings struct {
				ServiceName string `json:"serviceName"`
			} `json:"grpcSettings"`
			XhttpSettings struct {
				Path string `json:"path"`
				Host string `json:"host"`
				Mode string `json:"mode"`
			} `json:"xhttpSettings"`
		} `json:"streamSettings"`
	} `json:"outbounds"`
}

type SingBoxConfig struct {
	Outbounds []struct {
		Type       string `json:"type"`
		Tag        string `json:"tag"`
		Server     string `json:"server"`
		ServerPort int    `json:"server_port"`
		UUID       string `json:"uuid"`
		Password   string `json:"password"`
		Flow       string `json:"flow"`
		Transport  struct {
			Type        string            `json:"type"`
			Path        string            `json:"path"`
			Headers     map[string]string `json:"headers"`
			ServiceName string            `json:"service_name"`
		} `json:"transport"`
		Tls struct {
			Enabled    bool   `json:"enabled"`
			ServerName string `json:"server_name"`
			Utls       struct {
				Fingerprint string `json:"fingerprint"`
			} `json:"utls"`
			Alpn []string `json:"alpn"`
		} `json:"tls"`
		Reality struct {
			Enabled    bool   `json:"enabled"`
			ServerName string `json:"server_name"`
			PublicKey  string `json:"public_key"`
			ShortID    string `json:"short_id"`
		} `json:"reality"`
	} `json:"outbounds"`
}

type ClashConfig struct {
	Proxies []struct {
		Name              string `yaml:"name"`
		Type              string `yaml:"type"`
		Server            string `yaml:"server"`
		Port              int    `yaml:"port"`
		UUID              string `yaml:"uuid"`
		Password          string `yaml:"password"`
		Network           string `yaml:"network"`
		Tls               bool   `yaml:"tls"`
		Servername        string `yaml:"servername"`
		Sni               string `yaml:"sni"`
		ClientFingerprint string `yaml:"client-fingerprint"`
		RealityOpts       struct {
			PublicKey string `yaml:"public-key"`
			ShortId   string `yaml:"short-id"`
		} `yaml:"reality-opts"`
		WsOpts struct {
			Path    string `yaml:"path"`
			Headers struct {
				Host string `yaml:"Host"`
			} `yaml:"headers"`
		} `yaml:"ws-opts"`
		GrpcOpts struct {
			ServiceName string `yaml:"grpc-service-name"`
		} `yaml:"grpc-opts"`
	} `yaml:"proxies"`
}

// ===============================
// Кодирование и небольшие утилиты
// ===============================

func encodeName(name string) string {
	return strings.ReplaceAll(url.QueryEscape(name), "+", "%20")
}

func buildQuery(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}
	keys := make([]string, 0, len(params))
	for k, v := range params {
		if v != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var b strings.Builder
	b.Grow(len(keys) * 24)
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('&')
		}
		b.WriteString(url.QueryEscape(k))
		b.WriteByte('=')
		v := url.QueryEscape(params[k])
		v = strings.ReplaceAll(v, "%2F", "/")
		v = strings.ReplaceAll(v, "+", "%20")
		b.WriteString(v)
	}
	return b.String()
}

func makeURI(scheme, credential, host string, port int, params map[string]string, name string) string {
	query := buildQuery(params)
	authority := net.JoinHostPort(host, strconv.Itoa(port)) // корректно обрабатывает IPv6
	var b strings.Builder
	b.Grow(len(scheme) + len(credential) + len(authority) + len(query) + len(name) + 12)
	b.WriteString(scheme)
	b.WriteString("://")
	b.WriteString(credential)
	b.WriteByte('@')
	b.WriteString(authority)
	b.WriteByte('?')
	b.WriteString(query)
	b.WriteByte('#')
	b.WriteString(encodeName(name))
	return b.String()
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "https://", "")
	name = strings.ReplaceAll(name, "http://", "")
	name = strings.ReplaceAll(name, "t.me/", "")
	name = strings.ReplaceAll(name, "www.", "")
	name = invalidFilenameRE.ReplaceAllString(name, "_")
	name = strings.Join(strings.Fields(name), " ") // Убираем двойные пробелы
	name = strings.TrimSpace(name)
	name = strings.Trim(name, ".")
	if name == "" {
		return "unnamed"
	}
	return name
}

// Удаляем комментарии построчно без Scanner и промежуточных []string.
func stripMetaLines(content string) string {
	var b strings.Builder
	b.Grow(len(content))
	for len(content) > 0 {
		end := strings.IndexByte(content, '\n')
		line, rest := content, ""
		if end >= 0 {
			line, rest = content[:end], content[end+1:]
		}
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			b.WriteString(line)
			if end >= 0 {
				b.WriteByte('\n')
			}
		}
		content = rest
	}
	return b.String()
}

func isTargetScheme(line string) bool {
	i := strings.Index(line, "://")
	if i < 1 {
		return false
	}
	switch strings.ToLower(line[:i]) {
	case "vless", "trojan", "hysteria2", "hy2":
		return true
	default:
		return false
	}
}

func appendTextURIs(dst []string, text string) []string {
	for len(text) > 0 {
		end := strings.IndexByte(text, '\n')
		line, rest := text, ""
		if end >= 0 {
			line, rest = text[:end], text[end+1:]
		}
		line = strings.TrimSpace(line)
		if isTargetScheme(line) {
			dst = append(dst, line)
		}
		text = rest
	}
	return dst
}

// ===============================
// Построение URI из JSON/YAML
// ===============================

func getStreamParams(s struct {
	Network         string `json:"network"`
	Security        string `json:"security"`
	RealitySettings struct {
		ServerName  string `json:"serverName"`
		Fingerprint string `json:"fingerprint"`
		PublicKey   string `json:"publicKey"`
		ShortId     string `json:"shortId"`
	} `json:"realitySettings"`
	TlsSettings struct {
		ServerName  string `json:"serverName"`
		Fingerprint string `json:"fingerprint"`
	} `json:"tlsSettings"`
	WsSettings struct {
		Path string `json:"path"`
		Host string `json:"host"`
	} `json:"wsSettings"`
	GrpcSettings struct {
		ServiceName string `json:"serviceName"`
	} `json:"grpcSettings"`
	XhttpSettings struct {
		Path string `json:"path"`
		Host string `json:"host"`
		Mode string `json:"mode"`
	} `json:"xhttpSettings"`
}) map[string]string {
	p := make(map[string]string, 10)
	network := s.Network
	if network == "" {
		network = "tcp"
	}
	p["type"] = network
	if s.Security != "" && s.Security != "none" {
		p["security"] = s.Security
		if s.Security == "reality" {
			if s.RealitySettings.ServerName != "" {
				p["sni"] = s.RealitySettings.ServerName
			}
			if s.RealitySettings.Fingerprint != "" {
				p["fp"] = s.RealitySettings.Fingerprint
			}
			if s.RealitySettings.PublicKey != "" {
				p["pbk"] = s.RealitySettings.PublicKey
			}
			if s.RealitySettings.ShortId != "" {
				p["sid"] = s.RealitySettings.ShortId
			}
		} else if s.Security == "tls" {
			if s.TlsSettings.ServerName != "" {
				p["sni"] = s.TlsSettings.ServerName
			}
			if s.TlsSettings.Fingerprint != "" {
				p["fp"] = s.TlsSettings.Fingerprint
			}
		}
	}
	if network == "ws" {
		if s.WsSettings.Path != "" {
			p["path"] = s.WsSettings.Path
		}
		if s.WsSettings.Host != "" {
			p["host"] = s.WsSettings.Host
		}
	} else if network == "grpc" {
		if s.GrpcSettings.ServiceName != "" {
			p["serviceName"] = s.GrpcSettings.ServiceName
		}
	} else if network == "xhttp" {
		if s.XhttpSettings.Path != "" {
			p["path"] = s.XhttpSettings.Path
		}
		if s.XhttpSettings.Host != "" {
			p["host"] = s.XhttpSettings.Host
		}
		if s.XhttpSettings.Mode != "" {
			p["mode"] = s.XhttpSettings.Mode
		}
	}
	return p
}

func extractFromXray(cfg XrayConfig) []string {
	var uris []string
	for _, ob := range cfg.Outbounds {
		proto := strings.ToLower(ob.Protocol)
		if proto != "vless" && proto != "trojan" && proto != "hysteria2" {
			continue
		}
		params := getStreamParams(ob.StreamSettings)
		switch proto {
		case "vless":
			for _, vnext := range ob.Settings.Vnext {
				for _, user := range vnext.Users {
					p := make(map[string]string, len(params)+2)
					for k, v := range params {
						p[k] = v
					}
					p["encryption"] = user.Encryption
					if p["encryption"] == "" {
						p["encryption"] = "none"
					}
					if user.Flow != "" {
						p["flow"] = user.Flow
					}
					uris = append(uris, makeURI("vless", user.ID, vnext.Address, vnext.Port, p, ob.Tag))
				}
			}
		case "trojan":
			for _, srv := range ob.Settings.Servers {
				uris = append(uris, makeURI("trojan", srv.Password, srv.Address, srv.Port, params, ob.Tag))
			}
		case "hysteria2":
			for _, srv := range ob.Settings.Servers {
				pwd := srv.Password
				if pwd == "" {
					pwd = srv.Auth
				}
				hp := make(map[string]string, 1)
				if srv.Sni != "" {
					hp["sni"] = srv.Sni
				}
				uris = append(uris, makeURI("hysteria2", pwd, srv.Address, srv.Port, hp, ob.Tag))
			}
		}
	}
	return uris
}

func extractFromSingBox(cfg SingBoxConfig) []string {
	var uris []string
	for _, ob := range cfg.Outbounds {
		t := strings.ToLower(ob.Type)
		if t != "vless" && t != "trojan" && t != "hysteria2" {
			continue
		}
		p := make(map[string]string, 12)
		transport := ob.Transport.Type
		if transport == "" {
			transport = "tcp"
		}
		p["type"] = transport
		if ob.Tls.Enabled {
			p["security"] = "tls"
			if ob.Tls.ServerName != "" {
				p["sni"] = ob.Tls.ServerName
			}
			if ob.Tls.Utls.Fingerprint != "" {
				p["fp"] = ob.Tls.Utls.Fingerprint
			}
			if len(ob.Tls.Alpn) > 0 {
				p["alpn"] = strings.Join(ob.Tls.Alpn, ",")
			}
		}
		if ob.Reality.Enabled {
			p["security"] = "reality"
			p["sni"] = ob.Reality.ServerName
			p["pbk"] = ob.Reality.PublicKey
			p["sid"] = ob.Reality.ShortID
		}
		if transport == "ws" {
			if ob.Transport.Path != "" {
				p["path"] = ob.Transport.Path
			}
			if host, ok := ob.Transport.Headers["Host"]; ok {
				p["host"] = host
			}
		} else if transport == "grpc" && ob.Transport.ServiceName != "" {
			p["serviceName"] = ob.Transport.ServiceName
		}
		switch t {
		case "vless":
			p["encryption"] = "none"
			if ob.Flow != "" {
				p["flow"] = ob.Flow
			}
			uris = append(uris, makeURI("vless", ob.UUID, ob.Server, ob.ServerPort, p, ob.Tag))
		case "trojan":
			uris = append(uris, makeURI("trojan", ob.Password, ob.Server, ob.ServerPort, p, ob.Tag))
		case "hysteria2":
			hp := make(map[string]string, 1)
			if ob.Tls.ServerName != "" {
				hp["sni"] = ob.Tls.ServerName
			}
			uris = append(uris, makeURI("hysteria2", ob.Password, ob.Server, ob.ServerPort, hp, ob.Tag))
		}
	}
	return uris
}

func extractFromClash(cfg ClashConfig) []string {
	var uris []string
	for _, proxy := range cfg.Proxies {
		t := strings.ToLower(proxy.Type)
		if t != "vless" && t != "trojan" && t != "hysteria2" {
			continue
		}
		p := make(map[string]string, 10)
		network := proxy.Network
		if network == "" {
			network = "tcp"
		}
		p["type"] = network
		if proxy.Tls {
			p["security"] = "tls"
			if proxy.Servername != "" {
				p["sni"] = proxy.Servername
			} else if proxy.Sni != "" {
				p["sni"] = proxy.Sni
			}
			if proxy.ClientFingerprint != "" {
				p["fp"] = proxy.ClientFingerprint
			}
		}
		if proxy.RealityOpts.PublicKey != "" {
			p["security"] = "reality"
			p["pbk"] = proxy.RealityOpts.PublicKey
			p["sid"] = proxy.RealityOpts.ShortId
		}
		if network == "ws" {
			if proxy.WsOpts.Path != "" {
				p["path"] = proxy.WsOpts.Path
			}
			if proxy.WsOpts.Headers.Host != "" {
				p["host"] = proxy.WsOpts.Headers.Host
			}
		} else if network == "grpc" && proxy.GrpcOpts.ServiceName != "" {
			p["serviceName"] = proxy.GrpcOpts.ServiceName
		}
		switch t {
		case "vless":
			p["encryption"] = "none"
			uris = append(uris, makeURI("vless", proxy.UUID, proxy.Server, proxy.Port, p, proxy.Name))
		case "trojan":
			uris = append(uris, makeURI("trojan", proxy.Password, proxy.Server, proxy.Port, p, proxy.Name))
		case "hysteria2":
			hp := make(map[string]string, 1)
			if proxy.Sni != "" {
				hp["sni"] = proxy.Sni
			}
			uris = append(uris, makeURI("hysteria2", proxy.Password, proxy.Server, proxy.Port, hp, proxy.Name))
		}
	}
	return uris
}

func extractFromB64OrText(content string) []string {
	clean := stripMetaLines(content)
	trimmed := strings.TrimSpace(clean)
	if trimmed != "" {
		decoded, err := base64.StdEncoding.DecodeString(trimmed)
		if err != nil {
			decoded, err = base64.RawStdEncoding.DecodeString(trimmed)
		}
		if err != nil {
			decoded, err = base64.URLEncoding.DecodeString(trimmed)
		}
		if err != nil {
			decoded, err = base64.RawURLEncoding.DecodeString(trimmed)
		}
		if err == nil {
			if uris := appendTextURIs(nil, string(decoded)); len(uris) > 0 {
				return uris
			}
		}
	}
	return appendTextURIs(nil, clean)
}

// JSON сначала классифицируется по первому outbound, после чего декодируется
// ровно в одну целевую структуру. Это дешевле нескольких полных Unmarshal подряд.
func extractURIs(content string) ([]string, interface{}) {
	clean := strings.TrimSpace(stripMetaLines(strings.TrimPrefix(content, "\uFEFF")))
	if clean == "" {
		return nil, nil
	}

	if clean[0] == '[' {
		var xrayConfigs []XrayConfig
		if err := json.Unmarshal([]byte(clean), &xrayConfigs); err == nil {
			var uris []string
			for _, cfg := range xrayConfigs {
				uris = append(uris, extractFromXray(cfg)...)
			}
			if len(uris) > 0 {
				return uris, xrayConfigs
			}
		}
		var singConfigs []SingBoxConfig
		if err := json.Unmarshal([]byte(clean), &singConfigs); err == nil {
			var uris []string
			for _, cfg := range singConfigs {
				uris = append(uris, extractFromSingBox(cfg)...)
			}
			if len(uris) > 0 {
				return uris, singConfigs
			}
		}
	}

	if clean[0] == '{' {
		var envelope struct {
			Outbounds []json.RawMessage `json:"outbounds"`
		}
		if err := json.Unmarshal([]byte(clean), &envelope); err == nil && len(envelope.Outbounds) > 0 {
			var jsonKind string
			for _, rawOutbound := range envelope.Outbounds {
				var kind struct {
					Protocol string `json:"protocol"`
					Type     string `json:"type"`
				}
				if json.Unmarshal(rawOutbound, &kind) != nil {
					continue
				}
				if kind.Protocol != "" {
					jsonKind = "xray"
					break
				}
				if kind.Type != "" {
					jsonKind = "singbox"
					break
				}
			}
			switch jsonKind {
			case "xray":
				var cfg XrayConfig
				if json.Unmarshal([]byte(clean), &cfg) == nil {
					return extractFromXray(cfg), cfg
				}
			case "singbox":
				var cfg SingBoxConfig
				if json.Unmarshal([]byte(clean), &cfg) == nil {
					return extractFromSingBox(cfg), cfg
				}
			}
		}
	}

	var clash ClashConfig
	if err := yaml.Unmarshal([]byte(clean), &clash); err == nil && len(clash.Proxies) > 0 {
		return extractFromClash(clash), clash
	}
	return extractFromB64OrText(content), nil
}

// ===============================
// Разбор URI, DNS и TCP-проверка
// ===============================

type endpoint struct {
	URI    string
	Scheme string
	Host   string
	Port   string
}

type configEntry struct {
	URI      string
	Endpoint endpoint
}

func parseEndpoint(uri string) (endpoint, bool) {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme == "" || u.Hostname() == "" || u.Port() == "" {
		return endpoint{}, false
	}
	scheme := strings.ToLower(u.Scheme)
	switch scheme {
	case "vless", "trojan", "hysteria2", "hy2":
	default:
		return endpoint{}, false
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil || port < 1 || port > 65535 {
		return endpoint{}, false
	}
	return endpoint{URI: uri, Scheme: scheme, Host: u.Hostname(), Port: strconv.Itoa(port)}, true
}

func getFingerprint(uri string) string {
	if i := strings.IndexByte(uri, '#'); i >= 0 {
		uri = uri[:i]
	}
	q := strings.IndexByte(uri, '?')
	if q < 0 {
		return uri
	}
	parts := strings.Split(uri[q+1:], "&")
	sort.Strings(parts)
	return uri[:q+1] + strings.Join(parts, "&")
}

func filterAndDedupe(uris []string) []configEntry {
	if len(uris) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(uris))
	out := make([]configEntry, 0, len(uris))
	for _, uri := range uris {
		if !isTargetScheme(uri) {
			continue
		}
		fp := getFingerprint(uri)
		if _, exists := seen[fp]; exists {
			continue
		}
		seen[fp] = struct{}{}
		ep, ok := parseEndpoint(uri)
		if !ok {
			continue
		}
		out = append(out, configEntry{URI: uri, Endpoint: ep})
	}
	return out
}

type dnsRecord struct {
	once sync.Once
	ips  []net.IP
	err  error
}

func resolveHost(host string) *dnsRecord {
	value, _ := dnsCache.LoadOrStore(host, &dnsRecord{})
	record := value.(*dnsRecord)
	record.once.Do(func() {
		if ip := net.ParseIP(host); ip != nil {
			record.ips = []net.IP{ip}
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), dnsTimeout)
		defer cancel()
		record.ips, record.err = net.DefaultResolver.LookupIP(ctx, "ip", host)
	})
	return record
}

type probeResult struct {
	Keep       bool // оставлять ли конфиг в итоговом файле
	CheckedTCP bool
	AliveTCP   bool
	Latency    time.Duration
	IP         string
}

func probe(ep endpoint) probeResult {
	dns := resolveHost(ep.Host)
	result := probeResult{}
	if len(dns.ips) > 0 {
		result.IP = dns.ips[0].String()
	}

	// Hysteria2 использует UDP/QUIC. TCP connect к его порту не является
	// проверкой работоспособности, поэтому сохраняем запись без TCP-вердикта.
	if ep.Scheme == "hysteria2" || ep.Scheme == "hy2" {
		result.Keep = true
		return result
	}
	result.CheckedTCP = true
	if dns.err != nil || len(dns.ips) == 0 {
		return result
	}

	ctx, cancel := context.WithTimeout(context.Background(), tcpTimeout)
	defer cancel()
	start := time.Now()
	dialer := net.Dialer{KeepAlive: -1}
	for _, ip := range dns.ips {
		address := net.JoinHostPort(ip.String(), ep.Port)
		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err == nil {
			result.Latency = time.Since(start)
			if remote, ok := conn.RemoteAddr().(*net.TCPAddr); ok && remote.IP != nil {
				result.IP = remote.IP.String()
			}
			_ = conn.Close()
			result.Keep = true
			result.AliveTCP = true
			return result
		}
		if ctx.Err() != nil {
			break
		}
	}
	result.Latency = time.Since(start)
	return result
}

// ===============================
// GeoIP и имя выходного файла
// ===============================

func processGeoIP(uri string, ep endpoint, ip string) string {
	baseURI := uri
	if i := strings.IndexByte(baseURI, '#'); i >= 0 {
		baseURI = baseURI[:i]
	}

	protocol := strings.ToUpper(ep.Scheme)
	hostPort := net.JoinHostPort(ep.Host, ep.Port)
	name := "Unknown | " + protocol + " | " + hostPort
	if geoDb != nil && ip != "" {
		if record, err := geoDb.Country(net.ParseIP(ip)); err == nil {
			iso := record.Country.IsoCode
			if len(iso) == 2 {
				country := record.Country.Names["en"]
				if country == "" {
					country = iso
				}
				flag := string(rune(iso[0])+127397) + string(rune(iso[1])+127397)
				name = flag + " " + country + " | " + protocol + " | " + hostPort
			}
		}
	}
	return baseURI + "#" + encodeName(name)
}

func determineName(resp *http.Response, body, sourceURL string, parsedRoot interface{}) string {
	if title := resp.Header.Get("profile-title"); title != "" {
		return sanitizeFilename(title)
	}
	if match := profileTitleRE.FindStringSubmatch(body); len(match) > 1 {
		return sanitizeFilename(match[1])
	}

	validTag := func(tag, protocol string) bool {
		switch strings.ToLower(protocol) {
		case "freedom", "blackhole", "dns":
			return false
		default:
			return tag != ""
		}
	}
	findXray := func(cfg XrayConfig) string {
		if cfg.Remarks != "" {
			return cfg.Remarks
		}
		for _, ob := range cfg.Outbounds {
			if validTag(ob.Tag, ob.Protocol) {
				return ob.Tag
			}
		}
		return ""
	}

	switch cfg := parsedRoot.(type) {
	case []XrayConfig:
		for _, c := range cfg {
			if n := findXray(c); n != "" {
				return sanitizeFilename(n)
			}
		}
	case XrayConfig:
		if n := findXray(cfg); n != "" {
			return sanitizeFilename(n)
		}
	case SingBoxConfig:
		for _, ob := range cfg.Outbounds {
			if ob.Tag != "" {
				return sanitizeFilename(ob.Tag)
			}
		}
	case []SingBoxConfig:
		for _, c := range cfg {
			for _, ob := range c.Outbounds {
				if ob.Tag != "" {
					return sanitizeFilename(ob.Tag)
				}
			}
		}
	case ClashConfig:
		for _, p := range cfg.Proxies {
			if p.Name != "" {
				return sanitizeFilename(p.Name)
			}
		}
	}

	if pageURL := resp.Header.Get("profile-web-page-url"); pageURL != "" {
		if u, err := url.Parse(pageURL); err == nil {
			parts := strings.Split(strings.Trim(u.Path, "/"), "/")
			if len(parts) > 0 && parts[len(parts)-1] != "" {
				return sanitizeFilename(parts[len(parts)-1])
			}
			if u.Host != "" {
				return sanitizeFilename(u.Host)
			}
		}
	}
	u, err := url.Parse(sourceURL)
	if err != nil {
		return "unnamed"
	}
	return sanitizeFilename(u.Host + strings.ReplaceAll(u.Path, "/", "_"))
}

// ===============================
// Параллельная загрузка источников
// ===============================

type sourceResult struct {
	URL     string
	Name    string
	Entries []configEntry
	Err     error
}

func fetchAndParse(sourceURL string) sourceResult {
	result := sourceResult{URL: sourceURL}
	req, err := http.NewRequest(http.MethodGet, sourceURL, nil)
	if err != nil {
		result.Err = err
		return result
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoSubscriptionChecker/2.0)")

	resp, err := httpClient.Do(req)
	if err != nil {
		result.Err = err
		return result
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		result.Err = fmt.Errorf("HTTP %s", resp.Status)
		return result
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSourceSize+1))
	if err != nil {
		result.Err = err
		return result
	}
	if len(body) > maxSourceSize {
		result.Err = fmt.Errorf("source exceeds %d MiB", maxSourceSize>>20)
		return result
	}
	bodyText := string(body)
	uris, parsedRoot := extractURIs(bodyText)
	result.Name = determineName(resp, bodyText, sourceURL, parsedRoot)
	result.Entries = filterAndDedupe(uris)
	return result
}

func fetchAllSources() []sourceResult {
	results := make([]sourceResult, len(sources))
	workers := fetchWorkers
	if workers > len(sources) {
		workers = len(sources)
	}
	if workers == 0 {
		return results
	}

	jobs := make(chan int)
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for i := range jobs {
				results[i] = fetchAndParse(sources[i])
			}
		}()
	}
	for i := range sources {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return results
}

// Все TCP-задачи используют один фиксированный пул. Количество горутин не
// растёт пропорционально числу конфигураций.
func checkAllSources(results []sourceResult) [][]probeResult {
	checks := make([][]probeResult, len(results))
	total := 0
	for i := range results {
		checks[i] = make([]probeResult, len(results[i].Entries))
		total += len(results[i].Entries)
	}
	if total == 0 {
		return checks
	}

	workers := tcpWorkers
	if workers > total {
		workers = total
	}
	type job struct{ source, index int }
	jobs := make(chan job, workers*2)
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for j := range jobs {
				checks[j.source][j.index] = probe(results[j.source].Entries[j.index].Endpoint)
			}
		}()
	}
	for i := range results {
		for j := range results[i].Entries {
			jobs <- job{source: i, index: j}
		}
	}
	close(jobs)
	wg.Wait()
	return checks
}

// Добавлен аргумент index int для порядкового номера файла
func writeSourceResult(source sourceResult, checks []probeResult, index int) {
	fmt.Printf("\n[*] Source: %s\n", source.URL)
	if source.Err != nil {
		fmt.Printf("    [!] Error: %v\n", source.Err)
		return
	}
	
	// Формируем имя с префиксом (001_, 002_ и т.д.)
	prefixedName := fmt.Sprintf("%03d_%s", index+1, source.Name)
	
	fmt.Printf("    [i] Filename: %s.txt\n", prefixedName)
	if len(source.Entries) == 0 {
		fmt.Println("    [-] No valid configs found.")
		return
	}

	var tcpAlive, tcpChecked, udpSkipped int
	latencies := make([]time.Duration, 0, len(source.Entries))
	finalURIs := make([]string, 0, len(source.Entries))
	for i, entry := range source.Entries {
		check := checks[i]
		if check.CheckedTCP {
			tcpChecked++
			if !check.AliveTCP {
				continue
			}
			tcpAlive++
			latencies = append(latencies, check.Latency)
		} else {
			udpSkipped++
		}
		finalURIs = append(finalURIs, processGeoIP(entry.URI, entry.Endpoint, check.IP))
	}
	fmt.Printf("    [~] Unique configs: %d\n", len(source.Entries))
	fmt.Printf("    [~] TCP alive: %d / %d; Hysteria2 skipped (UDP/QUIC): %d\n", tcpAlive, tcpChecked, udpSkipped)
	if len(latencies) > 0 {
		sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
		var sum time.Duration
		for _, latency := range latencies {
			sum += latency
		}
		fmt.Printf("    [~] TCP connect latency: min %s | median %s | avg %s\n",
			latencies[0].Round(time.Microsecond),
			latencies[len(latencies)/2].Round(time.Microsecond),
			(sum / time.Duration(len(latencies))).Round(time.Microsecond))
	}
	if len(finalURIs) == 0 {
		fmt.Println("    [-] No TCP servers responded.")
		return
	}
	sort.Strings(finalURIs)
	if err := os.MkdirAll("subscriptions", 0755); err != nil {
		fmt.Printf("    [!] Cannot create output directory: %v\n", err)
		return
	}
	
	// Используем имя с префиксом для сохранения
	path := filepath.Join("subscriptions", prefixedName+".txt")
	file, err := os.Create(path)
	if err != nil {
		fmt.Printf("    [!] Cannot create %s: %v\n", path, err)
		return
	}
	writer := bufio.NewWriterSize(file, 1<<20)
	for _, uri := range finalURIs {
		if _, err = writer.WriteString(uri + "\n"); err != nil {
			break
		}
	}
	if flushErr := writer.Flush(); err == nil {
		err = flushErr
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		fmt.Printf("    [!] Write error: %v\n", err)
		return
	}
	fmt.Printf("    [OK] Saved %d configs -> %s\n", len(finalURIs), path)
}

func main() {
	var err error
	geoDb, err = geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		fmt.Println("[!] Warning: GeoLite2-Country.mmdb not found. GeoIP disabled.")
	} else {
		defer geoDb.Close()
	}

	started := time.Now()
	fmt.Printf("[*] Fetching %d sources (workers: %d)...\n", len(sources), fetchWorkers)
	results := fetchAllSources()

	total := 0
	for i := range results {
		total += len(results[i].Entries)
	}
	fmt.Printf("[*] Parsed %d unique configs. Starting checks (workers: %d)...\n", total, tcpWorkers)
	checks := checkAllSources(results)

	for i := range results {
		// Передаем индекс i в функцию записи
		writeSourceResult(results[i], checks[i], i)
	}
	fmt.Printf("\n[OK] Finished in %s\n", time.Since(started).Round(time.Millisecond))
}
