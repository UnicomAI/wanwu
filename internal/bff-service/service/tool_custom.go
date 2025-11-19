package service

import (
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

func CreateCustomTool(ctx *gin.Context, userID, orgID string, req request.CustomToolCreate) error {
	if err := openapi3_util.ValidateSchema(ctx.Request.Context(), []byte(req.Schema)); err != nil {
		return grpc_util.ErrorStatus(errs.Code_BFFInvalidArg, err.Error())
	}
	_, err := mcp.CreateCustomTool(ctx.Request.Context(), &mcp_service.CreateCustomToolReq{
		AvatarPath:    req.Avatar.Key,
		Schema:        req.Schema,
		Name:          req.Name,
		Description:   req.Description,
		PrivacyPolicy: req.PrivacyPolicy,
		ApiAuth: &common.ApiAuthWebRequest{
			AuthType:           req.ApiAuth.AuthType,
			ApiKeyHeaderPrefix: req.ApiAuth.ApiKeyHeaderPrefix,
			ApiKeyHeader:       req.ApiAuth.ApiKeyHeader,
			ApiKeyQueryParam:   req.ApiAuth.ApiKeyQueryParam,
			ApiKeyValue:        req.ApiAuth.ApiKeyValue,
		},
		Identity: &mcp_service.Identity{
			UserId: userID,
			OrgId:  orgID,
		},
	})
	return err
}

func GetCustomTool(ctx *gin.Context, userID, orgID string, customToolId string) (*response.CustomToolDetail, error) {
	info, err := mcp.GetCustomToolInfo(ctx.Request.Context(), &mcp_service.GetCustomToolInfoReq{
		CustomToolId: customToolId,
	})
	if err != nil {
		return nil, err
	}
	doc, err := openapi3_util.LoadFromData(ctx.Request.Context(), []byte(info.Schema))
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
	}
	return &response.CustomToolDetail{
		CustomToolInfo: response.CustomToolInfo{
			CustomToolId: info.CustomToolId,
			Name:         info.Name,
			Description:  info.Description,
			Avatar:       cacheToolAvatar(ctx, constant.ToolTypeCustom, info.AvatarPath),
		},
		Schema:        info.Schema,
		PrivacyPolicy: info.PrivacyPolicy,
		ApiAuth: util.ApiAuthWebRequest{
			AuthType:           info.ApiAuth.AuthType,
			ApiKeyHeaderPrefix: info.ApiAuth.ApiKeyHeaderPrefix,
			ApiKeyHeader:       info.ApiAuth.ApiKeyHeader,
			ApiKeyQueryParam:   info.ApiAuth.ApiKeyQueryParam,
			ApiKeyValue:        info.ApiAuth.ApiKeyValue,
		},
		ApiList: openapiSchema2ToolList(doc),
	}, nil
}

func DeleteCustomTool(ctx *gin.Context, userID, orgID string, req request.CustomToolIDReq) error {
	// 删除智能体AssistantCustom中记录 [EN] Delete the record in the agent AssistantCustom
	_, err := assistant.AssistantToolDeleteByToolId(ctx.Request.Context(), &assistant_service.AssistantToolDeleteByToolIdReq{
		ToolId:   req.CustomToolID,
		ToolType: constant.ToolTypeCustom,
	})
	if err != nil {
		return err
	}

	_, err = mcp.DeleteCustomTool(ctx.Request.Context(), &mcp_service.DeleteCustomToolReq{
		CustomToolId: req.CustomToolID,
	})
	return err
}

func UpdateCustomTool(ctx *gin.Context, userID, orgID string, req request.CustomToolUpdateReq) error {
	if err := openapi3_util.ValidateSchema(ctx.Request.Context(), []byte(req.Schema)); err != nil {
		return grpc_util.ErrorStatus(errs.Code_BFFInvalidArg, err.Error())
	}
	_, err := mcp.UpdateCustomTool(ctx.Request.Context(), &mcp_service.UpdateCustomToolReq{
		CustomToolId: req.CustomToolID,
		AvatarPath:   req.Avatar.Key,
		Name:         req.Name,
		Description:  req.Description,
		ApiAuth: &common.ApiAuthWebRequest{
			AuthType:           req.ApiAuth.AuthType,
			ApiKeyHeaderPrefix: req.ApiAuth.ApiKeyHeaderPrefix,
			ApiKeyHeader:       req.ApiAuth.ApiKeyHeader,
			ApiKeyQueryParam:   req.ApiAuth.ApiKeyQueryParam,
			ApiKeyValue:        req.ApiAuth.ApiKeyValue,
		},
		Schema:        req.Schema,
		PrivacyPolicy: req.PrivacyPolicy,
	})
	return err
}

func GetCustomToolList(ctx *gin.Context, userID, orgID, name string) (*response.ListResult, error) {
	resp, err := mcp.GetCustomToolList(ctx.Request.Context(), &mcp_service.GetCustomToolListReq{
		Identity: &mcp_service.Identity{
			UserId: userID,
			OrgId:  orgID,
		},
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	var list []response.CustomToolInfo
	for _, item := range resp.List {
		list = append(list, response.CustomToolInfo{
			CustomToolId: item.CustomToolId,
			Name:         item.Name,
			Description:  item.Description,
			Avatar:       cacheToolAvatar(ctx, constant.ToolTypeCustom, item.AvatarPath),
		})
	}
	return &response.ListResult{
		List:  list,
		Total: int64(len(list)),
	}, nil
}

func GetCustomToolActions(ctx *gin.Context, userID, orgID string, req request.CustomToolSchemaReq) (*response.ListResult, error) {
	doc, err := openapi3_util.LoadFromData(ctx.Request.Context(), []byte(req.Schema))
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFInvalidArg, err.Error())
	}
	list := openapiSchema2ToolList(doc)
	return &response.ListResult{
		List:  list,
		Total: int64(len(list)),
	}, nil
}

// --- internal ---

func openapiSchema2ToolList(doc *openapi3.T) []response.CustomToolActionInfo {
	var list []response.CustomToolActionInfo
	for path, pathItem := range doc.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			list = append(list, response.CustomToolActionInfo{
				Name:   operation.OperationID,
				Desc:   operation.Description,
				Method: method,
				Path:   path,
			})
		}
	}
	return list
}
