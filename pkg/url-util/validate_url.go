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
		if err := checkIP(ip); err != nil {
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
		if err := checkIP(ip); err != nil {
			return "", false, err
		}
		return host, true, nil
	}

	return host, false, nil
}

// checkIP 检查单个 IP 是否属于被禁止的网段。
// 不拒绝私有地址（IsPrivate），因为企业私有化部署用户合法访问内网服务。
func checkIP(ip net.IP) error {
	if ip.IsUnspecified() {
		return fmt.Errorf("address %s is unspecified (0.0.0.0/::)", ip)
	}
	if ip.IsLoopback() {
		return fmt.Errorf("address %s is loopback", ip)
	}
	if ip.IsLinkLocalUnicast() {
		return fmt.Errorf("address %s is link-local unicast", ip)
	}
	if ip.IsLinkLocalMulticast() {
		return fmt.Errorf("address %s is link-local multicast", ip)
	}
	if ip.IsMulticast() {
		return fmt.Errorf("address %s is multicast", ip)
	}
	return nil
}
