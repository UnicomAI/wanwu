package service

import (
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// GetKnowledgeExternalAPIList 获取外部知识库API列表
func GetKnowledgeExternalAPIList(ctx *gin.Context, userId, orgId string) (*response.KnowledgeExternalAPIListResp, error) {
	resp, err := knowledgeBase.SelectKnowledgeExternalAPIList(ctx.Request.Context(), &knowledgebase_service.KnowledgeExternalAPIListSelectReq{
		UserId: userId,
		OrgId:  orgId,
	})
	if err != nil {
		return nil, err
	}
	return buildKnowledgeExternalAPI(resp), nil
}

// CreateKnowledgeExternalAPI 创建外部知识库API
func CreateKnowledgeExternalAPI(ctx *gin.Context, userId, orgId string, r *request.CreateKnowledgeExternalAPIReq) (*response.CreateKnowledgeExternalAPIResp, error) {
	resp, err := knowledgeBase.CreateKnowledgeExternalAPI(ctx.Request.Context(), &knowledgebase_service.CreateKnowledgeExternalAPIReq{
		Name:        r.Name,
		Description: r.Description,
		BaseUrl:     r.BaseUrl,
		ApiKey:      r.ApiKey,
		UserId:      userId,
		OrgId:       orgId,
	})
	if err != nil {
		return nil, err
	}
	return &response.CreateKnowledgeExternalAPIResp{
		ExternalAPIId: resp.ExternalAPIId,
	}, nil
}

// UpdateKnowledgeExternalAPI 编辑外部知识库API
func UpdateKnowledgeExternalAPI(ctx *gin.Context, userId, orgId string, r *request.UpdateKnowledgeExternalAPIReq) error {
	existingExternalAPI, err := knowledgeBase.SelectKnowledgeExternalAPIInfo(ctx.Request.Context(), &knowledgebase_service.KnowledgeExternalAPIInfoSelectReq{
		UserId:        userId,
		OrgId:         orgId,
		ExternalAPIId: r.ExternalAPIId,
	})
	if err != nil {
		return err
	}
	if err := util.ValidateBriefUpdate(&r.Name, existingExternalAPI.Name, &r.Description, existingExternalAPI.Description, util.SubjectKnowledgeExternalAPI); err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, err.Error())
	}
	_, err = knowledgeBase.UpdateKnowledgeExternalAPI(ctx.Request.Context(), &knowledgebase_service.UpdateKnowledgeExternalAPIReq{
		ExternalAPIId: r.ExternalAPIId,
		Name:          r.Name,
		Description:   r.Description,
		BaseUrl:       r.BaseUrl,
		ApiKey:        r.ApiKey,
		UserId:        userId,
		OrgId:         orgId,
	})
	return err
}

// DeleteKnowledgeExternalAPI 删除外部知识库API
func DeleteKnowledgeExternalAPI(ctx *gin.Context, userId, orgId string, r *request.DeleteKnowledgeExternalAPIReq) error {
	_, err := knowledgeBase.DeleteKnowledgeExternalAPI(ctx.Request.Context(), &knowledgebase_service.DeleteKnowledgeExternalAPIReq{
		ExternalAPIId: r.ExternalAPIId,
		UserId:        userId,
		OrgId:         orgId,
	})
	return err
}

// GetKnowledgeExternalList 获取外部知识库列表
func GetKnowledgeExternalList(ctx *gin.Context, userId, orgId string, req *request.KnowledgeExternalListReq) (*response.KnowledgeExternalListResp, error) {
	resp, err := knowledgeBase.SelectKnowledgeExternalList(ctx.Request.Context(), &knowledgebase_service.KnowledgeExternalListSelectReq{
		UserId:        userId,
		OrgId:         orgId,
		ExternalAPIId: req.ExternalAPIId,
	})
	if err != nil {
		return nil, err
	}
	return buildKnowledgeExternal(resp), nil
}

// CreateKnowledgeExternal 创建外部知识库
func CreateKnowledgeExternal(ctx *gin.Context, userId, orgId string, r *request.CreateKnowledgeExternalReq) (*response.CreateKnowledgeExternalResp, error) {
	resp, err := knowledgeBase.CreateKnowledgeExternal(ctx.Request.Context(), &knowledgebase_service.CreateKnowledgeExternalReq{
		Name:                r.Name,
		Description:         r.Description,
		Provider:            r.ExternalSource,
		ExternalApiId:       r.ExternalAPIId,
		ExternalKnowledgeId: r.ExternalKnowledgeId,
		UserId:              userId,
		OrgId:               orgId,
	})
	if err != nil {
		return nil, err
	}
	return &response.CreateKnowledgeExternalResp{
		KnowledgeId: resp.KnowledgeId,
	}, nil
}

// UpdateKnowledgeExternal 编辑外部知识库
func UpdateKnowledgeExternal(ctx *gin.Context, userId, orgId string, r *request.UpdateKnowledgeExternalReq) error {
	existingKnowledge, err := knowledgeBase.SelectKnowledgeDetailById(ctx.Request.Context(), &knowledgebase_service.KnowledgeDetailSelectReq{
		UserId:      userId,
		OrgId:       orgId,
		KnowledgeId: r.KnowledgeId,
	})
	if err != nil {
		return err
	}
	if err := util.ValidateBriefUpdate(&r.Name, existingKnowledge.Name, &r.Description, existingKnowledge.Description, util.SubjectKnowledge); err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, err.Error())
	}
	_, err = knowledgeBase.UpdateKnowledgeExternal(ctx.Request.Context(), &knowledgebase_service.UpdateKnowledgeExternalReq{
		KnowledgeId:         r.KnowledgeId,
		Name:                r.Name,
		Description:         r.Description,
		Provider:            r.ExternalSource,
		ExternalApiId:       r.ExternalAPIId,
		ExternalKnowledgeId: r.ExternalKnowledgeId,
		UserId:              userId,
		OrgId:               orgId,
	})
	return err
}

// DeleteKnowledgeExternal 删除外部知识库
func DeleteKnowledgeExternal(ctx *gin.Context, userId, orgId string, r *request.DeleteKnowledgeExternalReq) error {
	_, err := knowledgeBase.DeleteKnowledgeExternal(ctx.Request.Context(), &knowledgebase_service.DeleteKnowledgeExternalReq{
		KnowledgeId: r.KnowledgeId,
		UserId:      userId,
		OrgId:       orgId,
	})
	return err
}
