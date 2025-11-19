package service

import (
	"sync"

	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	knowledgebase_permission_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-permission-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

const (
	SystemPermission int32 = 30
)

// SelectKnowledgeOrg 查询知识库组织 [EN] SelectKnowledgeOrg Query knowledge base organization
func SelectKnowledgeOrg(ctx *gin.Context, userId, orgId string, req *request.KnowledgeOrgSelectReq) (*response.KnowOrgInfoResp, error) {
	orgInfo, err := iam.GetOrgAndSubOrgSelectByUser(ctx.Request.Context(), &iam_service.GetOrgAndSubOrgSelectByUserReq{
		UserId: userId,
		OrgId:  orgId,
	})
	if err != nil {
		return nil, err
	}
	return buildKnowOrgInfo(orgInfo), nil
}

// SelectKnowledgePermissionUser 查询知识库有权限用户 [EN] SelectKnowledgePermissionUser Query knowledge base authorized users
func SelectKnowledgePermissionUser(ctx *gin.Context, userId, orgId string, req *request.KnowledgeUserSelectReq) (*response.KnowledgeUserPermissionResp, error) {
	dataListResp, err := knowledgeBasePermission.SelectKnowledgeUserPermission(ctx.Request.Context(), &knowledgebase_permission_service.KnowledgeUserPermissionReq{
		KnowledgeId: req.KnowledgeId,
		UserId:      userId,
		OrgId:       orgId,
	})
	if err != nil {
		return nil, err
	}

	return &response.KnowledgeUserPermissionResp{
		KnowledgeUserInfoList: buildKnowledgePermissionUserList(ctx, dataListResp.KnowledgeUserList, userId, orgId),
	}, err
}

// SelectKnowledgeNoPermissionUser 查询知识库没有权限用户 [EN] SelectKnowledgeNoPermissionUser queries the knowledge base for users without permission
func SelectKnowledgeNoPermissionUser(ctx *gin.Context, userId, orgId string, req *request.KnowledgeUserNoPermitSelectReq) (*response.KnowOrgUserInfoResp, error) {
	list, err := iam.GetUserList(ctx.Request.Context(), &iam_service.GetUserListReq{
		OrgId:    req.OrgId,
		PageNo:   1,
		PageSize: 500,
	})
	if err != nil {
		return nil, err
	}
	idMap := make(map[string]bool)
	if req.Transfer {
		for _, user := range list.Users {
			if user.UserId == userId && req.OrgId == orgId {
				continue
			}
			idMap[user.UserId] = true
		}
	} else {
		permissionList, err := knowledgeBasePermission.SelectKnowledgeUserPermission(ctx.Request.Context(), &knowledgebase_permission_service.KnowledgeUserPermissionReq{
			KnowledgeId: req.KnowledgeId,
			UserId:      userId,
			OrgId:       orgId,
		})
		if err != nil {
			return nil, err
		}
		idMap = buildPermissionUserIdMap(permissionList)
	}

	return &response.KnowOrgUserInfoResp{
		OrgId:        req.OrgId,
		OrgName:      "",
		UserInfoList: buildNoPermissionKnowledgeUserList(idMap, list),
	}, nil
}

func CheckKnowledgeUserPermission(ctx *gin.Context, userId, orgId, knowledgeId string, permissionType int32) error {
	_, err := knowledgeBasePermission.CheckKnowledgeUserPermission(ctx.Request.Context(), &knowledgebase_permission_service.CheckKnowledgeUserPermissionReq{
		KnowledgeId:    knowledgeId,
		PermissionType: permissionType,
		UserId:         userId,
		OrgId:          orgId,
	})
	return err
}

// AddKnowledgeUser 增加知识库用户 [EN] AddKnowledgeUser adds a knowledge base user
func AddKnowledgeUser(ctx *gin.Context, userId, orgId string, req *request.KnowledgeUserAddReq) error {
	_, err := knowledgeBasePermission.AddKnowledgeUser(ctx.Request.Context(), &knowledgebase_permission_service.AddKnowledgeUserReq{
		KnowledgeId:       req.KnowledgeId,
		PermissionType:    int32(req.PermissionType),
		KnowledgeUserList: buildKnowledgeUserList(req.KnowledgeUserList),
		UserId:            userId,
		OrgId:             orgId,
	})
	return err
}

// EditKnowledgeUser 修改知识库用户 [EN] EditKnowledgeUser Modify knowledge base user
func EditKnowledgeUser(ctx *gin.Context, userId, orgId string, req *request.KnowledgeUserEditReq) error {
	_, err := knowledgeBasePermission.EditKnowledgeUser(ctx.Request.Context(), &knowledgebase_permission_service.EditKnowledgeUserReq{
		KnowledgeId:   req.KnowledgeId,
		KnowledgeUser: buildKnowledgeUser(req.KnowledgeUser),
		UserId:        userId,
		OrgId:         orgId,
	})
	return err
}

// DeleteKnowledgeUser 删除知识库用户 [EN] DeleteKnowledgeUser Delete knowledge base user
func DeleteKnowledgeUser(ctx *gin.Context, userId, orgId string, req *request.KnowledgeUserDeleteReq) error {
	_, err := knowledgeBasePermission.DeleteKnowledgeUser(ctx.Request.Context(), &knowledgebase_permission_service.DeleteKnowledgeUserReq{
		KnowledgeId:  req.KnowledgeId,
		PermissionId: req.PermissionId,
		UserId:       userId,
		OrgId:        orgId,
	})
	return err
}

// TransferKnowledgeAdminUser 转让知识库管理员权限 [EN] TransferKnowledgeAdminUser transfers knowledge base administrator permissions
func TransferKnowledgeAdminUser(ctx *gin.Context, userId, orgId string, req *request.KnowledgeTransferUserAdminReq) error {
	_, err := knowledgeBasePermission.TransferKnowledgeAdminUser(ctx.Request.Context(), &knowledgebase_permission_service.TransferKnowledgeAdminUserReq{
		KnowledgeId:  req.KnowledgeId,
		PermissionId: req.PermissionId,
		KnowledgeUser: &knowledgebase_permission_service.KnowledgeUserInfo{
			UserId: req.KnowledgeUser.UserId,
			OrgId:  req.KnowledgeUser.OrgId,
		},
		UserId: userId,
		OrgId:  orgId,
	})
	return err
}

func buildKnowOrgInfo(orgInfo *iam_service.GetOrgAndSubOrgSelectByUserResp) *response.KnowOrgInfoResp {
	var retList []*response.KnowOrgInfo
	for _, org := range orgInfo.Orgs {
		if org.Id == config.TopOrgID {
			continue
		}
		retList = append(retList, &response.KnowOrgInfo{
			OrgId:   org.Id,
			OrgName: org.Name,
		})
	}
	return &response.KnowOrgInfoResp{
		KnowOrgInfoList: retList,
	}
}

func buildKnowledgeUserList(knowledgeUserList []*request.KnowledgeUserInfo) []*knowledgebase_permission_service.KnowledgeUserInfo {
	var list []*knowledgebase_permission_service.KnowledgeUserInfo
	for _, info := range knowledgeUserList {
		list = append(list, buildKnowledgeUser(info))
	}
	return list
}

func buildKnowledgeUser(knowledgeUser *request.KnowledgeUserInfo) *knowledgebase_permission_service.KnowledgeUserInfo {
	return &knowledgebase_permission_service.KnowledgeUserInfo{
		UserId:         knowledgeUser.UserId,
		OrgId:          knowledgeUser.OrgId,
		PermissionType: int32(knowledgeUser.PermissionType),
		PermissionId:   knowledgeUser.PermissionId,
	}
}

func buildNoPermissionKnowledgeUserList(permissionUserMap map[string]bool, userList *iam_service.GetUserListResp) []*response.KnowUserInfo {
	var list []*response.KnowUserInfo
	for _, info := range userList.Users {
		if permissionUserMap[info.UserId] {
			continue
		}
		list = append(list, &response.KnowUserInfo{
			UserId:   info.UserId,
			UserName: info.UserName,
		})
	}
	return list
}

// 构建知识库有权限用户id map [EN] Build a knowledge base authorized user id map
func buildPermissionUserIdMap(permissionList *knowledgebase_permission_service.KnowledgeUserPermissionResp) map[string]bool {
	m := make(map[string]bool)
	for _, info := range permissionList.KnowledgeUserList {
		m[info.UserId] = true
	}
	return m
}

// 构建知识库有权限用户列表 [EN] Build a list of authorized users in the knowledge base
func buildKnowledgePermissionUserList(ctx *gin.Context, knowledgeUserList []*knowledgebase_permission_service.KnowledgeUserInfo, userId, orgId string) []*response.KnowledgeUserInfo {
	if len(knowledgeUserList) > 0 {
		//并发请求userName 和orgName [EN] Concurrent requests for userName and orgName
		var userIdMap = make(map[string]bool)
		var orgIdMap = make(map[string]bool)
		for _, info := range knowledgeUserList {
			userIdMap[info.UserId] = true
			orgIdMap[info.OrgId] = true
		}
		userInfoMap, orgInfoMap := searchUserAndOrgInfo(ctx, userIdMap, orgIdMap)
		var retList []*response.KnowledgeUserInfo
		for _, info := range knowledgeUserList {
			retList = append(retList, &response.KnowledgeUserInfo{
				UserId:         info.UserId,
				UserName:       userInfoMap[info.UserId].Name,
				OrgId:          info.OrgId,
				OrgName:        orgInfoMap[info.OrgId].Name,
				PermissionType: int(info.PermissionType),
				PermissionId:   info.PermissionId,
				Transfer:       buildUserTransfer(info, userId, orgId),
			})
		}
		return retList
	}
	return make([]*response.KnowledgeUserInfo, 0)
}

func buildUserTransfer(userInfo *knowledgebase_permission_service.KnowledgeUserInfo, userId, orgId string) bool {
	//是系统管理员，同时当前用户 是 此权限记录的用户 [EN] Is the system administrator, and the current user is a user recorded with this permission
	return userInfo.UserId == userId && userInfo.OrgId == orgId && userInfo.PermissionType == SystemPermission
}

// 并发查询用户详情和组织详情 [EN] Query user details and organization details concurrently
func searchUserAndOrgInfo(ctx *gin.Context, userIdMap, orgIdMap map[string]bool) (map[string]*iam_service.IDName, map[string]*iam_service.IDName) {
	var userIdList, orgIdList []string
	for userId := range userIdMap {
		userIdList = append(userIdList, userId)
	}
	for orgId := range orgIdMap {
		orgIdList = append(orgIdList, orgId)
	}
	var wg = &sync.WaitGroup{}
	wg.Add(2)
	orgInfoMap := make(map[string]*iam_service.IDName)
	userInfoMap := make(map[string]*iam_service.IDName)
	//查询user详情信息 [EN] Query user details
	go func() {
		defer func() {
			wg.Done()
		}()
		defer util.PrintPanicStack()

		userInfoList, err := iam.GetUserSelectByUserIDs(ctx, &iam_service.GetUserSelectByUserIDsReq{
			UserIds: userIdList,
		})
		if err != nil {
			return
		}
		for _, info := range userInfoList.Selects {
			userInfoMap[info.Id] = info
		}
	}()
	//查询组织详情信息 [EN] Query organization details
	go func() {
		defer func() {
			wg.Done()
		}()
		defer util.PrintPanicStack()

		userInfoList, err := iam.GetOrgByOrgIDs(ctx, &iam_service.GetOrgByOrgIDsReq{
			OrgIds: orgIdList,
		})
		if err != nil {
			return
		}
		for _, info := range userInfoList.Orgs {
			orgInfoMap[info.Id] = info
		}
	}()
	wg.Wait()
	return userInfoMap, orgInfoMap
}
