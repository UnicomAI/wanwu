// Package iam 是 operate-service 访问 iam-service 的 grpc 客户端封装。
//
// 消息中心的受众计算与读侧可见性都以 iam 为数据源（org_users + 启用状态），
// 这里只暴露消息域用得到的两项 P0 能力，抽成接口是为了 DAO 层单测可以 mock、不必真起 iam。
//
// 失败策略两者相反，调用方不可混用：
//   - GetUserOrgMembership 失败 → fail-open（joinedAt=0 放行 + 告警）。
//     代价仅为新成员短暂看到加入前通知，优于读路径整体不可用。
//   - ValidateUserOrgPairs 失败 → fail-closed（放弃该事件消息 + 打点），
//     但不向业务主流程抛错——消息生成本就是 best-effort 旁路。
package iam

import (
	"context"

	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
)

// UserOrg 用户-组织二元组
type UserOrg struct {
	UserID string
	OrgID  string
}

// ValidateResult 名单入口清洗结果
type ValidateResult struct {
	// Pairs 去重 + 成员校验 + 禁用过滤后的合法二元组
	Pairs []UserOrg
	// Exceeded 名单超过 S 上限——拒绝生成该事件消息（打点告警），不阻塞业务主流程
	Exceeded bool
	// IgnoredCount 被忽略的二元组数量（打点用）
	IgnoredCount int32
}

// Membership 单用户在指定组织的成员关系
type Membership struct {
	Exists   bool
	JoinedAt int64 // org_users.created_at（毫秒）；fail-open 时为 0
	Active   bool
}

type IClient interface {
	// ValidateUserOrgPairs 验证 + 去重 + 过滤禁用 + 强制 S 上限（limit<=0 表示不限）
	ValidateUserOrgPairs(ctx context.Context, pairs []UserOrg, limit int32) (*ValidateResult, error)
	// GetUserOrgMembership 查询 joinedAt 与成员状态，供读侧 «vis» 的加入时间屏蔽使用
	GetUserOrgMembership(ctx context.Context, userID, orgID string) (*Membership, error)
	// ListUserOrgs 查询单账号的全部归属组织，供公告 TargetScope.userIds 展开为二元组。
	// 复用 iam 已有的 GetOrgSelect，不需要 IAM 额外交付能力。
	ListUserOrgs(ctx context.Context, userID string) ([]string, error)
}

type Client struct {
	iam iam_service.IAMServiceClient
}

func NewClient(host string) (*Client, error) {
	conn, err := trace_util.NewGrpcTracerConn(host, nil)
	if err != nil {
		return nil, err
	}
	return &Client{iam: iam_service.NewIAMServiceClient(conn)}, nil
}

func (c *Client) ValidateUserOrgPairs(ctx context.Context, pairs []UserOrg, limit int32) (*ValidateResult, error) {
	if len(pairs) == 0 {
		return &ValidateResult{}, nil
	}
	req := &iam_service.ValidateUserOrgPairsReq{Limit: limit}
	for _, p := range pairs {
		req.Pairs = append(req.Pairs, &iam_service.UserOrgPair{UserId: p.UserID, OrgId: p.OrgID})
	}
	resp, err := c.iam.ValidateUserOrgPairs(ctx, req)
	if err != nil {
		return nil, err
	}
	ret := &ValidateResult{
		Exceeded:     resp.Exceeded,
		IgnoredCount: resp.FilteredCount,
	}
	for _, p := range resp.Pairs {
		ret.Pairs = append(ret.Pairs, UserOrg{UserID: p.UserId, OrgID: p.OrgId})
	}
	return ret, nil
}

func (c *Client) ListUserOrgs(ctx context.Context, userID string) ([]string, error) {
	resp, err := c.iam.GetOrgSelect(ctx, &iam_service.GetOrgSelectReq{UserId: userID})
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0, len(resp.Selects))
	for _, org := range resp.Selects {
		ret = append(ret, org.Id)
	}
	return ret, nil
}

func (c *Client) GetUserOrgMembership(ctx context.Context, userID, orgID string) (*Membership, error) {
	resp, err := c.iam.GetUserOrgMembership(ctx, &iam_service.GetUserOrgMembershipReq{
		UserId: userID,
		OrgId:  orgID,
	})
	if err != nil {
		return nil, err
	}
	return &Membership{
		Exists:   resp.Exists,
		JoinedAt: resp.JoinedAt,
		Active:   resp.Active,
	}, nil
}
