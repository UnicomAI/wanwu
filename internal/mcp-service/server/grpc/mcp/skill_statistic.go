package mcp

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) IncrementCustomSkillDownloadCount(ctx context.Context, req *mcp_service.IncrementCustomSkillDownloadCountReq) (*emptypb.Empty, error) {
	if status := s.cli.IncrementCustomSkillDownloadCount(ctx, req.GetSkillId()); status != nil {
		return nil, errStatus(errs.Code_MCPGeneral, status)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) IncrementCustomSkillAcquiredCount(ctx context.Context, req *mcp_service.IncrementCustomSkillAcquiredCountReq) (*emptypb.Empty, error) {
	if status := s.cli.IncrementCustomSkillAcquiredCount(ctx, req.GetSkillId()); status != nil {
		return nil, errStatus(errs.Code_MCPGeneral, status)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) IncrementBuiltinSkillDownloadCount(ctx context.Context, req *mcp_service.IncrementBuiltinSkillDownloadCountReq) (*emptypb.Empty, error) {
	if status := s.cli.IncrementBuiltinSkillDownloadCount(ctx, req.GetSkillId()); status != nil {
		return nil, errStatus(errs.Code_MCPGeneral, status)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetBuiltinSkillDownloadCounts(ctx context.Context, req *mcp_service.GetBuiltinSkillDownloadCountsReq) (*mcp_service.GetBuiltinSkillDownloadCountsResp, error) {
	countMap, status := s.cli.GetBuiltinSkillDownloadCounts(ctx, req.GetSkillIds())
	if status != nil {
		return nil, errStatus(errs.Code_MCPGeneral, status)
	}
	return &mcp_service.GetBuiltinSkillDownloadCountsResp{
		CountMap: countMap,
	}, nil
}
