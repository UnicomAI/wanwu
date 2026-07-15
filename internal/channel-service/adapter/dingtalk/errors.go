package dingtalk

import (
	"fmt"
	"strings"

	"github.com/UnicomAI/wanwu/internal/channel-service/adapter/types"
)

// dingTalkRateLimitErrCodes 钉钉发送消息触发频控时返回的 errcode。
// 不同发送接口/限流维度返回的码并不统一，这里收录已知频控码：
//   - 130101：发送消息频率超限（机器人单聊/群消息频控）
//   - 4001003：请求过于频繁（部分接口 QPS 限流）
//
// 钉钉未在文档中给出完整频控码表，故额外用 errmsg 关键字兜底识别，
// 命中即包装为 types.ErrIMRateLimited，与微信 ret=-2 统一为"可退避重试"语义。
var dingTalkRateLimitErrCodes = map[int]bool{
	130101:   true,
	4001003:  true,
}

// rateLimitErrMsgKeywords 钉钉频控 errmsg 兜底关键字（小写匹配）。
// 用 "frequen" 词根覆盖 frequency / frequently，用 "频繁" 覆盖过于频繁。
var rateLimitErrMsgKeywords = []string{
	"frequen",     // frequency / frequently / send too frequently
	"rate limit",
	"too many",    // too many requests
	"过于频繁",      // 请求过于频繁
	"频率",
	"频控",
}

// classifyDingTalkSendErr 按钉钉发送响应的 errcode/errmsg 归类发送错误：
//   - 命中频控（已知 errcode 或 errmsg 关键字）：包装为 types.ErrIMRateLimited（%w 保留原信息），
//     供上层（grpc SendMessage / chat 退避重试）按 errors.Is 判定"可退避重试"；
//   - 其他非零 errcode：普通错误，原样透传。
//
// prefix 区分场景（如 "send message failed"），错误信息形如 "{prefix}: errcode={code}, errmsg={msg}"。
func classifyDingTalkSendErr(errcode int, errmsg, prefix string) error {
	msg := fmt.Sprintf("%s: errcode=%d, errmsg=%s", prefix, errcode, errmsg)
	if dingTalkRateLimitErrCodes[errcode] || containsRateLimitKeyword(errmsg) {
		return fmt.Errorf("%w: %s", types.ErrIMRateLimited, msg)
	}
	return fmt.Errorf("%s", msg)
}

// containsRateLimitKeyword 判断 errmsg 是否包含频控兜底关键字（大小写不敏感）。
func containsRateLimitKeyword(errmsg string) bool {
	lower := strings.ToLower(errmsg)
	for _, kw := range rateLimitErrMsgKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
