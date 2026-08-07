package service

import (
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/gin-gonic/gin"
)

// statisticOverviewProtoItem 各维度概览 proto（V1/V2 model/app/apiKey）共用 Getter。
type statisticOverviewProtoItem interface {
	GetValue() float32
	GetPeriodOverPeriod() float32
}

// convertStatisticOverviewItem 透传 app-service 已按两位小数取整后的概览项。
// proto nil 接收者的 Get* 返回零值，调用方无需再包一层类型专用 convert。
func convertStatisticOverviewItem(item statisticOverviewProtoItem) response.StatisticOverviewItem {
	if item == nil {
		return response.StatisticOverviewItem{}
	}
	return response.StatisticOverviewItem{
		Value:            item.GetValue(),
		PeriodOverPeriod: item.GetPeriodOverPeriod(),
	}
}

// convertStatisticV2CallOverview 应用 / API Key V2 Overview 共用透传（字段集相同）。
func convertStatisticV2CallOverview(
	callCount, callFailure, dailyAvgCallCount, dailyAvgCallFailure,
	dailyAvgStreamCount, dailyAvgNonStreamCount,
	avgFirstTokenLatency, avgCosts, streamCount, nonStreamCount statisticOverviewProtoItem,
) *response.StatisticV2CallOverview {
	return &response.StatisticV2CallOverview{
		CallCount:              convertStatisticOverviewItem(callCount),
		CallFailure:            convertStatisticOverviewItem(callFailure),
		DailyAvgCallCount:      convertStatisticOverviewItem(dailyAvgCallCount),
		DailyAvgCallFailure:    convertStatisticOverviewItem(dailyAvgCallFailure),
		DailyAvgStreamCount:    convertStatisticOverviewItem(dailyAvgStreamCount),
		DailyAvgNonStreamCount: convertStatisticOverviewItem(dailyAvgNonStreamCount),
		AvgFirstTokenLatency:   convertStatisticOverviewItem(avgFirstTokenLatency),
		AvgCosts:               convertStatisticOverviewItem(avgCosts),
		StreamCount:            convertStatisticOverviewItem(streamCount),
		NonStreamCount:         convertStatisticOverviewItem(nonStreamCount),
	}
}

// buildStatisticUserMaps 批量拉用户名；withAvatar=true 时一并返回头像（仅 chart.rank 需要），
// 否则头像 map 为 nil，避免 list/record 做无谓的头像缓存计算。
func buildStatisticUserMaps(ctx *gin.Context, userIDs []string, withAvatar bool) (map[string]string, map[string]request.Avatar, error) {
	nameMap := map[string]string{}
	var avatarMap map[string]request.Avatar
	if withAvatar {
		avatarMap = map[string]request.Avatar{}
	}
	resp, err := iam.GetUserSelectByUserIDs(ctx.Request.Context(), &iam_service.GetUserSelectByUserIDsReq{
		UserIds: userIDs,
	})
	if err != nil {
		return nil, nil, err
	}
	for _, sel := range resp.Selects {
		nameMap[sel.Id] = sel.Name
		if withAvatar {
			avatarMap[sel.Id] = cacheUserAvatar(sel.GetAvatarPath())
		}
	}
	return nameMap, avatarMap, nil
}

// buildStatisticOrgMaps 批量拉组织名；withAvatar=true 时一并返回头像（仅 chart.rank 需要）。
func buildStatisticOrgMaps(ctx *gin.Context, orgIDs []string, withAvatar bool) (map[string]string, map[string]request.Avatar, error) {
	nameMap := map[string]string{}
	var avatarMap map[string]request.Avatar
	if withAvatar {
		avatarMap = map[string]request.Avatar{}
	}
	resp, err := iam.GetOrgByOrgIDs(ctx.Request.Context(), &iam_service.GetOrgByOrgIDsReq{OrgIds: orgIDs})
	if err != nil {
		return nil, nil, err
	}
	for _, org := range resp.GetOrgs() {
		nameMap[org.Id] = org.Name
		if withAvatar {
			avatarMap[org.Id] = cacheOrgAvatar(org.GetAvatarPath())
		}
	}
	return nameMap, avatarMap, nil
}

// buildStatisticOrgUserNameMaps 批量拉组织名与用户名（list/record 用，不含头像）。
func buildStatisticOrgUserNameMaps(ctx *gin.Context, orgIDs []string, userIDs []string) (map[string]string, map[string]string, error) {
	orgNameMap, _, err := buildStatisticOrgMaps(ctx, orgIDs, false)
	if err != nil {
		return nil, nil, err
	}
	userNameMap, _, err := buildStatisticUserMaps(ctx, userIDs, false)
	if err != nil {
		return nil, nil, err
	}
	return orgNameMap, userNameMap, nil
}

func buildUserBriefInfo(userId, orgId string, userNameMap, orgNameMap map[string]string, userAvatarMap map[string]request.Avatar) response.UserBriefInfo {
	avatar := request.Avatar{}
	if userAvatarMap != nil {
		avatar = userAvatarMap[userId]
	}
	return response.UserBriefInfo{
		UserId:     userId,
		UserName:   userNameMap[userId],
		UserAvatar: avatar,
		OrgId:      orgId,
		OrgName:    orgNameMap[orgId],
	}
}
