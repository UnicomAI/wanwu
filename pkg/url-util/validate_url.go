package url_util

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// allowedSchemes 是允许的 URL scheme 白名单。
var allowedSchemes = map[string]bool{
	"http":  true,
	"https": true,
}

// ValidateURLSyntax 校验 URL 的格式与字面量 IP 安全性，不做 DNS 解析。
// 校验项：
//  1. 非空
//  2. 仅允许 http/https scheme
//  3. host 非空
//  4. host 为 IP 字面量时，拒绝回环/链路本地/未指定/组播地址
//
// 域名只做语法校验。解析后的 IP 校验应放在真正发起请求的 DialContext 阶段：
// 一方面校验时与请求时的 DNS 结果可能不同（rebinding），
// 另一方面启动期/内部可信数据不应依赖 DNS 可用性。
// 需要带 DNS 解析的完整校验时用 ValidateURL。
func ValidateURLSyntax(rawURL string) error {
	_, _, err := parseAndCheckURL(rawURL)
	return err
}

// ValidateURL 在 ValidateURLSyntax 的基础上，对域名做一次 DNS 解析并校验所有解析结果，
// 用于拦截解析到回环/链路本地地址的域名，防止 SSRF。
//
// 不拒绝私有地址（10/172.16/192.168），因为企业私有化部署的用户
// 合法地需要访问内网 API、MinIO 等内部服务。
// 链路本地地址（169.254.0.0/16）被拒绝，因为包含云元数据端点 169.254.169.254。
//
// 返回 nil 表示安全，返回 error 描述拒绝原因。
// 注意：
//   - 本函数做的是一次性校验，不防 DNS rebinding（解析后 IP 变更），
//     如需防 rebinding，应在 DialContext 阶段复核 IP；
//   - 启动期或内部可信数据不要调用本函数，DNS 不可用会导致校验失败，
//     这类场景用 ValidateURLSyntax。
func ValidateURL(rawURL string) error {
	host, isIPLiteral, err := parseAndCheckURL(rawURL)
	if err != nil {
		return err
	}
	if isIPLiteral {
		return nil
	}

	// 否则做 DNS 解析，校验所有解析结果
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("dns lookup %q failed: %w", host, err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("dns lookup %q returned no addresses", host)
	}
	for _, ip := range ips {
		if err := checkIPPolicy(ip, false); err != nil {
			return fmt.Errorf("host %q resolves to blocked address: %w", host, err)
		}
	}

	return nil
}

// parseAndCheckURL 解析并校验 URL 的格式与字面量 IP，返回 host 以及 host 是否为 IP 字面量。
// 返回的 isIPLiteral 用于让调用方决定是否跳过 DNS 解析。
func parseAndCheckURL(rawURL string) (host string, isIPLiteral bool, err error) {
	if strings.TrimSpace(rawURL) == "" {
		return "", false, errors.New("url is empty")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", false, fmt.Errorf("url parse error: %w", err)
	}

	scheme := strings.ToLower(u.Scheme)
	if !allowedSchemes[scheme] {
		return "", false, fmt.Errorf("url scheme %q not allowed, only http/https", u.Scheme)
	}

	host = u.Hostname()
	if host == "" {
		return "", false, errors.New("url host is empty")
	}

	// 如果 host 是 IP 字面量，直接校验
	if ip := net.ParseIP(host); ip != nil {
		if err := checkIPPolicy(ip, false); err != nil {
			return "", false, err
		}
		return host, true, nil
	}

	return host, false, nil
}

// ValidatePublicURL 校验 URL 指向公网地址，用于用户提交的自定义工具/自定义 MCP。
//
// 与 ValidateURL 的区别：除回环/链路本地/未指定/组播外，还拒绝私有地址、
// CGNAT、保留网段以及常见内部主机名（localhost、bff-service、
// *.internal、*.svc、*.cluster.local 等）。
//
// 内置工具、平台内部 schema 不应调用本函数；它们使用 ValidateURL 或
// ValidateURLSyntax，以便继续访问 bff-service 等内部依赖。
//
// 注意：本函数仍是一次性校验，不能单独防 DNS rebinding；用户可控请求
// 还必须在 DialContext 阶段调用 SafeDialContext 复核。
func ValidatePublicURL(rawURL string) error {
	host, isIPLiteral, err := parseAndCheckURL(rawURL)
	if err != nil {
		return err
	}
	host = normalizeHost(host)
	if isIPLiteral {
		return checkIPPolicy(net.ParseIP(host), true)
	}
	if isForbiddenHost(host) {
		return fmt.Errorf("host %q is an internal host", host)
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("dns lookup %q failed: %w", host, err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("dns lookup %q returned no addresses", host)
	}
	for _, ip := range ips {
		if err := checkIPPolicy(ip, true); err != nil {
			return fmt.Errorf("host %q resolves to blocked address: %w", host, err)
		}
	}
	return nil
}

// normalizeHost 统一小写并去掉 DNS 根域末尾的点，避免 bff-service. 绕过。
func normalizeHost(host string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
}

// isForbiddenHost 拒绝明显的内部主机名。私有 IP 的 DNS 解析结果会由
// checkIPPolicy(..., true) 兜底，这里主要用于给出明确错误并覆盖解析到
// 公网 IP 的内部名称。
func isForbiddenHost(host string) bool {
	if host == "" {
		return true
	}
	if _, ok := forbiddenHosts[host]; ok {
		return true
	}
	for _, suffix := range forbiddenHostSuffixes {
		if strings.HasSuffix(host, suffix) {
			return true
		}
	}
	return false
}

var forbiddenHosts = map[string]struct{}{
	"localhost": {},
	// docker compose / k8s 服务名
	"bff-service": {}, "mcp-service": {}, "iam-service": {}, "model-service": {},
	"knowledge-service": {}, "rag-service": {}, "assistant-service": {}, "agent-service": {},
	"app-service": {}, "channel-service": {}, "operate-service": {},
	"mysql": {}, "redis": {}, "minio": {}, "elastic": {}, "kafka": {},
	// 常见的宿主机/元数据别名
	"host.docker.internal": {}, "gateway.docker.internal": {}, "host.containers.internal": {},
	"metadata": {}, "metadata.google.internal": {},
}

var forbiddenHostSuffixes = []string{
	".localhost", ".local", ".internal", ".svc", ".cluster.local", ".localdomain",
}

var forbiddenCIDRs = mustParseCIDRs(
	"100.64.0.0/10",   // CGNAT
	"198.18.0.0/15",   // benchmarking
	"192.0.0.0/24",    // IETF protocol assignments
	"192.0.2.0/24",    // TEST-NET-1
	"198.51.100.0/24", // TEST-NET-2
	"203.0.113.0/24",  // TEST-NET-3
	"240.0.0.0/4",     // reserved
	"2001:db8::/32",   // documentation
	"64:ff9b::/96",    // NAT64
	"2002::/16",       // 6to4
)

func mustParseCIDRs(cidrs ...string) []*net.IPNet {
	ret := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Sprintf("invalid forbidden CIDR %q: %v", cidr, err))
		}
		ret = append(ret, ipNet)
	}
	return ret
}

// checkIPPolicy 是两种 IP 策略的共用实现。rejectPrivate=false 时允许私网；
// true 时额外拒绝私网和 forbiddenCIDRs 中的保留网段。
func checkIPPolicy(ip net.IP, rejectPrivate bool) error {
	if ip == nil {
		return fmt.Errorf("invalid ip address")
	}
	switch {
	case ip.IsUnspecified():
		return fmt.Errorf("address %s is unspecified (0.0.0.0/::)", ip)
	case ip.IsLoopback():
		return fmt.Errorf("address %s is loopback", ip)
	case ip.IsLinkLocalUnicast():
		return fmt.Errorf("address %s is link-local unicast", ip)
	case ip.IsLinkLocalMulticast():
		return fmt.Errorf("address %s is link-local multicast", ip)
	case ip.IsMulticast():
		return fmt.Errorf("address %s is multicast", ip)
	}
	if !rejectPrivate {
		return nil
	}
	if ip.IsPrivate() {
		return fmt.Errorf("address %s is private", ip)
	}
	for _, cidr := range forbiddenCIDRs {
		if cidr.Contains(ip) {
			return fmt.Errorf("address %s is not a public address", ip)
		}
	}
	return nil
}
