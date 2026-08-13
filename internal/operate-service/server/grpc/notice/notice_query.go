package notice

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	operate_service "github.com/UnicomAI/wanwu/api/proto/operate-service"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/orm"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GetUnreadCount 未读总数 + 分类角标，驱动头像红点与各 Tab 角标。
// 消息按 userId + orgId 双维度隔离，切换组织需重新拉取。
func (s *Service) GetUnreadCount(ctx context.Context, req *operate_service.GetUnreadCountReq) (*operate_service.GetUnreadCountResp, error) {
	count, err := s.cli.GetNoticeUnreadCount(ctx, req.UserId, req.OrgId)
	if err != nil {
		return nil, errStatus(errs.Code_OperateNotice, err)
	}
	return &operate_service.GetUnreadCountResp{
		Total:      count.Total,
		ByCategory: count.ByCategory,
	}, nil
}

// ListNotice 整页列表：已读与未读混排，每条带 isRead。
// 悬浮面板的未读列表也复用本接口（onlyUnread=true + 较小 pageSize）。
func (s *Service) ListNotice(ctx context.Context, req *operate_service.ListNoticeReq) (*operate_service.ListNoticeResp, error) {
	items, total, err := s.cli.ListNotice(ctx, &orm.ListNoticeReq{
		UserID:     req.UserId,
		OrgID:      req.OrgId,
		Category:   req.Category,
		OnlyUnread: req.OnlyUnread,
		Keyword:    req.Keyword,
		PageNo:     req.PageNo,
		PageSize:   req.PageSize,
	})
	if err != nil {
		return nil, errStatus(errs.Code_OperateNotice, err)
	}
	resp := &operate_service.ListNoticeResp{Total: total}
	for _, item := range items {
		resp.List = append(resp.List, toProtoNoticeItem(item))
	}
	return resp, nil
}

// ReadNotice 单条已读。重复对已读消息调用是幂等的，不会报错。
func (s *Service) ReadNotice(ctx context.Context, req *operate_service.ReadNoticeReq) (*emptypb.Empty, error) {
	messageID, convErr := util.I64(req.MessageId)
	if convErr != nil || messageID <= 0 {
		// 非法 ID 必然不可见，与"不可见 ID 静默跳过"同口径处理（防枚举）
		return &emptypb.Empty{}, nil
	}
	if err := s.cli.ReadNotice(ctx, req.UserId, req.OrgId, messageID); err != nil {
		return nil, errStatus(errs.Code_OperateNotice, err)
	}
	return &emptypb.Empty{}, nil
}

// ReadAllNotice 一键已读：作用于当前账号在当前组织上下文的全部未读，不分类别。
// 水位线实现，未读量不影响响应时间。
func (s *Service) ReadAllNotice(ctx context.Context, req *operate_service.ReadAllNoticeReq) (*emptypb.Empty, error) {
	if err := s.cli.ReadAllNotice(ctx, req.UserId, req.OrgId); err != nil {
		return nil, errStatus(errs.Code_OperateNotice, err)
	}
	return &emptypb.Empty{}, nil
}

// DeleteNotice 批量软删：只影响当前账号在当前组织上下文的列表。
// 不可见的 messageId 静默跳过（防越权、防枚举），调用方可用 affectedCount 与请求数量差值感知。
func (s *Service) DeleteNotice(ctx context.Context, req *operate_service.DeleteNoticeReq) (*operate_service.DeleteNoticeResp, error) {
	affected, err := s.cli.DeleteNotice(ctx, req.UserId, req.OrgId, toMessageIDs(req.MessageIds))
	if err != nil {
		return nil, errStatus(errs.Code_OperateNotice, err)
	}
	return &operate_service.DeleteNoticeResp{AffectedCount: affected}, nil
}
