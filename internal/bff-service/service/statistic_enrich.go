package service

import (
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	"github.com/gin-gonic/gin"
)

func buildStatisticUserNameMap(ctx *gin.Context, userIDs []string) (map[string]string, error) {
	if len(userIDs) == 0 {
		return map[string]string{}, nil
	}
	userIDSet := make(map[string]struct{}, len(userIDs))
	for _, id := range userIDs {
		if id != "" {
			userIDSet[id] = struct{}{}
		}
	}
	if len(userIDSet) == 0 {
		return map[string]string{}, nil
	}
	deduped := make([]string, 0, len(userIDSet))
	for id := range userIDSet {
		deduped = append(deduped, id)
	}
	resp, err := iam.GetUserSelectByUserIDs(ctx.Request.Context(), &iam_service.GetUserSelectByUserIDsReq{
		UserIds: deduped,
	})
	if err != nil {
		return nil, err
	}
	userNameMap := make(map[string]string, len(resp.Selects))
	for _, sel := range resp.Selects {
		userNameMap[sel.Id] = sel.Name
	}
	return userNameMap, nil
}

func buildStatisticOrgNameMap(ctx *gin.Context, orgIDs []string) (map[string]string, error) {
	if len(orgIDs) == 0 {
		return map[string]string{}, nil
	}
	orgResp, err := iam.GetOrgByOrgIDs(ctx.Request.Context(), &iam_service.GetOrgByOrgIDsReq{OrgIds: orgIDs})
	if err != nil {
		return nil, err
	}
	orgNameMap := make(map[string]string)
	if orgResp != nil && orgResp.Orgs != nil {
		for _, org := range orgResp.Orgs {
			orgNameMap[org.Id] = org.Name
		}
	}
	return orgNameMap, nil
}

func buildStatisticOrgUserNameMaps(ctx *gin.Context, orgIDs []string, userIDs []string) (map[string]string, map[string]string, error) {
	orgNameMap, err := buildStatisticOrgNameMap(ctx, orgIDs)
	if err != nil {
		return nil, nil, err
	}
	userNameMap, err := buildStatisticUserNameMap(ctx, userIDs)
	if err != nil {
		return nil, nil, err
	}
	return orgNameMap, userNameMap, nil
}
