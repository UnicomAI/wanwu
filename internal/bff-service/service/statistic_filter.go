package service

import (
	"slices"
	"sort"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	"github.com/gin-gonic/gin"
)

func GetStatisticOrgsSelect(ctx *gin.Context, userID, orgID string, isAdmin, isSystem bool) (*response.ListResult, error) {
	// 普通用户
	if !isAdmin {
		if orgID == "" {
			return &response.ListResult{List: []response.IDNameWithAvatar{}, Total: 0}, nil
		}
		org, err := iam.GetOrgInfo(ctx.Request.Context(), &iam_service.GetOrgInfoReq{OrgId: orgID})
		if err != nil {
			return nil, err
		}
		return &response.ListResult{List: []response.IDNameWithAvatar{{ID: org.OrgId, Name: org.Name}}, Total: 1}, nil
	}

	// 管理员
	resp, err := iam.GetOrgAndSubOrgSelectByUser(ctx.Request.Context(), &iam_service.GetOrgAndSubOrgSelectByUserReq{
		UserId: userID,
		OrgId:  orgID,
	})
	if err != nil {
		return nil, err
	}
	return &response.ListResult{List: toOrgsIDNameWithAvatar(ctx, resp.Orgs), Total: int64(len(resp.Orgs))}, nil
}

func GetStatisticUsersSelect(ctx *gin.Context, userID, orgID string, isAdmin bool) (*response.ListResult, error) {
	// 普通用户
	if !isAdmin {
		user, err := iam.GetUserInfo(ctx.Request.Context(), &iam_service.GetUserInfoReq{
			UserId: userID,
			OrgId:  orgID,
		})
		if err != nil {
			return nil, err
		}
		return &response.ListResult{List: []response.StatisticUserName{{UserID: user.UserId, UserName: user.UserName}}, Total: 1}, nil
	}

	// 管理员（系统管理员/组织管理员同逻辑）：批量接口一次拿全部可见组织用户（IAM 侧去重），
	// 避免 per-org GetUserList 的 N+1 RPC 与 IAM 侧逐用户组装 UserInfo 的放大开销。
	orgsResp, err := iam.GetOrgAndSubOrgSelectByUser(ctx.Request.Context(), &iam_service.GetOrgAndSubOrgSelectByUserReq{
		UserId: userID,
		OrgId:  orgID,
	})
	if err != nil {
		return nil, err
	}
	orgIds := make([]string, 0, len(orgsResp.Orgs))
	for _, org := range orgsResp.Orgs {
		if org != nil && org.Id != "" {
			orgIds = append(orgIds, org.Id)
		}
	}
	if len(orgIds) == 0 {
		return &response.ListResult{List: []response.StatisticUserName{}, Total: 0}, nil
	}
	usersResp, err := iam.GetUsersByOrgIDs(ctx.Request.Context(), &iam_service.GetUsersByOrgIDsReq{
		OrgIds: orgIds,
	})
	if err != nil {
		return nil, err
	}
	items := []response.StatisticUserName{}
	for _, user := range usersResp.Users {
		if user == nil || user.Id == "" {
			continue
		}
		items = append(items, response.StatisticUserName{
			UserID:   user.Id,
			UserName: user.Name,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].UserName < items[j].UserName
	})
	return &response.ListResult{List: items, Total: int64(len(items))}, nil
}

// GetStatisticModelSelect 模型 Tab 下拉（组织→用户→模型级联第三步）。
// filter 语义同统计接口；无 HasExpansion 时等同模型管理列表（仅 JWT 用户+组织下的模型）。
func GetStatisticModelSelect(ctx *gin.Context, modelType, userID, orgID string, filter *request.StatisticFilter, isAdmin, isSystem bool) (*response.ListResult, error) {
	scope, err := ResolveStatisticScope(ctx, *filter, userID, orgID, isAdmin, isSystem)
	if err != nil {
		return nil, err
	}

	if modelType == "" {
		modelType = mp.ModelTypeLLM
	}

	// 普通用户
	if !isAdmin {
		resp, err := model.ListModels(ctx.Request.Context(), &model_service.ListModelsReq{
			UserId:      userID,
			OrgId:       orgID,
			ModelType:   modelType,
			FilterScope: "private",
		})
		if err != nil {
			return nil, err
		}
		list, err := toModelInfos(ctx, resp.Models, &ModelInfoOptions{UserId: userID, OrgId: orgID})
		if err != nil {
			return nil, err
		}
		return &response.ListResult{List: list, Total: int64(len(list))}, nil
	}

	// 管理员
	resp, err := model.ListModelsInStatisticScope(ctx.Request.Context(), &model_service.ListModelsInStatisticScopeReq{
		OrgIds:      scope.OrgIds,
		UserIds:     scope.UserIds,
		ModelType:   modelType,
		FilterScope: "",
	})
	if err != nil {
		return nil, err
	}
	list, err := toModelInfos(ctx, resp.Models, &ModelInfoOptions{UserId: userID})
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].DisplayName < list[j].DisplayName
	})
	return &response.ListResult{List: list, Total: int64(len(list))}, nil
}

// statisticScope 统计查询解析结果，作为下游 gRPC 的 orgIds、userIds 入参。
type statisticScope struct {
	OrgIds  []string
	UserIds []string
}

// ResolveStatisticScope 将 filter 筛选解析为 orgIds、userIds。
//
// 系统管理员：ALL → 置空切片，下游 SQL 遇空切片跳过 WHERE 过滤，等价于查全量；
// 指定 ID → 原样传递，不做 IAM 展开调用。
// 组织管理员：ALL → 通过 IAM 展开为可见范围内的全部组织/用户；
// 指定 ID → 原样传递。
func ResolveStatisticScope(ctx *gin.Context, filter request.StatisticFilter, userID, orgID string, isAdmin, isSystem bool) (*statisticScope, error) {

	// 普通用户
	if !isAdmin {
		if len(filter.OrgIds) > 0 || len(filter.UserIds) > 0 {
			return nil, grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, "userIds and orgIds must be empty for non-admin users")
		}
		return &statisticScope{
			OrgIds:  []string{orgID},
			UserIds: []string{userID},
		}, nil
	}

	// 系统管理员：无需 IAM 展开可见范围，ALL 置空即全量，指定 ID 原样返回
	if isSystem {
		orgIds := filter.OrgIds
		userIds := filter.UserIds
		if slices.Contains(orgIds, request.StatisticFilterAll) {
			orgIds = []string{}
		}
		if slices.Contains(userIds, request.StatisticFilterAll) {
			userIds = []string{}
		}
		return &statisticScope{OrgIds: orgIds, UserIds: userIds}, nil
	}
	// 组织管理员：需要通过 IAM 解析可见范围
	var orgIds []string
	var err error
	if !slices.Contains(filter.OrgIds, request.StatisticFilterAll) {
		orgIds = filter.OrgIds
	} else {
		resp, err := iam.GetOrgAndSubOrgSelectByUser(ctx.Request.Context(), &iam_service.GetOrgAndSubOrgSelectByUserReq{
			UserId: userID,
			OrgId:  orgID,
		})
		if err != nil {
			return nil, err
		}
		for _, org := range resp.Orgs {
			if org != nil && org.Id != "" {
				orgIds = append(orgIds, org.Id)
			}
		}
	}

	var userIds []string
	if !slices.Contains(filter.UserIds, request.StatisticFilterAll) {
		userIds = filter.UserIds
	} else {
		userIds, err = collectStatisticUserIDsInOrgs(ctx, orgIds)
		if err != nil {
			return nil, err
		}
	}

	if len(orgIds) == 0 || len(userIds) == 0 {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, "筛选范围内无可用组织或用户")
	}
	return &statisticScope{OrgIds: orgIds, UserIds: userIds}, nil
}

// collectStatisticUserIDsInOrgs 批量取多个组织下的全部用户 ID（IAM 侧去重）。
func collectStatisticUserIDsInOrgs(ctx *gin.Context, orgIds []string) ([]string, error) {
	if len(orgIds) == 0 {
		return []string{}, nil
	}
	resp, err := iam.GetUsersByOrgIDs(ctx.Request.Context(), &iam_service.GetUsersByOrgIDsReq{
		OrgIds: orgIds,
	})
	if err != nil {
		return nil, err
	}
	userIds := make([]string, 0, len(resp.Users))
	for _, user := range resp.Users {
		if user == nil || user.Id == "" {
			continue
		}
		userIds = append(userIds, user.Id)
	}
	return userIds, nil
}
