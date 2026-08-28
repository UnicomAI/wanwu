package service

import (
	"time"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func CreateApiKey(ctx *gin.Context, userId, orgId string, req request.CreateAPIKeyRequest) (*response.APIKeyDetailResponse, error) {
	expiredAt, _ := util.Date2Time(req.ExpiredAt)
	keyInfo, err := app.CreateApiKey(ctx.Request.Context(), &app_service.CreateApiKeyReq{
		UserId:    userId,
		OrgId:     orgId,
		Name:      req.Name,
		Desc:      req.Desc,
		ExpiredAt: expiredAt,
	})
	if err != nil {
		return nil, err
	}
	return toApiKeyResponse(keyInfo, getUserNameById(ctx, userId)), nil
}

func DeleteApiKey(ctx *gin.Context, userId, orgId string, req request.DeleteAPIKeyRequest) error {
	_, err := app.DeleteApiKey(ctx.Request.Context(), &app_service.DeleteApiKeyReq{
		KeyId:  req.KeyID,
		UserId: userId,
		OrgId:  orgId,
	})
	if err != nil {
		return err
	}
	return nil
}

func ListApiKeys(ctx *gin.Context, userId, orgId string, pageNo, pageSize int32) (*response.PageResult, error) {
	keys, err := app.ListApiKeys(ctx.Request.Context(), &app_service.ListApiKeysReq{
		PageNo:   pageNo,
		PageSize: pageSize,
		UserIds:  []string{userId},
		OrgIds:   []string{orgId},
	})
	if err != nil {
		return nil, err
	}
	creatorName := getUserNameById(ctx, userId)
	var result []*response.APIKeyDetailResponse
	// 遍历keys返回给前端
	for _, key := range keys.Items {
		result = append(result, toApiKeyResponse(key, creatorName))
	}
	return &response.PageResult{
		List:     result,
		Total:    int64(keys.Total),
		PageNo:   int(pageNo),
		PageSize: int(pageSize),
	}, nil
}

func UpdateApiKey(ctx *gin.Context, userId, orgId string, req request.UpdateAPIKeyRequest) error {
	expiredAt, _ := util.Date2Time(req.ExpiredAt)
	_, err := app.UpdateApiKey(ctx.Request.Context(), &app_service.UpdateApiKeyReq{
		KeyId:     req.KeyID,
		Name:      req.Name,
		Desc:      req.Desc,
		ExpiredAt: expiredAt,
		UserId:    userId,
		OrgId:     orgId,
	})
	if err != nil {
		return err
	}
	return nil
}

func UpdateApiKeyStatus(ctx *gin.Context, req request.UpdateAPIKeyStatusRequest) error {
	_, err := app.UpdateApiKeyStatus(ctx.Request.Context(), &app_service.UpdateApiKeyStatusReq{
		KeyId:  req.KeyID,
		Status: req.Status,
	})
	if err != nil {
		return err
	}
	return nil
}

func GetApiKeyByKey(ctx *gin.Context, apiKey string) (*app_service.ApiKeyInfo, error) {
	return app.GetApiKeyByKey(ctx.Request.Context(), &app_service.GetApiKeyByKeyReq{ApiKey: apiKey})
}

// GetUserInfoByApiKey 通过api key获取用户信息（内部接口）
func GetUserInfoByApiKey(ctx *gin.Context, apiKey string) (*response.UserInfoByApiKey, error) {
	keyInfo, err := GetApiKeyByKey(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	if !keyInfo.Status {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "api key disabled")
	}
	if keyInfo.ExpiredAt != 0 && keyInfo.ExpiredAt < time.Now().UnixMilli() {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "api key expired")
	}
	ret, err := iam.GetUserSelectByUserIDs(ctx.Request.Context(), &iam_service.GetUserSelectByUserIDsReq{
		UserIds: []string{keyInfo.UserId},
	})
	if err != nil {
		return nil, err
	}
	name := ""
	if len(ret.Selects) > 0 {
		name = ret.Selects[0].Name
	}
	return &response.UserInfoByApiKey{
		UserID:   keyInfo.UserId,
		OrgID:    keyInfo.OrgId,
		Username: name,
	}, nil
}

// --- internal ---
func getUserNameById(ctx *gin.Context, userId string) string {
	ret, err := iam.GetUserSelectByUserIDs(ctx.Request.Context(), &iam_service.GetUserSelectByUserIDsReq{
		UserIds: []string{userId},
	})
	if err != nil || len(ret.Selects) == 0 {
		return ""
	}
	return ret.Selects[0].Name
}

func toApiKeyResponse(keyInfo *app_service.ApiKeyInfo, creatorName string) *response.APIKeyDetailResponse {
	expiredAtStr := ""
	if keyInfo.ExpiredAt != 0 {
		expiredAtStr = util.Time2Date(keyInfo.ExpiredAt)
	}
	return &response.APIKeyDetailResponse{
		KeyID:     keyInfo.KeyId,
		Key:       keyInfo.Key,
		Creator:   creatorName,
		Name:      keyInfo.Name,
		Desc:      keyInfo.Desc,
		ExpiredAt: expiredAtStr,
		CreatedAt: util.Time2Str(keyInfo.CreatedAt),
		Status:    keyInfo.Status,
	}
}
