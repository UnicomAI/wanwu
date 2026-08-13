package orm

import (
	"context"

	"github.com/UnicomAI/wanwu/internal/operate-service/client/iam"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/log"
)

// 发布范围（scope）归一化取值。
// bff 包装层负责把应用 PublishType（private/public/organization）与
// 模型 ScopeType（1私有/2公开/3组织）都归一到这一组字符串——受众映射对资源类型无感知
// 是本设计的核心不变量，两套枚举不允许同时进到消息域。
const (
	ScopeNone         = ""             // 删除 / 停用 / 取消发布：受众为空集
	ScopePrivate      = "private"      // 转私有：受众为空集
	ScopeOrganization = "organization" // 组织内公开，不级联子组织
	ScopePublic       = "public"       // 全局公开
	ScopeSpecific     = "specific"     // 特定用户名单（预留，本期规则型无生产者）
)

// audienceSpec 一个发布范围的受众表达方式（物理落表）。
//
// 落表方式为静态规则，由载体唯一决定、写死在代码里：
//   - 载体是组织（AT=2）→ Orgs 落 notice_message_orgs 正向组织行（每个目标组织一行）
//   - 载体是全局（AT=4）→ 纯字段判别，0 行（全员可见，无排除——排除语义已随产品设计整体移除）
//   - 载体是特定用户名单（AT=3）→ Pairs 落 notice_message_audience 正向名单行
type audienceSpec struct {
	AudienceType int8
	// Orgs AT=2 的目标组织列表（当前场景恒为归属组织单元素）
	Orgs []string
	// Pairs AT=3 名单行
	Pairs []iam.UserOrg
}

// scopeSpec 单一发布范围 → 受众；private/"" → nil（空受众，调用方跳过整条消息）。
func scopeSpec(scope, orgID string) *audienceSpec {
	switch scope {
	case ScopeOrganization:
		return &audienceSpec{AudienceType: model.AudienceTypeOrg, Orgs: []string{orgID}}
	case ScopePublic:
		return &audienceSpec{AudienceType: model.AudienceTypeGlobal}
	default:
		// ScopeNone / ScopePrivate → 空集
		return nil
	}
}

// specificSpec 名单型受众：载体为特定用户名单，恒落 notice_message_audience 正向行。空集返回 nil（跳过整条消息）。
func specificSpec(pairs []iam.UserOrg) *audienceSpec {
	if len(pairs) == 0 {
		return nil
	}
	return &audienceSpec{AudienceType: model.AudienceTypeSpecific, Pairs: pairs}
}

// cleanAudiencePairs 名单入口统一清洗：剔除操作者本人 → 去重 → ValidateUserOrgPairs
// 成员校验 → 禁用过滤 → 强制 S 上限。
//
// 返回 (合法名单, 被忽略数量, 是否超限, 校验错误)。
// 无效/超限只影响消息生成（忽略 + 打点 + 返回被忽略计数），不向业务主流程抛错——
// 消息生成本就是 best-effort 旁路。iam 不可用时 fail-closed：返回 error 让调用方
// 放弃整个事件（同一事件的多组名单只校验一次 IAM，任一组失败即整事件终止，杜绝部分投递）。
// 仅名单为空（无需校验）时不返回 error。
func (c *Client) cleanAudiencePairs(ctx context.Context, pairs []iam.UserOrg, senderID, senderOrgID string) ([]iam.UserOrg, int32, bool, error) {
	if len(pairs) == 0 {
		return nil, 0, false, nil
	}
	var ignored int32
	seen := make(map[iam.UserOrg]struct{}, len(pairs))
	dedup := make([]iam.UserOrg, 0, len(pairs))
	for _, p := range pairs {
		if p.UserID == "" || p.OrgID == "" {
			ignored++
			continue
		}
		// 操作者本人默认不收到自己触发的消息（二元组级）
		if p.UserID == senderID && p.OrgID == senderOrgID {
			continue
		}
		if _, ok := seen[p]; ok {
			continue // 去重不计入 ignored
		}
		seen[p] = struct{}{}
		dedup = append(dedup, p)
	}
	if len(dedup) == 0 {
		return nil, ignored, false, nil
	}

	result, err := c.iam.ValidateUserOrgPairs(ctx, dedup, int32(c.notice.MaxAudienceSize))
	if err != nil {
		// fail-closed：返回 error，调用方据此放弃整个事件的消息生成
		log.Errorf("notice: validate user-org pairs failed, skip notice generation: %v", err)
		return nil, ignored + int32(len(dedup)), false, err
	}
	if result.Exceeded {
		log.Errorf("notice: audience list exceeds max size %v (got %v), skip notice generation",
			c.notice.MaxAudienceSize, len(dedup))
		return nil, ignored + int32(len(dedup)), true, nil
	}
	return result.Pairs, ignored + result.IgnoredCount, false, nil
}
