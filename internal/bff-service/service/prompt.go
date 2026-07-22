package service

import (
	"context"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

type PromptBiz struct{}

func init() {
	InitBizService(&PromptBiz{})
}

func (*PromptBiz) BizType() string { return constant.BizModuleResourcePrompt }

func (*PromptBiz) SearchBizOwner(ctx context.Context, bizId string) (userId, orgId string, err error) {
	resp, err := assistant.CustomPromptGet(ctx, &assistant_service.CustomPromptGetReq{
		CustomPromptId: bizId,
		Identity:       &assistant_service.Identity{},
	})
	if err != nil {
		return "", "", err
	}
	if resp.Identity == nil {
		return "", "", nil
	}
	return resp.Identity.UserId, resp.Identity.OrgId, nil
}

func CreateCustomPrompt(ctx *gin.Context, userId, orgId string, req request.CustomPromptCreate) (*response.CustomPromptIDResp, error) {
	resp, err := assistant.CustomPromptCreate(ctx.Request.Context(), &assistant_service.CustomPromptCreateReq{
		AvatarPath: req.Avatar.Key,
		Name:       req.Name,
		Desc:       req.Desc,
		Prompt:     req.Prompt,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	return &response.CustomPromptIDResp{
		CustomPromptID: resp.CustomPromptId,
	}, nil
}

func GetCustomPrompt(ctx *gin.Context, userId, orgId, customPromptId string) (*response.CustomPrompt, error) {
	resp, err := assistant.CustomPromptGet(ctx.Request.Context(), &assistant_service.CustomPromptGetReq{
		CustomPromptId: customPromptId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	return toCustomPrompt(ctx, resp), nil
}

func DeleteCustomPrompt(ctx *gin.Context, userId, orgId string, req request.CustomPromptIDReq) error {
	_, err := assistant.CustomPromptDelete(ctx.Request.Context(), &assistant_service.CustomPromptDeleteReq{
		CustomPromptId: req.CustomPromptID,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func UpdateCustomPrompt(ctx *gin.Context, userId, orgId string, req request.UpdateCustomPrompt) error {
	existingPrompt, err := assistant.CustomPromptGet(ctx.Request.Context(), &assistant_service.CustomPromptGetReq{
		CustomPromptId: req.CustomPromptID,
		Identity:       &assistant_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return err
	}
	if err := util.ValidateBriefUpdate(&req.Name, existingPrompt.Name, &req.Desc, existingPrompt.Desc, util.SubjectPrompt); err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, err.Error())
	}
	_, err = assistant.CustomPromptUpdate(ctx.Request.Context(), &assistant_service.CustomPromptUpdateReq{
		CustomPromptId: req.CustomPromptID,
		AvatarPath:     req.Avatar.Key,
		Name:           req.Name,
		Desc:           req.Desc,
		Prompt:         req.Prompt,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func GetCustomPromptList(ctx *gin.Context, userId, orgId string, name string) (*response.ListResult, error) {
	resp, err := assistant.CustomPromptGetList(ctx.Request.Context(), &assistant_service.CustomPromptGetListReq{
		Name: name,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}

	customPromptList := make([]*response.CustomPrompt, 0, len(resp.CustomPromptInfos))
	for _, customPrompt := range resp.CustomPromptInfos {
		customPromptList = append(customPromptList, toCustomPrompt(ctx, customPrompt))
	}

	return &response.ListResult{
		List:  customPromptList,
		Total: resp.Total,
	}, nil
}

func CopyCustomPrompt(ctx *gin.Context, userId, orgId, customPromptId string) (*response.CustomPromptIDResp, error) {
	resp, err := assistant.CustomPromptCopy(ctx.Request.Context(), &assistant_service.CustomPromptCopyReq{
		CustomPromptId: customPromptId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	return &response.CustomPromptIDResp{
		CustomPromptID: resp.CustomPromptId,
	}, nil
}

// --- internal ---
func toCustomPrompt(ctx *gin.Context, resp *assistant_service.CustomPromptInfo) *response.CustomPrompt {
	return &response.CustomPrompt{
		CustomPromptIDResp: response.CustomPromptIDResp{
			CustomPromptID: resp.CustomPromptId,
		},
		Avatar:   cachePromptAvatar(ctx, resp.AvatarPath),
		Name:     resp.Name,
		Desc:     resp.Desc,
		Prompt:   resp.Prompt,
		UpdateAt: util.Time2Str(resp.UpdatedAt),
	}
}
