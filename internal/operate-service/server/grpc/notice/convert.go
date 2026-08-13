package notice

import (
	"encoding/json"
	"strings"

	operate_service "github.com/UnicomAI/wanwu/api/proto/operate-service"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/iam"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/orm"
	"github.com/UnicomAI/wanwu/pkg/util"
)

// parseActionParams 解析 actionParams JSON 字符串；失败返回 nil（后续 schema 校验会剔除该 action）
func parseActionParams(raw string) map[string]interface{} {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		return nil
	}
	return params
}

// marshalActionParams 把 actionParams 序列化为 JSON 字符串；失败返回空对象
func marshalActionParams(params map[string]interface{}) string {
	if params == nil {
		return "{}"
	}
	b, err := json.Marshal(params)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func toUserOrgs(pairs []*operate_service.NoticeUserOrgPair) []iam.UserOrg {
	if len(pairs) == 0 {
		return nil
	}
	ret := make([]iam.UserOrg, 0, len(pairs))
	for _, p := range pairs {
		ret = append(ret, iam.UserOrg{UserID: p.UserId, OrgID: p.OrgId})
	}
	return ret
}

// toOrmActions 把 proto 的 actions（actionParams 是 JSON 字符串）转成 orm 结构。
// actionParams 解析失败的 action 会带上空参数进入 schema 校验，从而被剔除——
// 不在这里报错，消息主体仍然有价值。
func toOrmActions(actions []*operate_service.NoticeAction) []orm.Action {
	if len(actions) == 0 {
		return nil
	}
	ret := make([]orm.Action, 0, len(actions))
	for _, a := range actions {
		ret = append(ret, orm.Action{
			MsgType:      a.MsgType,
			ActionType:   a.ActionType,
			ActionParams: parseActionParams(a.ActionParams),
		})
	}
	return ret
}

// toProtoActions 把落库的 actions JSON 解析并下发。
// actionParams 以 JSON 字符串透传，由 bff 解析成对象后交给前端（前端无需 JSON.parse）。
func toProtoActions(raw string) []*operate_service.NoticeAction {
	parsed := orm.ParseActions(raw)
	ret := make([]*operate_service.NoticeAction, 0, len(parsed))
	for _, a := range parsed {
		ret = append(ret, &operate_service.NoticeAction{
			MsgType:      a.MsgType,
			ActionType:   a.ActionType,
			ActionParams: marshalActionParams(a.ActionParams),
		})
	}
	return ret
}

func toProtoNoticeItem(item *orm.NoticeItem) *operate_service.NoticeItem {
	return &operate_service.NoticeItem{
		MessageId:  util.Int2Str(item.MessageID),
		Title:      item.Title,
		Category:   item.Category,
		Content:    item.Content,
		Actions:    toProtoActions(item.Actions),
		ReceivedAt: item.ReceivedAt,
		IsRead:     item.IsRead,
	}
}

func toProtoCreateResp(result *orm.CreateNoticeResult) *operate_service.CreateNoticeResp {
	return &operate_service.CreateNoticeResp{
		MessagesCreated:     result.MessagesCreated,
		AudienceRowsCreated: result.AudienceRowsCreated,
		IgnoredCount:        result.IgnoredCount,
	}
}

// toMessageIDs 把 proto 的 string 消息 ID 转成 int64。
// 非法 ID 直接丢弃——它们必然不可见，可见性过滤本来也会跳过。
func toMessageIDs(ids []string) []int64 {
	ret := make([]int64, 0, len(ids))
	for _, id := range ids {
		if v, err := util.I64(id); err == nil && v > 0 {
			ret = append(ret, v)
		}
	}
	return ret
}
