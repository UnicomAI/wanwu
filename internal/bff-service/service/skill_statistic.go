package service

import (
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/gin-gonic/gin"
)

// incrementCustomSkillDownload 递增 custom skill 下载计数
func incrementCustomSkillDownload(ctx *gin.Context, skillId string) {
	_, err := mcp.IncrementCustomSkillDownloadCount(ctx.Request.Context(), &mcp_service.IncrementCustomSkillDownloadCountReq{
		SkillId: skillId,
	})
	if err != nil {
		log.Errorf("incrementCustomSkillDownload failed: skillId=%s err=%v", skillId, err)
	}
}

// incrementCustomSkillAcquired 递增 custom skill 添加计数
func incrementCustomSkillAcquired(ctx *gin.Context, skillId string) {
	_, err := mcp.IncrementCustomSkillAcquiredCount(ctx.Request.Context(), &mcp_service.IncrementCustomSkillAcquiredCountReq{
		SkillId: skillId,
	})
	if err != nil {
		log.Errorf("incrementCustomSkillAcquired failed: skillId=%s err=%v", skillId, err)
	}
}

// incrementBuiltinSkillDownload 递增内置 skill 下载计数
func incrementBuiltinSkillDownload(ctx *gin.Context, skillId string) {
	_, err := mcp.IncrementBuiltinSkillDownloadCount(ctx.Request.Context(), &mcp_service.IncrementBuiltinSkillDownloadCountReq{
		SkillId: skillId,
	})
	if err != nil {
		log.Errorf("incrementBuiltinSkillDownload failed: skillId=%s err=%v", skillId, err)
	}
}

// getBuiltinSkillDownloadCounts 批量获取内置 skill 下载计数
func getBuiltinSkillDownloadCounts(ctx *gin.Context, skillIds []string) map[string]int32 {
	if len(skillIds) == 0 {
		return make(map[string]int32)
	}
	resp, err := mcp.GetBuiltinSkillDownloadCounts(ctx.Request.Context(), &mcp_service.GetBuiltinSkillDownloadCountsReq{
		SkillIds: skillIds,
	})
	if err != nil {
		log.Errorf("getBuiltinSkillDownloadCounts failed: err=%v", err)
		return make(map[string]int32)
	}
	if resp == nil || resp.CountMap == nil {
		return make(map[string]int32)
	}
	return resp.CountMap
}
