package service

import (
	"fmt"
	"strings"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	wga_persistent "github.com/UnicomAI/wanwu/pkg/wga-persistent"
	"github.com/gin-gonic/gin"
)

// digitalEmployeeChatAgentID 数字员工发布会话对话固定使用 DIP Agent
// （与 wga 通用智能体的 @数字员工 模式共用同一 agent，区别在注入来源：员工详情直接按 employeeId 拉取）
const digitalEmployeeChatAgentID = "DIP Agent"

// convertDigitalEmployeeModelConfig 转换并校验数字员工发布会话的模型配置
// modelConfig 为 nil 时返回 nil,nil（对话使用默认配置）
func convertDigitalEmployeeModelConfig(ctx *gin.Context, modelConfig *request.AppModelConfig) (*common.AppModelConfig, error) {
	if modelConfig == nil {
		return nil, nil
	}
	protoConfig, err := appModelConfigModel2Proto(*modelConfig)
	if err != nil {
		return nil, err
	}
	if err := checkModelConfigFromProto(ctx, protoConfig); err != nil {
		return nil, err
	}
	return protoConfig, nil
}

// SyncDigitalEmployeePublish 数字员工发布状态同步（外部系统回调 BFF callback 端口）
// SyncDigitalEmployeePublish 数字员工发布同步（外部系统回调 BFF callback 端口，万悟自拟格式）
// 仅支持发布：app.PublishApp（upsert app 表 publish_type，幂等）；身份（userId/orgId）由外部系统在 body 提供。
// 取消发布走 DELETE 接口（SyncDigitalEmployeeUnpublish）。
// 不做任何快照/版本校验——发布流程在外部，万悟只登记展示状态。
func SyncDigitalEmployeePublish(ctx *gin.Context, req *request.DigitalEmployeePublishSyncReq) error {
	employeeId := strings.TrimSpace(req.EmployeeId)
	if employeeId == "" {
		return grpc_util.ErrorStatus(errs.Code_BFFGeneral, "employeeId is required")
	}
	publishType := strings.TrimSpace(req.PublishType)
	if publishType == "" {
		return grpc_util.ErrorStatus(errs.Code_BFFGeneral, "publishType is required")
	}
	_, err := app.PublishApp(ctx.Request.Context(), &app_service.PublishAppReq{
		AppId:       employeeId,
		AppType:     constant.AppTypeDigitalEmployee,
		PublishType: publishType,
		UserId:      req.UserId,
		OrgId:       req.OrgId,
	})
	return err
}

// SyncDigitalEmployeeUnpublish 数字员工删除/下架同步（外部系统回调 BFF callback 端口，万悟自拟格式）
// 以 employeeId 为准调用 app.UnPublishApp 删除 app 行（幂等，目标不存在也返回成功）。
func SyncDigitalEmployeeUnpublish(ctx *gin.Context, req *request.DeleteDigitalEmployeePublishSyncReq) error {
	employeeId := strings.TrimSpace(req.EmployeeId)
	if employeeId == "" {
		return grpc_util.ErrorStatus(errs.Code_BFFGeneral, "employeeId is required")
	}
	_, err := app.UnPublishApp(ctx.Request.Context(), &app_service.UnPublishAppReq{
		AppId:   employeeId,
		AppType: constant.AppTypeDigitalEmployee,
		UserId:  req.UserId,
	})
	return err
}

// GetDigitalEmployeeSquareDetail 数字员工广场详情（实时调外部详情，仅返回前端需要的 name/avatar）
// 决策 D6：展示字段实时调外部详情，不做本地缓存
func GetDigitalEmployeeSquareDetail(ctx *gin.Context, userId, orgId, employeeId string) (*response.DigitalEmployeeSquareDetail, error) {
	employee, err := GetDigitalEmployeeInfo(ctx, userId, orgId, employeeId)
	if err != nil {
		return nil, err
	}
	if employee == nil { // 未发布/不存在（契约：200 + data:null）
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_digital_employee_info", fmt.Sprintf("digital employee (%v) not found", employeeId))
	}

	detail := &response.DigitalEmployeeSquareDetail{
		Name:        employee.Name,
		Avatar:      request.Avatar{Path: "/v1/static/icon/wga-digital-employee-icon.svg"}, // 契约 latest 无头像字段，用默认图标
		Placeholder: config.Cfg().Ontology.DigitalEmployeeChatPlaceholder,                  // DE 对话输入框占位文案（区别于通用智能体 DIP Agent 的 placeholder）
	}
	return detail, nil
}

// CreateDigitalEmployeeConversation 创建数字员工发布会话（独立表 digital_employee_conversation，绑定数字员工，创建后不可切换）
func CreateDigitalEmployeeConversation(ctx *gin.Context, userId, orgId string, req request.CreateDigitalEmployeeConversationReq) (*response.CreateDigitalEmployeeConversationResp, error) {
	modelConfig, err := convertDigitalEmployeeModelConfig(ctx, req.ModelConfig)
	if err != nil {
		return nil, err
	}

	resp, err := assistant.DigitalEmployeeConversationCreate(ctx.Request.Context(), &assistant_service.DigitalEmployeeConversationCreateReq{
		Prompt:      strings.TrimSpace(req.Title),
		ModelConfig: modelConfig,
		EmployeeId:  req.EmployeeId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}

	return &response.CreateDigitalEmployeeConversationResp{ConversationID: resp.ThreadId}, nil
}

// DeleteDigitalEmployeeConversation 删除数字员工发布会话（DB 行 + 独立 ES 索引历史，对齐 wga 删除级联）
func DeleteDigitalEmployeeConversation(ctx *gin.Context, userId, orgId string, req request.DeleteDigitalEmployeeConversationReq) error {
	if _, err := assistant.DigitalEmployeeConversationDelete(ctx.Request.Context(), &assistant_service.DigitalEmployeeConversationDeleteReq{
		ThreadId: req.ConversationId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	}); err != nil {
		return err
	}

	if _, err := assistant.DeleteFromES(ctx.Request.Context(), &assistant_service.DeleteFromESReq{
		IndexName: digitalEmployeeChatHistoryEventESIndexName,
		Conditions: map[string]string{
			"threadId": req.ConversationId,
			"userId":   userId,
			"orgId":    orgId,
		},
	}); err != nil && !esIndexNotFound(err, digitalEmployeeChatHistoryEventESIndexName) {
		log.Errorf("[digital-employee] conversation %v delete chat history from ES err: %v", req.ConversationId, err)
	}

	return nil
}

// GetDigitalEmployeeConversationList 数字员工发布会话列表（按 employeeId 维度过滤）
func GetDigitalEmployeeConversationList(ctx *gin.Context, userId, orgId string, req request.GetDigitalEmployeeConversationListReq) (*response.ListResult, error) {
	resp, err := assistant.DigitalEmployeeConversationList(ctx.Request.Context(), &assistant_service.DigitalEmployeeConversationListReq{
		PageNo:     int32(req.PageNo),
		PageSize:   int32(req.PageSize),
		SearchText: strings.TrimSpace(req.SearchText),
		EmployeeId: req.EmployeeId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}

	items := make([]response.DigitalEmployeeConversationInfo, 0, len(resp.Data))
	for _, info := range resp.Data {
		items = append(items, response.DigitalEmployeeConversationInfo{
			ConversationID: info.ThreadId,
			EmployeeID:     info.EmployeeId,
			Title:          info.Title,
			CreatedAt:      util.Time2Str(info.CreatedAt),
			UpdatedAt:      util.Time2Str(info.UpdatedAt),
		})
	}
	return &response.ListResult{List: items, Total: resp.Total}, nil
}

// GetDigitalEmployeeConversationDetail 数字员工发布会话详情（从独立 ES 索引回放历史）
func GetDigitalEmployeeConversationDetail(ctx *gin.Context, userId, orgId string, req request.GetDigitalEmployeeConversationDetailReq) (*response.ListResult, error) {
	if _, err := assistant.GetDigitalEmployeeConversationConfig(ctx.Request.Context(), &assistant_service.GetDigitalEmployeeConversationConfigReq{
		ThreadId: req.ConversationId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	}); err != nil {
		return &response.ListResult{}, nil
	}

	return getWgaConversationDetailFromES(ctx, userId, orgId, req.ConversationId, digitalEmployeeChatHistoryEventESIndexName)
}

// GetDigitalEmployeeConversationConfig 数字员工发布会话配置（读回 modelConfig，对齐 wga GET /general/agent/conversation/config）
func GetDigitalEmployeeConversationConfig(ctx *gin.Context, userId, orgId, conversationId string) (*response.GetDigitalEmployeeConversationConfigResp, error) {
	resp, err := assistant.GetDigitalEmployeeConversationConfig(ctx.Request.Context(), &assistant_service.GetDigitalEmployeeConversationConfigReq{
		ThreadId: conversationId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	cfg := resp.GetConfig()
	if cfg == nil {
		cfg = &assistant_service.DigitalEmployeeConversationConfig{}
	}

	result := &response.GetDigitalEmployeeConversationConfigResp{
		ConversationID: cfg.ThreadId,
		EmployeeID:     cfg.EmployeeId,
		Title:          cfg.Title,
		ModelConfig:    request.AppModelConfig{},
	}

	// 处理模型配置 - 验证模型是否存在且已启用，返回 DisplayName（对齐 wga GET config）
	if cfg.ModelConfig != nil && cfg.ModelConfig.ModelId != "" {
		modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: cfg.ModelConfig.ModelId})
		if err == nil && modelInfo != nil && modelInfo.IsActive {
			result.ModelConfig = request.AppModelConfig{
				Provider:    cfg.ModelConfig.Provider,
				Model:       cfg.ModelConfig.Model,
				ModelId:     cfg.ModelConfig.ModelId,
				ModelType:   cfg.ModelConfig.ModelType,
				DisplayName: modelInfo.DisplayName,
			}
		}
		// 模型不存在或未启用 → 返回空 ModelConfig
	}
	return result, nil
}

// UpdateDigitalEmployeeConversationConfig 更新数字员工发布会话的模型配置（对齐 wga PUT /general/agent/conversation/config）
func UpdateDigitalEmployeeConversationConfig(ctx *gin.Context, userId, orgId string, req request.UpdateDigitalEmployeeConversationConfigReq) error {
	modelConfig, err := convertDigitalEmployeeModelConfig(ctx, req.ModelConfig)
	if err != nil {
		return err
	}
	if modelConfig == nil {
		return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, "modelConfig is required for conversation")
	}
	_, err = assistant.UpdateDigitalEmployeeConversationConfig(ctx.Request.Context(), &assistant_service.UpdateDigitalEmployeeConversationConfigReq{
		ThreadId:    req.ConversationId,
		ModelConfig: modelConfig,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

// DigitalEmployeeChat 数字员工发布对话（wga 模式，固定 DIP Agent，SSE 流式返回）
// 两步模式：会话须先通过 CreateDigitalEmployeeConversation 创建，此处仅续聊；
// modelConfig 从会话表读回、不重发即生效；会话绑定数字员工，不可切换
func DigitalEmployeeChat(ctx *gin.Context, userId, orgId, clientId string, req request.DigitalEmployeeChatReq) error {
	if config.Cfg().Ontology.Enable == 0 {
		return errDigitalEmployeeNotReady()
	}

	threadID := strings.TrimSpace(req.ConversationId)

	// 读回会话配置，校验数字员工不可切换，modelConfig 从表读回
	configResp, err := assistant.GetDigitalEmployeeConversationConfig(ctx.Request.Context(), &assistant_service.GetDigitalEmployeeConversationConfigReq{
		ThreadId: threadID,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return err
	}
	if configResp.Config == nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_digital_employee_conversation_not_found", fmt.Sprintf("conversation (%v) not found", threadID))
	}
	if configResp.Config.EmployeeId != req.EmployeeId {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_digital_employee_switch_forbidden", "数字员工不可切换")
	}
	modelConfig := configResp.Config.ModelConfig

	// 获取 threadId 的 workspace store（对齐通用智能体，DIP Agent 运行产出写回 workspace，
	// 与 /general/agent/conversation/workspace* 接口按 threadId 寻址一致）
	var workspaceStore *wga_persistent.Store
	if config.WgaCfg().Persistent.Enabled {
		store, err := NewGeneralAgentWorkspaceStore(threadID)
		if err != nil {
			log.Errorf("[digital-employee] thread %v failed to create persistent store: %v", threadID, err)
		} else {
			workspaceStore = store
		}
	}

	return WgaConversationChat(ctx, &WgaChatParams{
		UserID:               userId,
		OrgID:                orgId,
		AgentID:              digitalEmployeeChatAgentID,
		ThreadID:             threadID,
		Messages:             req.Messages,
		ClientID:             clientId,
		ModelConfig:          modelConfig,
		WorkspaceStore:       workspaceStore,
		SendWorkspaceEvent:   true,
		EnableHumanInTheLoop: true,
		ESIndexName:          digitalEmployeeChatHistoryEventESIndexName,
		EmployeeID:           req.EmployeeId,
	})
}
