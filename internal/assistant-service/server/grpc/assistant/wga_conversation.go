package assistant

import (
	"context"
	"encoding/json"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

// WgaConversationCreate 创建WGA对话
func (s *Service) WgaConversationCreate(ctx context.Context, req *assistant_service.WgaConversationCreateReq) (*assistant_service.WgaConversationCreateResp, error) {
	// 创建对话配置
	threadId := util.GenUUID()

	var modelConfigStr string
	if req.ModelConfig != nil {
		modelConfigBytes, _ := json.Marshal(req.ModelConfig)
		modelConfigStr = string(modelConfigBytes)
	} else {
		// 设置为有效的 JSON null 值
		modelConfigStr = "null"
	}

	config := &model.WgaConversationConfig{
		ThreadID:    threadId,
		Title:       req.Prompt,
		UserID:      req.Identity.UserId,
		OrgID:       req.Identity.OrgId,
		ModelConfig: modelConfigStr,
	}

	if status := s.cli.CreateWgaConversationConfig(ctx, config); status != nil {
		return nil, errStatus(errs.Code_WgaConversationGetErr, status)
	}

	return &assistant_service.WgaConversationCreateResp{
		ThreadId: threadId,
	}, nil
}

// WgaConversationDelete 删除WGA对话
func (s *Service) WgaConversationDelete(ctx context.Context, req *assistant_service.WgaConversationDeleteReq) (*emptypb.Empty, error) {
	if status := s.cli.DeleteWgaConversationConfig(ctx, req.ThreadId); status != nil {
		return nil, errStatus(errs.Code_WgaConversationGetErr, status)
	}
	return &emptypb.Empty{}, nil
}

// WgaConversationList 获取WGA对话列表
func (s *Service) WgaConversationList(ctx context.Context, req *assistant_service.WgaConversationListReq) (*assistant_service.WgaConversationListResp, error) {
	offset := (req.PageNo - 1) * req.PageSize

	configs, total, status := s.cli.GetWgaConversationConfigList(ctx, req.Identity.UserId, req.Identity.OrgId, offset, req.PageSize)
	if status != nil {
		return nil, errStatus(errs.Code_WgaConversationGetErr, status)
	}

	var conversationInfos []*assistant_service.WgaConversationInfo
	for _, config := range configs {
		conversationInfos = append(conversationInfos, &assistant_service.WgaConversationInfo{
			ThreadId:  config.ThreadID,
			Title:     config.Title,
			CreatedAt: config.CreatedAt,
			UpdatedAt: config.UpdatedAt,
		})
	}

	return &assistant_service.WgaConversationListResp{
		Data:     conversationInfos,
		Total:    total,
		PageSize: req.PageSize,
		PageNo:   req.PageNo,
	}, nil
}

// WgaConversationExists 检查WGA对话是否存在
func (s *Service) WgaConversationExists(ctx context.Context, req *assistant_service.WgaConversationExistsReq) (*assistant_service.WgaConversationExistsResp, error) {
	exists, status := s.cli.WgaConversationConfigExists(ctx, req.ThreadId, req.Identity.UserId, req.Identity.OrgId)
	if status != nil {
		return nil, errStatus(errs.Code_WgaConversationGetErr, status)
	}
	return &assistant_service.WgaConversationExistsResp{
		Exists: exists,
	}, nil
}
