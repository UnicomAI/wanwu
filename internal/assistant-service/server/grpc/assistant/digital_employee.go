package assistant

import (
	"context"
	"encoding/json"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

// DigitalEmployeeConversationCreate 创建数字员工发布会话
func (s *Service) DigitalEmployeeConversationCreate(ctx context.Context, req *assistant_service.DigitalEmployeeConversationCreateReq) (*assistant_service.DigitalEmployeeConversationCreateResp, error) {
	threadId := util.GenUUID()

	var modelConfigStr string
	if req.ModelConfig != nil {
		modelConfigBytes, _ := json.Marshal(req.ModelConfig)
		modelConfigStr = string(modelConfigBytes)
	} else {
		modelConfigStr = "null"
	}

	config := &model.DigitalEmployeeConversationConfig{
		ThreadID:    threadId,
		EmployeeID:  req.EmployeeId,
		Title:       req.Prompt,
		UserID:      req.Identity.UserId,
		OrgID:       req.Identity.OrgId,
		ModelConfig: modelConfigStr,
	}

	if status := s.cli.CreateDigitalEmployeeConversationConfig(ctx, config); status != nil {
		return nil, errStatus(errs.Code_WgaConversationGetErr, status)
	}

	return &assistant_service.DigitalEmployeeConversationCreateResp{
		ThreadId: threadId,
	}, nil
}

// DigitalEmployeeConversationDelete 删除数字员工发布会话
func (s *Service) DigitalEmployeeConversationDelete(ctx context.Context, req *assistant_service.DigitalEmployeeConversationDeleteReq) (*emptypb.Empty, error) {
	if status := s.cli.DeleteDigitalEmployeeConversationConfig(ctx, req.ThreadId, req.Identity.UserId, req.Identity.OrgId); status != nil {
		return nil, errStatus(errs.Code_WgaConversationGetErr, status)
	}
	return &emptypb.Empty{}, nil
}

// DigitalEmployeeConversationList 获取数字员工发布会话列表（按 employeeId 维度）
func (s *Service) DigitalEmployeeConversationList(ctx context.Context, req *assistant_service.DigitalEmployeeConversationListReq) (*assistant_service.DigitalEmployeeConversationListResp, error) {
	offset := (req.PageNo - 1) * req.PageSize

	configs, total, status := s.cli.GetDigitalEmployeeConversationConfigList(ctx, req.Identity.UserId, req.Identity.OrgId, req.EmployeeId, req.SearchText, offset, req.PageSize)
	if status != nil {
		return nil, errStatus(errs.Code_WgaConversationGetErr, status)
	}

	var infos []*assistant_service.DigitalEmployeeConversationInfo
	for _, config := range configs {
		infos = append(infos, &assistant_service.DigitalEmployeeConversationInfo{
			ThreadId:   config.ThreadID,
			EmployeeId: config.EmployeeID,
			Title:      config.Title,
			CreatedAt:  config.CreatedAt,
			UpdatedAt:  config.UpdatedAt,
		})
	}

	return &assistant_service.DigitalEmployeeConversationListResp{
		Data:     infos,
		Total:    total,
		PageSize: req.PageSize,
		PageNo:   req.PageNo,
	}, nil
}

// GetDigitalEmployeeConversationConfig 获取数字员工发布会话配置（含 modelConfig，供对话时读回）
func (s *Service) GetDigitalEmployeeConversationConfig(ctx context.Context, req *assistant_service.GetDigitalEmployeeConversationConfigReq) (*assistant_service.GetDigitalEmployeeConversationConfigResp, error) {
	config, status := s.cli.GetDigitalEmployeeConversationConfig(ctx, req.ThreadId, req.Identity.UserId, req.Identity.OrgId)
	if status != nil {
		return nil, errStatus(errs.Code_WgaConversationGetErr, status)
	}

	respConfig := &assistant_service.DigitalEmployeeConversationConfig{
		ThreadId:   config.ThreadID,
		EmployeeId: config.EmployeeID,
		Title:      config.Title,
		UserId:     config.UserID,
		OrgId:      config.OrgID,
		CreatedAt:  config.CreatedAt,
		UpdatedAt:  config.UpdatedAt,
	}
	// modelConfig 反序列化为 proto
	if config.ModelConfig != "" && config.ModelConfig != "null" {
		var mc common.AppModelConfig
		if json.Unmarshal([]byte(config.ModelConfig), &mc) == nil {
			respConfig.ModelConfig = &mc
		}
	}

	return &assistant_service.GetDigitalEmployeeConversationConfigResp{
		Config: respConfig,
	}, nil
}

// TouchDigitalEmployeeConversationConfig 刷新数字员工发布会话的 updated_at（列表按 updated_at DESC 排序）
func (s *Service) TouchDigitalEmployeeConversationConfig(ctx context.Context, req *assistant_service.GetDigitalEmployeeConversationConfigReq) (*emptypb.Empty, error) {
	if status := s.cli.TouchDigitalEmployeeConversationConfig(ctx, req.ThreadId, req.Identity.UserId, req.Identity.OrgId); status != nil {
		return nil, errStatus(errs.Code_WgaConversationGetErr, status)
	}
	return &emptypb.Empty{}, nil
}

// UpdateDigitalEmployeeConversationConfig 更新数字员工发布会话的模型配置（同时刷新 updated_at）
func (s *Service) UpdateDigitalEmployeeConversationConfig(ctx context.Context, req *assistant_service.UpdateDigitalEmployeeConversationConfigReq) (*emptypb.Empty, error) {
	var modelConfigStr string
	if req.ModelConfig != nil {
		modelConfigBytes, _ := json.Marshal(req.ModelConfig)
		modelConfigStr = string(modelConfigBytes)
	} else {
		modelConfigStr = "null"
	}
	if status := s.cli.UpdateDigitalEmployeeConversationConfig(ctx, req.ThreadId, modelConfigStr, req.Identity.UserId, req.Identity.OrgId); status != nil {
		return nil, errStatus(errs.Code_WgaConversationGetErr, status)
	}
	return &emptypb.Empty{}, nil
}
