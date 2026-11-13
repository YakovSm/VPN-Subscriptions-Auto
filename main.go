package main

import (
	"bufio"
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
	"strings"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang"
	"gopkg.in/yaml.v3"
)

// ==========================================
// НАСТРОЙКИ (ВСТАВЬТЕ СЮДА ВАШИ ССЫЛКИ)
// ==========================================
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

var targetSchemes = map[string]bool{"vless": true, "trojan": true, "hysteria2": true, "hy2": true}
var geoDb *geoip2.Reader
var ipCache sync.Map

// ==========================================
// СТРУКТУРЫ ДЛЯ ПАРСИНГА
// ==========================================

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
			Network  string `json:"network"`
			Security string `json:"security"`
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
			Enabled    bool     `json:"enabled"`
			ServerName string   `json:"server_name"`
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

// ==========================================
// ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ
// ==========================================

func encodeName(name string) string {
	s := url.QueryEscape(name)
	s = strings.ReplaceAll(s, "+", "%20")
	return s
}

func buildQuery(params map[string]string) string {
	var keys []string
	for k, v := range params {
		if v != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var sb strings.Builder
	for i, k := range keys {
		if i > 0 { sb.WriteString("&") }
		sb.WriteString(url.QueryEscape(k))
		sb.WriteString("=")
		val := url.QueryEscape(params[k])
		val = strings.ReplaceAll(val, "%2F", "/")
		val = strings.ReplaceAll(val, "+", "%20")
		sb.WriteString(val)
	}
	return sb.String()
}

func sanitizeFilename(name string) string {
	re := regexp.MustCompile(`[<>:"/\\|?*]`)
	name = re.ReplaceAllString(name, "_")
	name = strings.TrimSpace(name)
	name = strings.Trim(name, ".")
	if name == "" { return "unnamed" }
	return name
}

func stripMetaLines(content string) string {
	var sb strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func isTargetScheme(line string) bool {
	return strings.HasPrefix(line, "vless://") || strings.HasPrefix(line, "trojan://") ||
		strings.HasPrefix(line, "hysteria2://") || strings.HasPrefix(line, "hy2://")
}

// ==========================================
// ПАРСИНГ ФОРМАТОВ
// ==========================================

func getStreamParams(s struct {
	Network  string `json:"network"`
	Security string `json:"security"`
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
	p := make(map[string]string)
	network := s.Network
	if network == "" { network = "tcp" }
	p["type"] = network
	if s.Security != "" && s.Security != "none" {
		p["security"] = s.Security
		if s.Security == "reality" {
			if s.RealitySettings.ServerName != "" { p["sni"] = s.RealitySettings.ServerName }
			if s.RealitySettings.Fingerprint != "" { p["fp"] = s.RealitySettings.Fingerprint }
			if s.RealitySettings.PublicKey != "" { p["pbk"] = s.RealitySettings.PublicKey }
			if s.RealitySettings.ShortId != "" { p["sid"] = s.RealitySettings.ShortId }
		} else if s.Security == "tls" {
			if s.TlsSettings.ServerName != "" { p["sni"] = s.TlsSettings.ServerName }
			if s.TlsSettings.Fingerprint != "" { p["fp"] = s.TlsSettings.Fingerprint }
		}
	}
	if network == "ws" {
		if s.WsSettings.Path != "" { p["path"] = s.WsSettings.Path }
		if s.WsSettings.Host != "" { p["host"] = s.WsSettings.Host }
	} else if network == "grpc" {
		if s.GrpcSettings.ServiceName != "" { p["serviceName"] = s.GrpcSettings.ServiceName }
	} else if network == "xhttp" {
		if s.XhttpSettings.Path != "" { p["path"] = s.XhttpSettings.Path }
		if s.XhttpSettings.Host != "" { p["host"] = s.XhttpSettings.Host }
		if s.XhttpSettings.Mode != "" { p["mode"] = s.XhttpSettings.Mode }
	}
	return p
}

func extractFromXray(cfg XrayConfig) []string {
	var uris []string
	for _, ob := range cfg.Outbounds {
		proto := strings.ToLower(ob.Protocol)
		if proto != "vless" && proto != "trojan" && proto != "hysteria2" { continue }
		stream := ob.StreamSettings
		params := getStreamParams(stream)

		if proto == "vless" {
			for _, vnext := range ob.Settings.Vnext {
				for _, user := range vnext.Users {
					p := make(map[string]string)
					for k, v := range params { p[k] = v }
					p["encryption"] = user.Encryption
					if p["encryption"] == "" { p["encryption"] = "none" }
					if user.Flow != "" { p["flow"] = user.Flow }
					uris = append(uris, fmt.Sprintf("vless://%s@%s:%d?%s#%s", user.ID, vnext.Address, vnext.Port, buildQuery(p), encodeName(ob.Tag)))
				}
			}
		} else if proto == "trojan" {
			for _, srv := range ob.Settings.Servers {
				uris = append(uris, fmt.Sprintf("trojan://%s@%s:%d?%s#%s", srv.Password, srv.Address, srv.Port, buildQuery(params), encodeName(ob.Tag)))
			}
		} else if proto == "hysteria2" {
			for _, srv := range ob.Settings.Servers {
				pwd := srv.Password
				if pwd == "" { pwd = srv.Auth }
				hp := make(map[string]string)
				if srv.Sni != "" { hp["sni"] = srv.Sni }
				uris = append(uris, fmt.Sprintf("hysteria2://%s@%s:%d?%s#%s", pwd, srv.Address, srv.Port, buildQuery(hp), encodeName(ob.Tag)))
			}
		}
	}
	return uris
}

func extractFromSingBox(cfg SingBoxConfig) []string {
	var uris []string
	for _, ob := range cfg.Outbounds {
		t := strings.ToLower(ob.Type)
		if t != "vless" && t != "trojan" && t != "hysteria2" { continue }
		p := make(map[string]string)
		transport := ob.Transport.Type
		if transport == "" { transport = "tcp" }
		p["type"] = transport
		if ob.Tls.Enabled {
			p["security"] = "tls"
			if ob.Tls.ServerName != "" { p["sni"] = ob.Tls.ServerName }
			if ob.Tls.Utls.Fingerprint != "" { p["fp"] = ob.Tls.Utls.Fingerprint }
			if len(ob.Tls.Alpn) > 0 { p["alpn"] = strings.Join(ob.Tls.Alpn, ",") }
		}
		if ob.Reality.Enabled {
			p["security"] = "reality"
			p["sni"] = ob.Reality.ServerName
			p["pbk"] = ob.Reality.PublicKey
			p["sid"] = ob.Reality.ShortID
		}
		if transport == "ws" {
			if ob.Transport.Path != "" { p["path"] = ob.Transport.Path }
			if host, ok := ob.Transport.Headers["Host"]; ok { p["host"] = host }
		} else if transport == "grpc" {
			if ob.Transport.ServiceName != "" { p["serviceName"] = ob.Transport.ServiceName }
		}

		if t == "vless" {
			p["encryption"] = "none"
			if ob.Flow != "" { p["flow"] = ob.Flow }
			uris = append(uris, fmt.Sprintf("vless://%s@%s:%d?%s#%s", ob.UUID, ob.Server, ob.ServerPort, buildQuery(p), encodeName(ob.Tag)))
		} else if t == "trojan" {
			uris = append(uris, fmt.Sprintf("trojan://%s@%s:%d?%s#%s", ob.Password, ob.Server, ob.ServerPort, buildQuery(p), encodeName(ob.Tag)))
		} else if t == "hysteria2" {
			hp := make(map[string]string)
			if ob.Tls.ServerName != "" { hp["sni"] = ob.Tls.ServerName }
			uris = append(uris, fmt.Sprintf("hysteria2://%s@%s:%d?%s#%s", ob.Password, ob.Server, ob.ServerPort, buildQuery(hp), encodeName(ob.Tag)))
		}
	}
	return uris
}

func extractFromClash(cfg ClashConfig) []string {
	var uris []string
	for _, proxy := range cfg.Proxies {
		t := strings.ToLower(proxy.Type)
		if t != "vless" && t != "trojan" && t != "hysteria2" { continue }
		p := make(map[string]string)
		network := proxy.Network
		if network == "" { network = "tcp" }
		p["type"] = network
		if proxy.Tls {
			p["security"] = "tls"
			if proxy.Servername != "" { p["sni"] = proxy.Servername } else if proxy.Sni != "" { p["sni"] = proxy.Sni }
			if proxy.ClientFingerprint != "" { p["fp"] = proxy.ClientFingerprint }
		}
		if proxy.RealityOpts.PublicKey != "" {
			p["security"] = "reality"
			p["pbk"] = proxy.RealityOpts.PublicKey
			p["sid"] = proxy.RealityOpts.ShortId
		}
		if network == "ws" {
			if proxy.WsOpts.Path != "" { p["path"] = proxy.WsOpts.Path }
			if proxy.WsOpts.Headers.Host != "" { p["host"] = proxy.WsOpts.Headers.Host }
		} else if network == "grpc" {
			if proxy.GrpcOpts.ServiceName != "" { p["serviceName"] = proxy.GrpcOpts.ServiceName }
		}

		if t == "vless" {
			p["encryption"] = "none"
			uris = append(uris, fmt.Sprintf("vless://%s@%s:%d?%s#%s", proxy.UUID, proxy.Server, proxy.Port, buildQuery(p), encodeName(proxy.Name)))
		} else if t == "trojan" {
			uris = append(uris, fmt.Sprintf("trojan://%s@%s:%d?%s#%s", proxy.Password, proxy.Server, proxy.Port, buildQuery(p), encodeName(proxy.Name)))
		} else if t == "hysteria2" {
			hp := make(map[string]string)
			if proxy.Sni != "" { hp["sni"] = proxy.Sni }
			uris = append(uris, fmt.Sprintf("hysteria2://%s@%s:%d?%s#%s", proxy.Password, proxy.Server, proxy.Port, buildQuery(hp), encodeName(proxy.Name)))
		}
	}
	return uris
}

func extractFromB64OrText(content string) []string {
	var uris []string
	clean := stripMetaLines(content)
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(clean))
	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(decoded)))
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if isTargetScheme(line) { uris = append(uris, line) }
		}
	}
	if len(uris) == 0 {
		scanner := bufio.NewScanner(strings.NewReader(clean))
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if isTargetScheme(line) { uris = append(uris, line) }
		}
	}
	return uris
}

func extractURIs(content string) ([]string, interface{}) {
	clean := strings.TrimSpace(stripMetaLines(content))
	if strings.HasPrefix(clean, "[") || strings.HasPrefix(clean, "{") {
		var xrayConfigs []XrayConfig
		if err := json.Unmarshal([]byte(clean), &xrayConfigs); err == nil {
			var uris []string
			for _, c := range xrayConfigs { uris = append(uris, extractFromXray(c)...) }
			if len(uris) > 0 { return uris, xrayConfigs }
		}
		var xrayConfig XrayConfig
		if err := json.Unmarshal([]byte(clean), &xrayConfig); err == nil && len(xrayConfig.Outbounds) > 0 {
			return extractFromXray(xrayConfig), xrayConfig
		}
		var singboxConfig SingBoxConfig
		if err := json.Unmarshal([]byte(clean), &singboxConfig); err == nil && len(singboxConfig.Outbounds) > 0 {
			return extractFromSingBox(singboxConfig), singboxConfig
		}
	}
	var clashConfig ClashConfig
	if err := yaml.Unmarshal([]byte(clean), &clashConfig); err == nil && len(clashConfig.Proxies) > 0 {
		return extractFromClash(clashConfig), clashConfig
	}
	return extractFromB64OrText(content), nil
}

// ==========================================
// СЕТЕВЫЕ И ГЕО-ФУНКЦИИ
// ==========================================

func checkTCP(uri string) bool {
	at := strings.Index(uri, "@")
	if at == -1 { return false }
	hostPort := uri[at+1:]
	if slash := strings.IndexAny(hostPort, "/?#"); slash != -1 { hostPort = hostPort[:slash] }
	colon := strings.LastIndex(hostPort, ":")
	if colon == -1 { return false }
	host, port := hostPort[:colon], hostPort[colon+1:]
	host = strings.TrimPrefix(host, "[")
	host = strings.TrimSuffix(host, "]")
	if host == "" || port == "" { return false }

	ip := host
	if net.ParseIP(host) == nil {
		if cached, ok := ipCache.Load(host); ok {
			ip = cached.(string)
			if ip == "" { return false }
		} else {
			ips, err := net.LookupIP(host)
			if err != nil || len(ips) == 0 {
				ipCache.Store(host, "")
				return false
			}
			ip = ips[0].String()
			ipCache.Store(host, ip)
		}
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), 2500*time.Millisecond)
	if err != nil { return false }
	conn.Close()
	return true
}

func processGeoIP(uri string) string {
	at := strings.Index(uri, "@")
	if at == -1 { return uri }
	hostPort := uri[at+1:]
	if slash := strings.IndexAny(hostPort, "/?#"); slash != -1 { hostPort = hostPort[:slash] }
	colon := strings.LastIndex(hostPort, ":")
	if colon == -1 { return uri }
	host, port := hostPort[:colon], hostPort[colon+1:]
	host = strings.TrimPrefix(host, "[")
	host = strings.TrimSuffix(host, "]")

	ip := host
	if net.ParseIP(host) == nil {
		if cached, ok := ipCache.Load(host); ok { ip = cached.(string) }
	}
	
	baseUri := uri
	if idx := strings.Index(uri, "#"); idx != -1 { baseUri = uri[:idx] }

	if geoDb == nil || ip == "" {
		return fmt.Sprintf("%s#%s", baseUri, encodeName("Unknown | "+strings.ToUpper(uri[:strings.Index(uri, "://")])+" | "+host+":"+port))
	}
	record, err := geoDb.Country(net.ParseIP(ip))
	if err != nil || record.Country.IsoCode == "" {
		return fmt.Sprintf("%s#%s", baseUri, encodeName("Unknown | "+strings.ToUpper(uri[:strings.Index(uri, "://")])+" | "+host+":"+port))
	}
	iso := record.Country.IsoCode
	country := record.Country.Names["en"]
	flag := string(rune(iso[0])+127397) + string(rune(iso[1])+127397)
	name := fmt.Sprintf("%s %s | %s | %s:%s", flag, country, strings.ToUpper(uri[:strings.Index(uri, "://")]), host, port)
	return fmt.Sprintf("%s#%s", baseUri, encodeName(name))
}

func getFingerprint(uri string) string {
	if hashIdx := strings.Index(uri, "#"); hashIdx != -1 { uri = uri[:hashIdx] }
	qIdx := strings.Index(uri, "?")
	if qIdx == -1 { return uri }
	mainPart := uri[:qIdx]
	query := uri[qIdx+1:]
	parts := strings.Split(query, "&")
	sort.Strings(parts)
	return mainPart + "?" + strings.Join(parts, "&")
}

// ==========================================
// ГЛАВНАЯ ФУНКЦИЯ
// ==========================================

func determineName(resp *http.Response, body string, urlStr string, parsedRoot interface{}) string {
	if t := resp.Header.Get("profile-title"); t != "" { return sanitizeFilename(t) }
	re := regexp.MustCompile(`(?mi)^#\s*profile-title:\s*(.+)$`)
	if m := re.FindStringSubmatch(body); len(m) > 1 { return sanitizeFilename(m[1]) }
	
	if configs, ok := parsedRoot.([]XrayConfig); ok {
		for _, c := range configs {
			if c.Remarks != "" { return sanitizeFilename(c.Remarks) }
			for _, ob := range c.Outbounds {
				if ob.Tag != "" && ob.Protocol != "freedom" && ob.Protocol != "blackhole" && ob.Protocol != "dns" { return sanitizeFilename(ob.Tag) }
			}
		}
	}
	if cfg, ok := parsedRoot.(XrayConfig); ok {
		for _, ob := range cfg.Outbounds {
			if ob.Tag != "" && ob.Protocol != "freedom" && ob.Protocol != "blackhole" && ob.Protocol != "dns" { return sanitizeFilename(ob.Tag) }
		}
	}
	if a := resp.Header.Get("profile-web-page-url"); a != "" {
		if u, err := url.Parse(a); err == nil {
			parts := strings.Split(strings.Trim(u.Path, "/"), "/")
			if len(parts) > 0 && parts[len(parts)-1] != "" { return sanitizeFilename(parts[len(parts)-1]) }
			return sanitizeFilename(u.Host)
		}
	}
	u, _ := url.Parse(urlStr)
	return sanitizeFilename(u.Host + strings.ReplaceAll(u.Path, "/", "_"))
}

func processSource(urlStr string) {
	fmt.Printf("\n[*] Source: %s\n", urlStr)
	client := &http.Client{Timeout: 30 * time.Second}
	req, _ := http.NewRequest("GET", urlStr, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("    [!] Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)
	uris, parsedRoot := extractURIs(bodyStr)
	name := determineName(resp, bodyStr, urlStr, parsedRoot)
	fmt.Printf("    [i] Filename: %s.txt\n", name)
	
	if len(uris) == 0 {
		fmt.Println("    [-] No configs found.")
		return
	}
	
	var filtered []string
	for _, u := range uris {
		scheme := u
		if idx := strings.Index(u, "://"); idx != -1 { scheme = u[:idx] }
		if targetSchemes[strings.ToLower(scheme)] { filtered = append(filtered, u) }
	}
	if len(filtered) == 0 {
		fmt.Println("    [-] Empty after scheme filter.")
		return
	}
	
	seen := make(map[string]bool)
	var deduped []string
	for _, u := range filtered {
		fp := getFingerprint(u)
		if !seen[fp] { seen[fp] = true; deduped = append(deduped, u) }
	}
	fmt.Printf("    [~] Unique configs: %d\n", len(deduped))
	
	fmt.Printf("    [~] Checking TCP availability (Go engine)...\n")
	alive := make([]string, 0, len(deduped))
	var mu sync.Mutex
	sem := make(chan struct{}, 1000) 
	var wg sync.WaitGroup
	
	for _, u := range deduped {
		wg.Add(1)
		go func(uri string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if checkTCP(uri) {
				mu.Lock()
				alive = append(alive, uri)
				mu.Unlock()
			}
		}(u)
	}
	wg.Wait()
	
	fmt.Printf("    [~] Alive servers: %d / %d\n", len(alive), len(deduped))
	if len(alive) == 0 {
		fmt.Println("    [-] All servers dead.")
		return
	}
	
	var finalURIs []string
	for _, u := range alive { finalURIs = append(finalURIs, processGeoIP(u)) }
	sort.Strings(finalURIs)
	
	os.MkdirAll("subscriptions", 0755)
	path := filepath.Join("subscriptions", name+".txt")
	file, _ := os.Create(path)
	defer file.Close()
	for _, u := range finalURIs { file.WriteString(u + "\n") }
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

	for _, urlStr := range sources {
		processSource(urlStr)
	}
}