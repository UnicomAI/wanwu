package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/middleware"
	"github.com/UnicomAI/wanwu/pkg/constant"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerModel(apiV1 *gin.RouterGroup) {
	mid.Sub("model.model_management").Reg(apiV1, "/model", http.MethodPost, v1.ImportModel, "模型导入")
	mid.Sub("model.model_management").Reg(apiV1, "/model", http.MethodPut, v1.UpdateModel, "导入模型更新")
	mid.Sub("model.model_management").Reg(apiV1, "/model", http.MethodDelete, v1.DeleteModel, "导入模型删除")
	mid.Sub("model.model_management").Reg(apiV1, "/model", http.MethodGet, v1.GetModel, "查询单个模型", middleware.AuthModelByModelId([]string{"modelId"}))
	mid.Sub("model.model_management").Reg(apiV1, "/model/list", http.MethodGet, v1.ListModels, "导入模型列表展示")
	mid.Sub("model.model_management").Reg(apiV1, "/model/status", http.MethodPut, v1.ChangeModelStatus, "模型启用/关闭")
	mid.Sub("model.model_management").Reg(apiV1, "/model/recommend", http.MethodGet, v1.RecommendModels, "获取推荐模型列表")
	mid.Sub("model.model_management").Reg(apiV1, "/model/validate-thinking", http.MethodPost, v1.ValidateModelThinking, "深度思考校验")
	mid.Sub("model.model_management").Reg(apiV1, "/model/import/providers", http.MethodGet, v1.ListImportProviders, "模型导入-获取供应商列表")

	mid.Sub("model.model_management").Reg(apiV1, "/model/experience/llm/connect", http.MethodPost, v1.ModelExperienceLLMConnect, "模型体验流式问答断开后重连")
	mid.Sub("model.model_management").Reg(apiV1, "/model/experience/llm/cancel", http.MethodPost, v1.ModelExperienceLLMCancel, "模型体验流式问答手动停止")
	mid.Sub("model.model_management").Reg(apiV1, "/model/experience/llm", http.MethodPost, v1.ModelExperienceLLM, "LLM模型体验", append([]gin.HandlerFunc{middleware.TraceWeb(constant.BizModuleModel, middleware.WithModelResource("modelId"))}, middleware.AuthModelByModelId([]string{"modelId"}))...)
	mid.Sub("model.model_management").Reg(apiV1, "/model/experience/dialog", http.MethodPost, v1.ModelExperienceSaveDialog, "保存模型体验对话", middleware.AuthModelByModelId([]string{"modelId"}))
	mid.Sub("model.model_management").Reg(apiV1, "/model/experience/dialogs", http.MethodGet, v1.ModelExperienceListDialogs, "获取模型体验对话列表")
	mid.Sub("model.model_management").Reg(apiV1, "/model/experience/dialog", http.MethodDelete, v1.ModelExperienceDeleteDialog, "删除模型体验对话")
	mid.Sub("model.model_management").Reg(apiV1, "/model/experience/dialog/records", http.MethodGet, v1.ModelExperienceListDialogRecords, "查询模型体验对话历史记录")

	mid.Sub("model.model_management").Reg(apiV1, "/prompt/custom/list", http.MethodGet, v1.GetCustomPromptList, "获取自定义Prompt列表")
	mid.Sub("model.model_management").Reg(apiV1, "/prompt/template/list", http.MethodGet, v1.GetPromptTemplateList, "获取提示词模板列表")
}
