package notice

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	operate_service "github.com/UnicomAI/wanwu/api/proto/operate-service"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/orm"
)

// CreateAppNotice 应用消息：发布/改范围/发新版本/删除。
// oldScope/newScope 由 bff 归一化后传入（应用 PublishType）。
func (s *Service) CreateAppNotice(ctx context.Context, req *operate_service.CreateAppNoticeReq) (*operate_service.CreateNoticeResp, error) {
	result, err := s.cli.CreateAppNotice(ctx, &orm.AppNoticeReq{
		EventID:     req.EventId,
		AppType:     req.AppType,
		AppID:       req.AppId,
		AppName:     req.AppName,
		OrgID:       req.OrgId,
		OldScope:    req.OldScope,
		NewScope:    req.NewScope,
		Version:     req.Version,
		SenderID:    req.SenderId,
		SenderOrgID: req.SenderOrgId,
	})
	if err != nil {
		return nil, errStatus(errs.Code_OperateNotice, err)
	}
	return toProtoCreateResp(result), nil
}

// CreateModelNotice 模型消息：导入/删除/停用启用。
// oldScope/newScope 由 bff 归一化后传入（模型 ScopeType）。
func (s *Service) CreateModelNotice(ctx context.Context, req *operate_service.CreateModelNoticeReq) (*operate_service.CreateNoticeResp, error) {
	result, err := s.cli.CreateModelNotice(ctx, &orm.ModelNoticeReq{
		EventID:     req.EventId,
		ModelID:     req.ModelId,
		ModelName:   req.ModelName,
		OrgID:       req.OrgId,
		OldScope:    req.OldScope,
		NewScope:    req.NewScope,
		SenderID:    req.SenderId,
		SenderOrgID: req.SenderOrgId,
	})
	if err != nil {
		return nil, errStatus(errs.Code_OperateNotice, err)
	}
	return toProtoCreateResp(result), nil
}

// CreateKnowledgeNotice 知识库消息：共享/取消共享/改权限/删除。
func (s *Service) CreateKnowledgeNotice(ctx context.Context, req *operate_service.CreateKnowledgeNoticeReq) (*operate_service.CreateNoticeResp, error) {
	result, err := s.cli.CreateKnowledgeNotice(ctx, &orm.KnowledgeNoticeReq{
		EventID:       req.EventId,
		KnowledgeID:   req.KnowledgeId,
		KnowledgeName: req.KnowledgeName,
		Gained:        toUserOrgs(req.Gained),
		Lost:          toUserOrgs(req.Lost),
		Changed:       toUserOrgs(req.Changed),
		SenderID:      req.SenderId,
		SenderOrgID:   req.SenderOrgId,
		ChangedDetail: req.ChangedDetail,
		Variant:       req.Variant,
		Actions:       toOrmActions(req.Actions),
	})
	if err != nil {
		return nil, errStatus(errs.Code_OperateNotice, err)
	}
	return toProtoCreateResp(result), nil
}
