package url_util

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

// SafeDialContext 返回一个在拨号前解析并校验目标 IP 的 DialContext。
// 解析结果只使用一次，直接拨号已校验的 IP，避免校验与连接之间的 DNS rebinding。
func SafeDialContext(dialer *net.Dialer) func(ctx context.Context, network, addr string) (net.Conn, error) {
	if dialer == nil {
		dialer = &net.Dialer{}
	}
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil {
			return nil, fmt.Errorf("resolve %q: %w", host, err)
		}
		if len(ips) == 0 {
			return nil, fmt.Errorf("resolve %q: no addresses", host)
		}
		for _, ip := range ips {
			if err := checkIPPolicy(ip.IP, true); err != nil {
				return nil, fmt.Errorf("dial %q blocked: %w", host, err)
			}
		}

		var lastErr error
		for _, ip := range ips {
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
			if err == nil {
				return conn, nil
			}
			lastErr = err
		}
		return nil, lastErr
	}
}

// CheckRedirectPublic 限制重定向次数，并重新校验重定向目标为公网地址。
func CheckRedirectPublic(req *http.Request, via []*http.Request) error {
	if len(via) >= 5 {
		return fmt.Errorf("too many redirects")
	}
	return ValidatePublicURL(req.URL.String())
}
