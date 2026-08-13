package service

import (
	"fmt"

	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	knowledgebase_permission_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-permission-service"
	operate_service "github.com/UnicomAI/wanwu/api/proto/operate-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/log"
	safe_go_util "github.com/UnicomAI/wanwu/pkg/safe-go-util"
	"github.com/gin-gonic/gin"
)

const (
	SystemPermission int32 = 30
)

// SelectKnowledgeOrg 查询知识库组织
func SelectKnowledgeOrg(ctx *gin.Context, userId, orgId string, req *request.KnowledgeOrgSelectReq) (*response.KnowOrgInfoResp, error) {
	orgInfo, err := iam.GetFirstClassOrgAndSubs(ctx.Request.Context(), &iam_service.GetFirstClassOrgAndSubsReq{
		UserId: userId,
		OrgId:  orgId,
	})
	if err != nil {
		return nil, err
	}
	return buildKnowOrgInfo(orgInfo), nil
}

// SelectKnowledgePermissionUser 查询知识库有权限用户
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

// SelectKnowledgeNoPermissionUser 查询知识库没有权限用户
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
		idMap[userId] = true
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

// AddKnowledgeUser 增加知识库用户
func AddKnowledgeUser(ctx *gin.Context, userId, orgId string, req *request.KnowledgeUserAddReq) error {
	_, err := knowledgeBasePermission.AddKnowledgeUser(ctx.Request.Context(), &knowledgebase_permission_service.AddKnowledgeUserReq{
		KnowledgeId:       req.KnowledgeId,
		PermissionType:    int32(req.PermissionType),
		KnowledgeUserList: buildKnowledgeUserList(req.KnowledgeUserList),
		UserId:            userId,
		OrgId:             orgId,
	})
	if err != nil {
		return err
	}
	// 共享知识库：名单内每个人收「已共享给你」（best-effort）。
	// 请求体的 knowledgeUserList 天然就是 gained，无需差分。
	gained := make([]*operate_service.NoticeUserOrgPair, 0, len(req.KnowledgeUserList))
	for _, info := range req.KnowledgeUserList {
		gained = append(gained, &operate_service.NoticeUserOrgPair{UserId: info.UserId, OrgId: info.OrgId})
	}
	notifyKnowledgeDelta(ctx, userId, orgId, req.KnowledgeId,
		resolveKnowledgeName(ctx, userId, orgId, req.KnowledgeId),
		gained, nil, nil, noticeVariantShared, "",
		fmt.Sprintf("knowledge:%v:shared:%v", req.KnowledgeId, len(gained)))
	return nil
}

// EditKnowledgeUser 修改知识库用户
func EditKnowledgeUser(ctx *gin.Context, userId, orgId string, req *request.KnowledgeUserEditReq) error {
	_, err := knowledgeBasePermission.EditKnowledgeUser(ctx.Request.Context(), &knowledgebase_permission_service.EditKnowledgeUserReq{
		KnowledgeId:   req.KnowledgeId,
		KnowledgeUser: buildKnowledgeUser(req.KnowledgeUser),
		UserId:        userId,
		OrgId:         orgId,
	})
	if err != nil {
		return err
	}
	// 改权限级别：可见性没变、只是权限级别变——独立文案，非 online/offline
	changed := []*operate_service.NoticeUserOrgPair{{
		UserId: req.KnowledgeUser.UserId,
		OrgId:  req.KnowledgeUser.OrgId,
	}}
	notifyKnowledgeDelta(ctx, userId, orgId, req.KnowledgeId,
		resolveKnowledgeName(ctx, userId, orgId, req.KnowledgeId),
		nil, nil, changed,
		noticeVariantPermChanged, knowledgePermissionLabel(int32(req.KnowledgeUser.PermissionType)),
		fmt.Sprintf("knowledge:%v:perm:%v:%v", req.KnowledgeId, req.KnowledgeUser.UserId, req.KnowledgeUser.PermissionType))
	return nil
}

// DeleteKnowledgeUser 删除知识库用户
func DeleteKnowledgeUser(ctx *gin.Context, userId, orgId string, req *request.KnowledgeUserDeleteReq) error {
	// 取消共享：请求体只带 permissionId，必须在删除前反查被移除者的 (userId, orgId)
	lostUser := findKnowledgePermissionUser(ctx, userId, orgId, req.KnowledgeId, req.PermissionId)
	knowledgeName := resolveKnowledgeName(ctx, userId, orgId, req.KnowledgeId)

	_, err := knowledgeBasePermission.DeleteKnowledgeUser(ctx.Request.Context(), &knowledgebase_permission_service.DeleteKnowledgeUserReq{
		KnowledgeId:  req.KnowledgeId,
		PermissionId: req.PermissionId,
		UserId:       userId,
		OrgId:        orgId,
	})
	if err != nil {
		return err
	}
	if lostUser != nil {
		notifyKnowledgeDelta(ctx, userId, orgId, req.KnowledgeId, knowledgeName,
			nil, []*operate_service.NoticeUserOrgPair{lostUser}, nil,
			noticeVariantUnshared, "",
			fmt.Sprintf("knowledge:%v:unshared:%v", req.KnowledgeId, req.PermissionId))
	}
	return nil
}

// TransferKnowledgeAdminUser 转让知识库管理员权限
func TransferKnowledgeAdminUser(ctx *gin.Context, userId, orgId string, req *request.KnowledgeTransferUserAdminReq) error {
	// 转让后原管理员的权限记录会变，先把双方快照下来
	oldAdmin := findKnowledgePermissionUser(ctx, userId, orgId, req.KnowledgeId, req.PermissionId)
	knowledgeName := resolveKnowledgeName(ctx, userId, orgId, req.KnowledgeId)

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
	if err != nil {
		return err
	}
	// 转让管理员：双方可见性没变但权限级别变了，走"权限已变更"文案。
	// 新管理员必发；原管理员是操作者本人时会被消息域按二元组剔除。
	changed := []*operate_service.NoticeUserOrgPair{{
		UserId: req.KnowledgeUser.UserId,
		OrgId:  req.KnowledgeUser.OrgId,
	}}
	if oldAdmin != nil {
		changed = append(changed, oldAdmin)
	}
	notifyKnowledgeDelta(ctx, userId, orgId, req.KnowledgeId, knowledgeName,
		nil, nil, changed,
		noticeVariantPermChanged, knowledgePermissionLabel(SystemPermission),
		fmt.Sprintf("knowledge:%v:transfer:%v", req.KnowledgeId, req.KnowledgeUser.UserId))
	return nil
}

func buildKnowOrgInfo(orgInfo *iam_service.GetFirstClassOrgAndSubsResp) *response.KnowOrgInfoResp {
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

// 构建知识库有权限用户id map
func buildPermissionUserIdMap(permissionList *knowledgebase_permission_service.KnowledgeUserPermissionResp) map[string]bool {
	m := make(map[string]bool)
	for _, info := range permissionList.KnowledgeUserList {
		m[info.UserId] = true
	}
	return m
}

// 构建知识库有权限用户列表
func buildKnowledgePermissionUserList(ctx *gin.Context, knowledgeUserList []*knowledgebase_permission_service.KnowledgeUserInfo, userId, orgId string) []*response.KnowledgeUserInfo {
	if len(knowledgeUserList) > 0 {
		//并发请求userName 和orgName
		var userIdMap = make(map[string]bool)
		var orgIdMap = make(map[string]bool)
		for _, info := range knowledgeUserList {
			userIdMap[info.UserId] = true
			orgIdMap[info.OrgId] = true
		}
		userInfoMap, orgInfoMap := searchUserAndOrgInfo(ctx, userIdMap, orgIdMap)
		var retList []*response.KnowledgeUserInfo
		for _, info := range knowledgeUserList {
			userInfo := userInfoMap[info.UserId]
			orgInfo := orgInfoMap[info.OrgId]
			if userInfo == nil || orgInfo == nil {
				log.Errorf("user or org is nil, userId %s, orgId %s", info.UserId, info.OrgId)
				continue
			}
			retList = append(retList, &response.KnowledgeUserInfo{
				UserId:         info.UserId,
				UserName:       userInfo.Name,
				OrgId:          info.OrgId,
				OrgName:        orgInfo.Name,
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
	//是系统管理员，同时当前用户 是 此权限记录的用户
	return userInfo.UserId == userId && userInfo.OrgId == orgId && userInfo.PermissionType == SystemPermission
}

// 并发查询用户详情和组织详情
func searchUserAndOrgInfo(ctx *gin.Context, userIdMap, orgIdMap map[string]bool) (map[string]*iam_service.IDNameWithAvatar, map[string]*iam_service.IDFullName) {
	var userIdList, orgIdList []string
	for userId := range userIdMap {
		userIdList = append(userIdList, userId)
	}
	for orgId := range orgIdMap {
		orgIdList = append(orgIdList, orgId)
	}
	orgInfoMap := make(map[string]*iam_service.IDFullName)
	userInfoMap := make(map[string]*iam_service.IDNameWithAvatar)
	//并发查询
	safe_go_util.SageGoWaitGroup(searchUser(ctx, userIdList, userInfoMap), searchOrg(ctx, orgIdList, orgInfoMap))
	return userInfoMap, orgInfoMap
}

// 查询用户信息
func searchUser(ctx *gin.Context, userIdList []string, userInfoMap map[string]*iam_service.IDNameWithAvatar) func() {
	return func() {
		userInfoList, err := iam.GetUserSelectByUserIDs(ctx.Request.Context(), &iam_service.GetUserSelectByUserIDsReq{
			UserIds: userIdList,
		})
		if err != nil {
			return
		}
		for _, info := range userInfoList.Selects {
			userInfoMap[info.Id] = info
		}
	}
}

// 查询组织信息
func searchOrg(ctx *gin.Context, orgIdList []string, orgInfoMap map[string]*iam_service.IDFullName) func() {
	return func() {
		orgInfoList, err := iam.GetOrgByOrgIDs(ctx.Request.Context(), &iam_service.GetOrgByOrgIDsReq{
			OrgIds: orgIdList,
		})
		if err != nil {
			return
		}
		for _, info := range orgInfoList.Orgs {
			orgInfoMap[info.Id] = info
		}
	}
}
