package task

import (
	"context"
	"encoding/json"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/log"
)

// ConversationDetailHandler 按 appType 获取对话详情（导出 CSV 末列「对话详情」）。
type ConversationDetailHandler interface {
	GetAppName(ctx context.Context, appId string) string
	GetConversationDetail(ctx context.Context, conversationLog *model.ConversationLog) (string, error)
}

// conversationDetailHandlers 按 appType 注册的详情处理器。
var conversationDetailHandlers = make(map[string]ConversationDetailHandler)

// conversationDetailHooks 各 appType 必须实现的两个钩子；D 为该 appType 的原始详情类型
type conversationDetailHooks[D any] interface {
	GetAppName(ctx context.Context, appId string) (string, error)
	FetchDetails(ctx context.Context, conversationLog *model.ConversationLog) ([]D, error)
	DetailToQA(D) (conversationDetailQA, bool)
}

// conversationDetailAdapter 把泛型钩子桥接成非泛型 ConversationDetailHandler，
type conversationDetailAdapter[D any] struct{ hooks conversationDetailHooks[D] }

// RegisterConDetailHandler 注册某 appType 的详情处理器（在各 handler 文件的 init 中调用）。
func RegisterConDetailHandler[D any](appType string, handler conversationDetailHooks[D]) {
	conversationDetailHandlers[appType] = conversationDetailAdapter[D]{hooks: handler}
}

// GetAppName 拉详情。
func (a conversationDetailAdapter[D]) GetAppName(ctx context.Context, appId string) string {
	name, err := a.hooks.GetAppName(ctx, appId)
	if err != nil {
		log.Errorf("appId %s get app name %v", appId, err)
	}
	return name
}

// GetConversationDetail 拉详情 → 跳过 toQA 标记 false 的条目 → 逐条转 QA → 序列化。
func (a conversationDetailAdapter[D]) GetConversationDetail(ctx context.Context, conversationLog *model.ConversationLog) (string, error) {
	details, err := a.hooks.FetchDetails(ctx, conversationLog)
	if err != nil {
		return "", err
	}
	qaList := make([]conversationDetailQA, 0, len(details))
	for _, d := range details {
		qa, keep := a.hooks.DetailToQA(d)
		if !keep {
			continue
		}
		qaList = append(qaList, qa)
	}
	return marshalQAList(qaList), nil
}

// conversationDetailQA 对话详情单轮 QA，序列化为表格末列「对话详情」的 JSON 元素。
type conversationDetailQA struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// marshalQAList 把 QA 列表序列化成 CSV 末列「对话详情」的 JSON 字符串。
func marshalQAList(qaList []conversationDetailQA) string {
	b, _ := json.Marshal(qaList)
	return string(b)
}
