package service

import (
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/gin-gonic/gin"
)

func FillOrgIds(ctx *gin.Context, userId, orgId string, params *request.AdminUserSelect) error {
	if params.IsAllOrg {
		if isSystem(ctx) {
			params.OrgIdList = make([]string, 0)
			return nil
		}
		orgIdResp, err := iam.GetAdminOrgIDs(ctx, &iam_service.GetAdminOrgIDsReq{
			UserId: userId,
		})
		if err != nil {
			log.Errorf("get all org ids fail, err %v", err)
			return err
		}
		params.OrgIdList = orgIdResp.OrgIds
	}
	return nil
}

// fillOwnerList 填充用户和组织信息
func fillOwnerList[T response.OwnerInfoService](ctx *gin.Context, dataList []T) {
	if len(dataList) == 0 {
		return
	}
	var userIdMap = make(map[string]bool)
	var orgIdMap = make(map[string]bool)
	for _, item := range dataList {
		owner := item.GetOwnerInfo()
		userIdMap[owner.OwnerUserId] = true
		orgIdMap[owner.OwnerOrgId] = true
	}
	userInfoMap, orgInfoMap := searchUserAndOrgInfo(ctx, userIdMap, orgIdMap)
	for _, item := range dataList {
		owner := item.GetOwnerInfo()
		ownerInfo := buildOwnerInfo(owner.OwnerOrgId, owner.OwnerUserId, userInfoMap, orgInfoMap, false)
		if ownerInfo != nil {
			item.SetOwnerInfo(*ownerInfo)
		}
	}
}

// fillOwner 填充用户和组织信息
func fillOwner[T response.OwnerInfoService](ctx *gin.Context, data T) {
	owner := data.GetOwnerInfo()
	ownerInfo := searchOwnerInfo(ctx, owner.OwnerOrgId, owner.OwnerUserId, true)
	if ownerInfo != nil {
		data.SetOwnerInfo(*ownerInfo)
	}
}

// searchOwnerInfo 查询用户和组织信息
func searchOwnerInfo(ctx *gin.Context, ownerOrgId, ownerUserId string, fullName bool) *response.OwnerInfo {
	if ownerOrgId != "" && ownerUserId != "" {
		userInfoMap, orgInfoMap := searchUserAndOrgInfo(ctx, map[string]bool{ownerUserId: true}, map[string]bool{ownerOrgId: true})
		return buildOwnerInfo(ownerOrgId, ownerUserId, userInfoMap, orgInfoMap, fullName)
	}
	return nil
}

// buildOwnerInfo 构建用户和组织信息
func buildOwnerInfo(ownerOrgId, ownerUserId string, userInfoMap map[string]*iam_service.IDNameWithAvatar, orgInfoMap map[string]*iam_service.IDFullName, fullName bool) *response.OwnerInfo {
	var userName, orgName string
	if userInfo := userInfoMap[ownerUserId]; userInfo != nil {
		userName = userInfo.Name
	}
	if orgInfo := orgInfoMap[ownerOrgId]; orgInfo != nil {
		if fullName {
			orgName = orgInfo.FullName
		} else {
			orgName = orgInfo.Name
		}
	}
	if userName != "" && orgName != "" {
		return &response.OwnerInfo{
			OwnerOrgId:    ownerOrgId,
			OwnerOrgName:  orgName,
			OwnerUserId:   ownerUserId,
			OwnerUserName: userName,
		}
	}
	return nil
}

// 当前组织是否是内置顶级【系统】组织
func isSystem(ctx *gin.Context) bool {
	return ctx.GetBool(gin_util.IS_SYSTEM)
}
