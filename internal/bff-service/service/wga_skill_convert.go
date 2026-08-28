package service

import (
	"fmt"
	"path/filepath"
	"strings"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func ConvertGeneralAgentSkillConversation(ctx *gin.Context, userId, orgId string, req request.ConvertGeneralAgentSkillConversationReq) (*response.ConvertGeneralAgentSkillConversationResp, error) {
	if err := checkModelConfig(ctx, req.ModelConfig); err != nil {
		return nil, err
	}
	modelConfigString, err := req.ModelConfig.ConfigString()
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, err.Error())
	}

	sourceType := strings.TrimSpace(strings.ToLower(req.Type))

	// 检查应用是否已发布
	switch sourceType {
	case constant.AppTypeAgent, constant.AppTypeWorkflow, constant.AppTypeRag:
		_, err := app.GetAppInfo(ctx.Request.Context(), &app_service.GetAppInfoReq{
			AppId:   req.ID,
			AppType: sourceType,
		})
		if err != nil {
			return nil, grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("%s(%s) not published", sourceType, req.ID))
		}
	}

	title := generalAgentSkillConvertTitle(ctx, sourceType)
	previewID := util.GenUUID()

	threadResp, err := assistant.WgaConversationCreate(ctx.Request.Context(), &assistant_service.WgaConversationCreateReq{
		Prompt: title,
		ModelConfig: &common.AppModelConfig{
			ModelId:   req.ModelConfig.ModelId,
			Provider:  req.ModelConfig.Provider,
			Model:     req.ModelConfig.Model,
			ModelType: req.ModelConfig.ModelType,
			Config:    modelConfigString,
		},
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}

	customSkillResp, err := mcp.CustomSkillCreate(ctx.Request.Context(), &mcp_service.CustomSkillCreateReq{
		Name:            title,
		Author:          getUserNameById(ctx, userId),
		WgaThreadId:     threadResp.ThreadId,
		PreviewThreadId: previewID,
		Identity:        &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		rollbackImportedSkillConversation(ctx, userId, orgId, threadResp.ThreadId, "")
		return nil, err
	}
	customSkillID := customSkillResp.SkillId

	outputDir, err := prepareGeneralAgentSkillConvertOutputDir(customSkillID)
	if err != nil {
		rollbackImportedSkillConversation(ctx, userId, orgId, threadResp.ThreadId, customSkillID)
		return nil, err
	}
	if err := generateGeneralAgentSkillFromSource(ctx, sourceType, req.ID, outputDir); err != nil {
		rollbackImportedSkillConversation(ctx, userId, orgId, threadResp.ThreadId, customSkillID)
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("convert skill err: %v", err))
	}

	fm, err := readImportedSkillFrontMatter(outputDir)
	if err != nil {
		rollbackImportedSkillConversation(ctx, userId, orgId, threadResp.ThreadId, customSkillID)
		return nil, grpc_util.ErrorStatus(errs.Code_BFFSkillParse, err.Error())
	}
	if _, err = mcp.UpdateCustomSkillBasicMeta(ctx.Request.Context(), &mcp_service.UpdateCustomSkillBasicMetaReq{
		SkillId: customSkillID,
		Name:    fm.Name,
		Desc:    fm.Description,
	}); err != nil {
		rollbackImportedSkillConversation(ctx, userId, orgId, threadResp.ThreadId, customSkillID)
		return nil, err
	}

	// 首次产物自动提交（从未提交时）
	commitSkillWorkspaceIfFirstCommit(customSkillID, "skill convert")

	return &response.ConvertGeneralAgentSkillConversationResp{
		CustomSkillID: customSkillID,
		ThreadID:      threadResp.ThreadId,
		PreviewID:     previewID,
	}, nil
}

func prepareGeneralAgentSkillConvertOutputDir(customSkillID string) (string, error) {
	store, err := NewGeneralAgentSkillWorkspaceStore(customSkillID)
	if err != nil {
		return "", err
	}
	outputDir := filepath.Join(GetWgaWorkspaceThreadDir(store), generalAgentWorkspaceSkillDirName)
	if err := util.RecreateDir(outputDir); err != nil {
		return "", grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("prepare skill dir err: %v", err))
	}
	return outputDir, nil
}

func generateGeneralAgentSkillFromSource(ctx *gin.Context, sourceType, id, outputDir string) error {
	switch sourceType {
	case "mcp":
		return GenerateSkillFromMCP(ctx, id, outputDir)
	case "tool":
		return GenerateSkillFromCustomTool(ctx, id, outputDir)
	case "agent":
		return GenerateSkillFromAgent(ctx, id, outputDir)
	case "workflow":
		return GenerateSkillFromWorkflow(ctx, id, outputDir)
	case "rag":
		return GenerateSkillFromRAG(ctx, id, outputDir)
	default:
		return fmt.Errorf("unsupported type: %s", sourceType)
	}
}

func generalAgentSkillConvertTitle(ctx *gin.Context, sourceType string) string {
	switch sourceType {
	case "mcp":
		return gin_util.I18nKey(ctx, "wga_skill_convert_mcp_title")
	case "tool":
		return gin_util.I18nKey(ctx, "wga_skill_convert_tool_title")
	case "agent":
		return gin_util.I18nKey(ctx, "wga_skill_convert_agent_title")
	case "workflow":
		return gin_util.I18nKey(ctx, "wga_skill_convert_workflow_title")
	case "rag":
		return gin_util.I18nKey(ctx, "wga_skill_convert_rag_title")
	default:
		return gin_util.I18nKey(ctx, "wga_skill_convert_default_title")
	}
}
