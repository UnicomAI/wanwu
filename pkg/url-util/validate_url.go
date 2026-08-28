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

// ValidateURL 校验 URL 安全性，防止 SSRF 攻击。
// 校验项：
//  1. 非空
//  2. 仅允许 http/https scheme
//  3. host 非空
//  4. 解析 host 为 IP 或经 DNS 解析后，拒绝回环/链路本地/未指定地址
//
// 不拒绝私有地址（10/172.16/192.168），因为企业私有化部署的用户
// 合法地需要访问内网 API、MinIO 等内部服务。
// 链路本地地址（169.254.0.0/16）被拒绝，因为包含云元数据端点 169.254.169.254。
//
// 返回 nil 表示安全，返回 error 描述拒绝原因。
// 注意：本函数做的是一次性校验，不防 DNS rebinding（解析后 IP 变更）。
// 如需防 rebinding，应在 DialContext 阶段复核 IP。
func ValidateURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return errors.New("url is empty")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("url parse error: %w", err)
	}

	scheme := strings.ToLower(u.Scheme)
	if !allowedSchemes[scheme] {
		return fmt.Errorf("url scheme %q not allowed, only http/https", u.Scheme)
	}

	host := u.Hostname()
	if host == "" {
		return errors.New("url host is empty")
	}

	// 如果 host 是 IP 字面量，直接校验
	if ip := net.ParseIP(host); ip != nil {
		if err := checkIP(ip); err != nil {
			return err
		}
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
