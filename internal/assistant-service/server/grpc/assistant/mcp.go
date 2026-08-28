package assistant

import (
	"context"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"google.golang.org/protobuf/types/known/emptypb"
)

// AssistantMCPCreate 添加mcp
func (s *Service) AssistantMCPCreate(ctx context.Context, req *assistant_service.AssistantMCPCreateReq) (*emptypb.Empty, error) {
	assistant, status := s.cli.GetAssistantByUuidWithPerm(ctx, req.AssistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantMCPErr, status)
	}
	assistantId := assistant.ID

	if status := s.cli.CreateAssistantMCP(ctx, assistantId, req.McpId, req.McpType, req.ActionName, req.Identity.UserId, req.Identity.OrgId); status != nil {
		return nil, errStatus(errs.Code_AssistantMCPErr, status)
	}

	return &emptypb.Empty{}, nil
}

// AssistantMCPDelete 删除mcp
func (s *Service) AssistantMCPDelete(ctx context.Context, req *assistant_service.AssistantMCPDeleteReq) (*emptypb.Empty, error) {
	assistant, status := s.cli.GetAssistantByUuidWithPerm(ctx, req.AssistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantMCPErr, status)
	}
	assistantId := assistant.ID

	if status := s.cli.DeleteAssistantMCP(ctx, assistantId, req.McpId, req.McpType, req.ActionName); status != nil {
		return nil, errStatus(errs.Code_AssistantMCPErr, status)
	}
	return &emptypb.Empty{}, nil
}

// AssistantMCPDeleteByMCPId 删除mcp
func (s *Service) AssistantMCPDeleteByMCPId(ctx context.Context, req *assistant_service.AssistantMCPDeleteByMCPIdReq) (*emptypb.Empty, error) {
	userId, orgId := req.Identity.UserId, req.Identity.OrgId
	if status := s.cli.DeleteAssistantMCPByMCPId(ctx, req.McpId, req.McpType, userId, orgId); status != nil {
		return nil, errStatus(errs.Code_AssistantMCPErr, status)
	}
	return &emptypb.Empty{}, nil
}

// AssistantMCPEnableSwitch mcp开关
func (s *Service) AssistantMCPEnableSwitch(ctx context.Context, req *assistant_service.AssistantMCPEnableSwitchReq) (*emptypb.Empty, error) {
	assistant, status := s.cli.GetAssistantByUuidWithPerm(ctx, req.AssistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantMCPErr, status)
	}
	assistantId := assistant.ID

	existingMCP, status := s.cli.GetAssistantMCP(ctx, assistantId, req.McpId, req.McpType, req.ActionName)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantMCPErr, status)
	}

	existingMCP.Enable = req.Enable
	if status = s.cli.UpdateAssistantMCP(ctx, existingMCP); status != nil {
		return nil, errStatus(errs.Code_AssistantMCPErr, status)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) AssistantMCPGetList(ctx context.Context, req *assistant_service.AssistantMCPGetListReq) (*assistant_service.AssistantMCPList, error) {
	assistant, status := s.cli.GetAssistantByUuidWithPerm(ctx, req.AssistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantMCPErr, status)
	}
	assistantId := assistant.ID

	mcpList, status := s.cli.GetAssistantMCPList(ctx, assistantId)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantMCPErr, status)
	}

	assistantMCPInfos := make([]*assistant_service.AssistantMCPInfo, len(mcpList))
	for i, mcp := range mcpList {
		assistantMCPInfos[i] = &assistant_service.AssistantMCPInfo{
			Id:             mcp.ID,
			McpAssistantId: assistant.UUID,
			McpId:          mcp.MCPId,
			Enable:         mcp.Enable,
			UserId:         mcp.UserId,
			OrgId:          mcp.OrgId,
			CreatedAt:      mcp.CreatedAt,
			UpdatedAt:      mcp.UpdatedAt,
		}
	}

	return &assistant_service.AssistantMCPList{
		AssistantMCPInfos: assistantMCPInfos,
	}, nil

}
