package mcp

import (
	"context"
	"encoding/json"

	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/db"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) CreateCustomTool(ctx context.Context, req *mcp_service.CreateCustomToolReq) (*emptypb.Empty, error) {
	if req.Identity == nil {
		return nil, errStatus(errs.Code_MCPCreateCustomToolErr, toErrStatus("mcp_create_custom_tool_err", "identity is empty"))
	}
	if req.ApiAuth == nil {
		return nil, errStatus(errs.Code_MCPCreateCustomToolErr, toErrStatus("mcp_create_custom_tool_err", "apiAuth is empty"))
	}
	apiAuthBytes, err := json.Marshal(req.ApiAuth)
	if err != nil {
		return nil, errStatus(errs.Code_MCPCreateCustomToolErr, toErrStatus("mcp_create_custom_tool_err", err.Error()))
	}
	if err := s.cli.CreateCustomTool(ctx, &model.CustomTool{
		AvatarPath:    req.AvatarPath,
		Name:          req.Name,
		Description:   db.LongText(req.Description),
		Schema:        db.LongText(req.Schema),
		PrivacyPolicy: db.LongText(req.PrivacyPolicy),
		AuthJSON:      db.LongText(apiAuthBytes),
		UserID:        req.Identity.UserId,
		OrgID:         req.Identity.OrgId,
	}); err != nil {
		return nil, errStatus(errs.Code_MCPCreateCustomToolErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetCustomToolInfo(ctx context.Context, req *mcp_service.GetCustomToolInfoReq) (*mcp_service.GetCustomToolInfoResp, error) {
	if req.CustomToolId == "" {
		return nil, errStatus(errs.Code_MCPGetCustomToolInfoErr, toErrStatus("mcp_get_custom_tool_info_err", "customToolId is empty"))
	}
	info, err := s.cli.GetCustomTool(ctx, &model.CustomTool{
		ID: util.MustU32(req.CustomToolId),
	})
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_MCPGetCustomToolInfoErr)
	}
	apiAuthJson := info.AuthJSON
	apiAuth := &common.ApiAuthWebRequest{}
	if err := json.Unmarshal([]byte(apiAuthJson), apiAuth); err != nil {
		return nil, errStatus(errs.Code_MCPGetCustomToolInfoErr, toErrStatus("mcp_get_custom_tool_info_err", err.Error()))
	}
	return &mcp_service.GetCustomToolInfoResp{
		CustomToolId:  util.Int2Str(info.ID),
		AvatarPath:    info.AvatarPath,
		Name:          info.Name,
		Description:   string(info.Description),
		Schema:        string(info.Schema),
		PrivacyPolicy: string(info.PrivacyPolicy),
		ApiAuth:       apiAuth,
	}, nil
}

func (s *Service) GetCustomToolList(ctx context.Context, req *mcp_service.GetCustomToolListReq) (*mcp_service.GetCustomToolListResp, error) {
	if req.Identity == nil {
		return nil, errStatus(errs.Code_MCPGetCustomToolListErr, toErrStatus("mcp_get_custom_tool_list_err", "identity is empty"))
	}
	infos, err := s.cli.ListCustomTools(ctx, req.Identity.OrgId, req.Identity.UserId, req.Name)
	if err != nil {
		return nil, errStatus(errs.Code_MCPGetCustomToolListErr, err)
	}
	list := make([]*mcp_service.GetCustomToolItem, 0)
	for _, info := range infos {
		list = append(list, &mcp_service.GetCustomToolItem{
			CustomToolId: util.Int2Str(info.ID),
			Name:         info.Name,
			Description:  string(info.Description),
			AvatarPath:   info.AvatarPath,
		})
	}
	return &mcp_service.GetCustomToolListResp{
		List: list,
	}, nil
}

func (s *Service) GetCustomToolByCustomToolIdList(ctx context.Context, req *mcp_service.GetCustomToolByCustomToolIdListReq) (*mcp_service.GetCustomToolListResp, error) {
	if len(req.CustomToolIdList) == 0 {
		return nil, errStatus(errs.Code_MCPGetCustomToolListErr, toErrStatus("mcp_get_custom_tool_list_err", "customToolIdList is empty"))
	}

	// 批量转换 string 为 uint32
	var ids []uint32
	for _, idStr := range req.CustomToolIdList {
		id := util.MustU32(idStr)
		ids = append(ids, id)
	}

	infos, err := s.cli.ListCustomToolsByCustomToolIDs(ctx, ids)
	if err != nil {
		return nil, errStatus(errs.Code_MCPGetCustomToolListErr, err)
	}
	list := make([]*mcp_service.GetCustomToolItem, 0)
	for _, info := range infos {
		list = append(list, &mcp_service.GetCustomToolItem{
			CustomToolId: util.Int2Str(info.ID),
			Name:         info.Name,
			Description:  string(info.Description),
		})
	}
	return &mcp_service.GetCustomToolListResp{
		List: list,
	}, nil
}

func (s *Service) UpdateCustomTool(ctx context.Context, req *mcp_service.UpdateCustomToolReq) (*emptypb.Empty, error) {
	if req.CustomToolId == "" {
		return nil, errStatus(errs.Code_MCPUpdateCustomToolErr, toErrStatus("mcp_update_custom_tool_err", "customToolId is empty"))
	}
	if req.ApiAuth == nil {
		return nil, errStatus(errs.Code_MCPUpdateCustomToolErr, toErrStatus("mcp_update_custom_tool_err", "apiAuth is empty"))
	}
	apiAuthBytes, err := json.Marshal(req.ApiAuth)
	if err != nil {
		return nil, errStatus(errs.Code_MCPUpdateCustomToolErr, toErrStatus("mcp_update_custom_tool_err", err.Error()))
	}
	if err := s.cli.UpdateCustomTool(ctx, &model.CustomTool{
		ID:            util.MustU32(req.CustomToolId),
		AvatarPath:    req.AvatarPath,
		Name:          req.Name,
		Description:   db.LongText(req.Description),
		Schema:        db.LongText(req.Schema),
		PrivacyPolicy: db.LongText(req.PrivacyPolicy),
		AuthJSON:      db.LongText(apiAuthBytes),
	}); err != nil {
		return nil, errStatus(errs.Code_MCPUpdateCustomToolErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteCustomTool(ctx context.Context, req *mcp_service.DeleteCustomToolReq) (*emptypb.Empty, error) {
	if req.CustomToolId == "" {
		return nil, errStatus(errs.Code_MCPDeleteCustomToolErr, toErrStatus("mcp_delete_custom_tool_err", "customToolId is empty"))
	}
	if err := s.cli.DeleteCustomTool(ctx, util.MustU32(req.CustomToolId)); err != nil {
		return nil, errStatus(errs.Code_MCPDeleteCustomToolErr, err)
	}
	return &emptypb.Empty{}, nil
}
