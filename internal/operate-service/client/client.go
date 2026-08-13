package client

import (
	"context"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/orm"
)

type IClient interface {
	// 系统自定义配置
	CreateSystemCustom(ctx context.Context, userID, orgID string, key orm.SystemCustomKey, mode orm.SystemCustomMode, custom orm.SystemCustom) *err_code.Status
	GetSystemCustom(ctx context.Context, mode orm.SystemCustomMode) (*orm.SystemCustom, *err_code.Status)

	AddClientRecord(ctx context.Context, clientId string) *err_code.Status
	GetClientStatistic(ctx context.Context, startDate, endDate string) (*orm.ClientStatistic, *err_code.Status)

	// --- 消息中心 · 写入（按业务场景拆分） ---

	// CreateAppNotice 应用：发布/改范围/发新版本/删除（scope 差分）
	CreateAppNotice(ctx context.Context, req *orm.AppNoticeReq) (*orm.CreateNoticeResult, *err_code.Status)
	// CreateModelNotice 模型：导入/删除/停用启用（scope 差分）
	CreateModelNotice(ctx context.Context, req *orm.ModelNoticeReq) (*orm.CreateNoticeResult, *err_code.Status)
	// CreateKnowledgeNotice 知识库：共享/取消共享/改权限/删除（名单差分）
	CreateKnowledgeNotice(ctx context.Context, req *orm.KnowledgeNoticeReq) (*orm.CreateNoticeResult, *err_code.Status)

	// --- 消息中心 · 查询/操作（userId + orgId 双维度隔离） ---

	// GetNoticeUnreadCount 未读总数 + 分类角标
	GetNoticeUnreadCount(ctx context.Context, userID, orgID string) (*orm.NoticeUnreadCount, *err_code.Status)
	// ListNotice 整页列表（已读+未读混排，每条带 isRead）
	ListNotice(ctx context.Context, req *orm.ListNoticeReq) ([]*orm.NoticeItem, int64, *err_code.Status)
	// ReadNotice 单条已读
	ReadNotice(ctx context.Context, userID, orgID string, messageID int64) *err_code.Status
	// ReadAllNotice 一键已读（全部未读，不分类别；水位线实现）
	ReadAllNotice(ctx context.Context, userID, orgID string) *err_code.Status
	// DeleteNotice 批量软删，返回实际删除条数
	DeleteNotice(ctx context.Context, userID, orgID string, messageIDs []int64) (int32, *err_code.Status)
}
