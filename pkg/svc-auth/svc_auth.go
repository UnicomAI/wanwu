// Package svc_auth 提供服务间身份头的 HMAC 签名与校验。
//
// 网关侧（bff / channel）对已认证的用户身份（uid/org）签发 X-Auth-Sign / X-Auth-Ts；
// 下游服务（workflow）验签通过后才信任 X-User-Id / X-Org-Id，否则剥离该身份头——
// 以关闭"任何能直连下游的容器伪造身份头"的漏洞（内部鉴权 / 正向信任）。
package svc_auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

const (
	// HeaderSign 签名头：hex(HMAC-SHA256)。
	HeaderSign = "X-Auth-Sign"
	// HeaderTS 签名时间戳头：Unix 秒。
	HeaderTS = "X-Auth-Ts"
	// TimestampTTL 时间戳允许的正负偏移（秒），用于防重放。
	TimestampTTL = 300
)

// Sign 计算签名：hex(HMAC-SHA256(secret, "<userID>\n<orgID>\n<ts>"))。
func Sign(secret, userID, orgID string, ts int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(userID))
	mac.Write([]byte{'\n'})
	mac.Write([]byte(orgID))
	mac.Write([]byte{'\n'})
	mac.Write([]byte(strconv.Itoa(int(ts))))
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify 校验签名：参数非空、ts 在 TimestampTTL 窗口内、且签名常量时间相等；任一不满足返回 false
// （调用方据此剥离未通过校验的身份头）。
func Verify(secret, userID, orgID, ts, sig string) bool {
	if secret == "" || sig == "" || ts == "" {
		return false
	}
	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return false
	}
	if d := time.Now().Unix() - tsInt; d > TimestampTTL || d < -TimestampTTL {
		return false
	}
	return hmac.Equal([]byte(Sign(secret, userID, orgID, tsInt)), []byte(sig))
}
