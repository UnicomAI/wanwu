package service

import (
	"encoding/json"

	operate_service "github.com/UnicomAI/wanwu/api/proto/operate-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/gin-gonic/gin"
)

// 消息类别名：byCategory 的 key 固定为这三个语义单词，前端据此渲染各 Tab 角标
const (
	noticeCategoryKeyAnnouncement   = "announcement"
	noticeCategoryKeyProductService = "productService"
	noticeCategoryKeyTicket         = "ticket"
)

// 跳转动作的 msgType 取值。写入侧落库值集与 operate-service 的 orm.MsgTypeXxx 一致
// （agent/workflow/chatflow/rag/skill/knowledge/model/about/custom_url）；
// bff 侧不跨服务引内部包，按前端约定的归一化形态做映射。
// qaKnowledge 是知识问答类旧 msgType 的读侧兼容映射，不在落库值集内。
const (
	noticeMsgTypeAgent       = "agent"
	noticeMsgTypeWorkflow    = "workflow"
	noticeMsgTypeChatflow    = "chatflow"
	noticeMsgTypeRag         = "rag"
	noticeMsgTypeSkill       = "skill"
	noticeMsgTypeKnowledge   = "knowledge"
	noticeMsgTypeQAKnowledge = "qaKnowledge"
	noticeMsgTypeModel       = "model"
)

// noticeCategoryKeys category 数值 → 语义单词
var noticeCategoryKeys = map[int32]string{
	1: noticeCategoryKeyAnnouncement,
	2: noticeCategoryKeyProductService,
	3: noticeCategoryKeyTicket,
}

// GetNoticeUnreadCount 未读总数 + 分类角标，驱动头像红点与各 Tab 角标。
func GetNoticeUnreadCount(ctx *gin.Context, userId, orgId string) (*response.NoticeUnreadCountResp, error) {
	resp, err := notice.GetUnreadCount(ctx.Request.Context(), &operate_service.GetUnreadCountReq{
		UserId: userId,
		OrgId:  orgId,
	})
	if err != nil {
		return nil, err
	}
	// 某类别为 0 时 key 仍要返回，前端才能把角标归零
	byCategory := map[string]int32{
		noticeCategoryKeyAnnouncement:   0,
		noticeCategoryKeyProductService: 0,
		noticeCategoryKeyTicket:         0,
	}
	for category, n := range resp.ByCategory {
		if key, ok := noticeCategoryKeys[category]; ok {
			byCategory[key] = n
		}
	}
	return &response.NoticeUnreadCountResp{
		Total:      resp.Total,
		ByCategory: byCategory,
	}, nil
}

// ListNotice 消息中心整页列表，已读与未读混排，每条带 isRead。
func ListNotice(ctx *gin.Context, userId, orgId string, req *request.NoticeListReq) (*response.PageResult, error) {
	pageNo, pageSize := normalizePage(req.PageNo, req.PageSize)
	resp, err := notice.ListNotice(ctx.Request.Context(), &operate_service.ListNoticeReq{
		UserId:     userId,
		OrgId:      orgId,
		Category:   int32(req.Category),
		OnlyUnread: req.OnlyUnread,
		Keyword:    req.Keyword,
		PageNo:     int32(pageNo),
		PageSize:   int32(pageSize),
	})
	if err != nil {
		return nil, err
	}
	return &response.PageResult{
		List:     buildNoticeList(resp.List),
		Total:    resp.Total,
		PageNo:   pageNo,
		PageSize: pageSize,
	}, nil
}

// ReadNotice 单条消息标记已读。重复对已读消息调用是幂等的。
func ReadNotice(ctx *gin.Context, userId, orgId string, req *request.NoticeReadReq) error {
	_, err := notice.ReadNotice(ctx.Request.Context(), &operate_service.ReadNoticeReq{
		UserId:    userId,
		OrgId:     orgId,
		MessageId: req.MessageId,
	})
	return err
}

// ReadAllNotice 一键已读：当前账号在当前组织上下文的全部未读，不分类别。
func ReadAllNotice(ctx *gin.Context, userId, orgId string) error {
	_, err := notice.ReadAllNotice(ctx.Request.Context(), &operate_service.ReadAllNoticeReq{
		UserId: userId,
		OrgId:  orgId,
	})
	return err
}

// DeleteNotice 批量删除：只从当前账号在当前组织上下文的列表移除，不影响其他收件人。
func DeleteNotice(ctx *gin.Context, userId, orgId string, req *request.NoticeDeleteReq) (*response.NoticeDeleteResp, error) {
	resp, err := notice.DeleteNotice(ctx.Request.Context(), &operate_service.DeleteNoticeReq{
		UserId:     userId,
		OrgId:      orgId,
		MessageIds: req.Ids,
	})
	if err != nil {
		return nil, err
	}
	return &response.NoticeDeleteResp{AffectedCount: resp.AffectedCount}, nil
}

// --- internal ---

func buildNoticeList(items []*operate_service.NoticeItem) []*response.NoticeItem {
	ret := make([]*response.NoticeItem, 0, len(items))
	for _, item := range items {
		ret = append(ret, &response.NoticeItem{
			MessageId:  item.MessageId,
			Title:      item.Title,
			Category:   item.Category,
			Content:    item.Content,
			Actions:    buildNoticeActions(item.Actions),
			ReceivedAt: item.ReceivedAt,
			IsRead:     item.IsRead,
		})
	}
	return ret
}

// buildNoticeActions 把 grpc 侧的 actionParams JSON 字符串解析成对象下发，
// 前端拿到就能直接用，不必再 JSON.parse。
// 无跳转时返回空数组而非 null，前端不必判空。
func buildNoticeActions(actions []*operate_service.NoticeAction) []response.NoticeAction {
	ret := make([]response.NoticeAction, 0, len(actions))
	for _, a := range actions {
		raw := map[string]interface{}{}
		if a.ActionParams != "" {
			if err := json.Unmarshal([]byte(a.ActionParams), &raw); err != nil {
				// 服务端已做 schema 校验，此处解析失败属脏数据：剔除该 action，
				// 对应占位符由前端降级为纯文本，消息主体正常展示
				continue
			}
		}
		ret = append(ret, response.NoticeAction{
			MsgType:      a.MsgType,
			ActionType:   a.ActionType,
			ActionParams: normalizeActionParams(a.MsgType, raw),
		})
	}
	return ret
}

// normalizeActionParams 把库内字段名各异的 actionParams 归一化成前端约定的形态：
// 各资源类型统一用 id 承载资源 ID，msgType 已表明资源类型，无需再带 appType/modelId 等。
//   - agent/workflow/chatflow/rag/skill：appId → id；skill 额外补 skillType:"shared"
//   - knowledge/qaKnowledge：knowledgeId → id
//   - model：modelId → id
//   - custom_url：原样透传（外部链接，字段不归一化）
func normalizeActionParams(msgType string, raw map[string]interface{}) map[string]interface{} {
	idKey := ""
	switch msgType {
	case noticeMsgTypeAgent, noticeMsgTypeWorkflow, noticeMsgTypeChatflow, noticeMsgTypeRag, noticeMsgTypeSkill:
		idKey = "appId"
	case noticeMsgTypeKnowledge, noticeMsgTypeQAKnowledge:
		idKey = "knowledgeId"
	case noticeMsgTypeModel:
		idKey = "modelId"
	default:
		// custom_url 等不归一化的类型原样透传
		return raw
	}
	id, _ := raw[idKey].(string)
	params := map[string]interface{}{"id": id}
	if msgType == noticeMsgTypeSkill {
		// 消息中心跳转的 skill 均为共享技能（固定值）
		params["skillType"] = "shared"
	}
	return params
}
