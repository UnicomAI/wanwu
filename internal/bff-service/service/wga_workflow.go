package service

import (
	"fmt"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	wga_option "github.com/UnicomAI/wanwu/pkg/wga/wga-option"
	"github.com/gin-gonic/gin"
)

// --- internal wga workflow ---

// checkWgaWorkflowConfig 校验wga Workflow配置（用于更新配置）
func checkWgaWorkflowConfig(ctx *gin.Context, userId, orgId string, workflowList []*assistant_service.WgaConfigWorkflow) error {
	if len(workflowList) == 0 {
		return nil
	}

	workflowIds := make([]string, 0, len(workflowList))
	for _, w := range workflowList {
		workflowIds = append(workflowIds, w.WorkflowId)
	}

	validIds, err := getValidWorkflowIds(ctx, workflowIds)
	if err != nil {
		return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, "workflow not found")
	}

	for _, w := range workflowList {
		if !validIds[w.WorkflowId] {
			return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("workflow not found: %s", w.WorkflowId))
		}
	}
	return nil
}

func buildWgaWorkflowOptions(ctx *gin.Context, userId, orgId string, workflowList []*assistant_service.WgaConfigWorkflow) ([]wga_option.Option, error) {
	if len(workflowList) == 0 {
		return nil, nil
	}

	var workflowIDs []string
	for _, wf := range workflowList {
		workflowIDs = append(workflowIDs, wf.WorkflowId)
	}
	workflowSchemas, err := GetWorkflowSchemas(ctx, workflowIDs)
	if err != nil {
		return nil, err
	}
	var opts []wga_option.Option
	for _, schema := range workflowSchemas {
		opts = append(opts, wga_option.WithExtraTool(wga_option.ExtraTool{OpenAPI3Schema: schema}))
	}
	return opts, nil
}

// getValidWorkflowIds 批量获取有效的Workflow ID映射
func getValidWorkflowIds(ctx *gin.Context, workflowIds []string) (map[string]bool, error) {
	if len(workflowIds) == 0 {
		return make(map[string]bool), nil
	}
	workflowResp, err := ListWorkflowByIDs(ctx, "", workflowIds)
	if err != nil {
		return nil, err
	}
	validIds := make(map[string]bool)
	for _, w := range workflowResp.Workflows {
		validIds[w.WorkflowId] = true
	}
	return validIds, nil
}
