package task

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/util"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/log"
)

var agentHandler = &agentDetailHandler{}

func init() {
	RegisterConDetailHandler[*assistant_service.ConversionDetailInfo](constant.AppTypeAgent, agentHandler)
}

// agentDetailHandler 智能体(agent)对话详情查询实现。
type agentDetailHandler struct{}

func (h *agentDetailHandler) GetAppName(ctx context.Context, appId string) (string, error) {
	info, err := assistant.GetAssistantInfo(ctx, &assistant_service.GetAssistantInfoReq{AssistantId: appId, Identity: &assistant_service.Identity{}})
	if err != nil {
		return "", err
	}
	if info.AssistantBrief == nil {
		return "", nil
	}
	return info.AssistantBrief.Name, nil
}

// FetchDetails agent 实现：先取 owner 身份，再以该身份分页拉 assistant-service 详情。
func (h *agentDetailHandler) FetchDetails(ctx context.Context, lg *model.ConversationLog) ([]*assistant_service.ConversionDetailInfo, error) {
	if lg == nil {
		return nil, errors.New("lg is nil")
	}
	resp, err := assistant.InternalGetConversationDetailList(ctx, &assistant_service.GetConversationDetailListReq{
		ConversationId: resolveLegacyConversationId(lg),
		PageSize:       conversationDetailPageSize,
		PageNo:         1,
		Identity: &assistant_service.Identity{
			UserId: lg.UserId,
			OrgId:  lg.OrgId,
		},
		ExcludeDeleted: false,
	})
	if err != nil {
		// 单条会话详情查询失败不中断整批导出，记日志后该条详情留空。
		log.Errorf("fetch conversation detail err, conversationId: %s, err: %v", lg.ConversationId, err)
		return nil, err
	}
	return resp.GetData(), err
}

// DetailToQA agent 钩子：question=prompt，answer=ConversationResponse 中 order 最大的 response（并列最大拼接）。
// 返回 false 表示跳过该条（用于跳过 nil detail）；不同 appType 的「取 answer」逻辑各自实现。
func (h *agentDetailHandler) DetailToQA(d *assistant_service.ConversionDetailInfo) (conversationDetailQA, bool) {
	if d == nil {
		return conversationDetailQA{}, false
	}
	return conversationDetailQA{
		Question: d.Prompt,
		Answer:   pickMaxOrderResponse(d.ConversationResponse),
	}, true
}

// conversationDetailPageSize 单次查询对话详情的页大小。
const conversationDetailPageSize = 1000

// pickMaxOrderResponse 取 responses 中 order 最大的那些 response 拼接结果；order 相同的并列最大全部拼接。
// 空列表返回空串。一次遍历完成：遇到更大的 order 重置累积，遇到等于当前最大 order 则追加。
func pickMaxOrderResponse(responses []*assistant_service.ConversationResponse) string {
	n := len(responses)
	if n == 0 {
		return ""
	}
	maxOrder := responses[n-1].Order
	var stack []string
	for i := n - 1; i >= 0; i-- {
		if responses[i].Order != maxOrder {
			break
		}
		stack = append(stack, responses[i].Response) // 倒序入栈: D, C, B
	}
	var sb strings.Builder
	for i := len(stack) - 1; i >= 0; i-- { // 出栈即正序: B, C, D
		sb.WriteString(stack[i])
	}
	return sb.String()
}

func resolveLegacyConversationId(lg *model.ConversationLog) string {
	if lg.Ext != "" {
		var ext struct {
			ConversationIDMark uint32 `json:"conversation_id_mark"`
			ConversationMark   uint32 `json:"conversation_mark"`
		}
		if json.Unmarshal([]byte(lg.Ext), &ext) == nil && ext.ConversationMark == 1 && ext.ConversationIDMark > 0 {
			return util.Int2Str(ext.ConversationIDMark)
		}
	}
	return lg.ConversationId
}
