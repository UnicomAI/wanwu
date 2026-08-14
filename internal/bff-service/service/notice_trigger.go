package service

import (
	"context"
	"fmt"
	"time"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	knowledgebase_permission_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-permission-service"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	operate_service "github.com/UnicomAI/wanwu/api/proto/operate-service"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/gin-gonic/gin"
)

// 本文件收敛消息中心的全部触发埋点（11 个触发场景：应用发布/删除，模型导入/删除/停用启用，
// 知识库删库/共享/取消共享/改权限/转让管理员），避免散落在各业务函数里漏埋。
//
// best-effort 契约（强制）：消息生成是旁路，失败只打日志 + 打点，
// 绝不返回 error 到业务主流程、绝不回滚发布/共享。允许极低概率永久丢失——
// 通知类消息非事务类，丢失不产生数据不一致，探索广场/知识库列表是信息的兜底获取路径。

// 归一化后的发布范围（与 operate-service 的 orm.ScopeXxx 对齐）。
// 应用 PublishType（字符串三态）与模型 ScopeType（1/2/3）语义同构，
// 在这里统一归一化后再进消息域——差分矩阵对资源类型无感知是设计的核心不变量。
const (
	noticeScopeNone         = ""
	noticeScopePrivate      = "private"
	noticeScopeOrganization = "organization"
	noticeScopePublic       = "public"
)

// 消息来源资源类型（应用五类直接用 constant.AppTypeXxx）
const (
	noticeResourceModel     = "model"
	noticeResourceKnowledge = "knowledge"
)

// 名单型文案变体（一段式，与 operate-service 的 model.EventVariantXxx 对齐）
const (
	noticeVariantShared      = "shared"
	noticeVariantUnshared    = "unshared"
	noticeVariantDeleted     = "deleted"
	noticeVariantPermChanged = "perm_changed"
)

// 知识库权限级别 → 文案（proto/knowledgebase-permission-service 的 permissionType）
var knowledgePermissionLabels = map[int32]string{
	0:                "查看",
	10:               "编辑",
	20:               "授权",
	SystemPermission: "管理员",
}

// noticeTimeout 旁路通知 RPC 的独立超时。正常通知 RPC 是毫秒级，2 秒是"下游卡死时
// 兜底止损"的保险丝而非常态——卡到 2 秒说明 operate-service 已异常，掐掉比让 goroutine
// 永久挂起（内存泄漏）强。
const noticeTimeout = 2 * time.Second

// safeNotice 统一的 best-effort 包裹：异步执行 + 独立 ctx + 短超时 + panic 兜底，不向上抛错。
//
// 异步（go）让通知不阻塞已完成发布/共享的 HTTP 响应；独立 ctx（context.Background 派生，
// 不复用请求 ctx）让通知的生命周期与请求解耦——请求结束/客户端断开不会取消正在发的通知；
// 2 秒超时兜底防 goroutine 因下游卡死永久挂起。
//
// fn 收到的 ctx 即该独立超时 ctx，fn 内的 RPC 必须用它而非请求 ctx。
//
// 已知遗留：失败仅 log 打点，未接入真实计数器（notice_create_failed_total{source_type,reason}）
// 与告警。全仓暂无 metrics 基础设施，待统一搭建后补。
//
// 进程关停风险：go 出去的通知在进程重启时会丢。消息中心通知是 best-effort 附带功能
// （丢失用户无感、列表是兜底获取路径），此代价可接受，不引入关停等待机制以免复杂化。
func safeNotice(sourceType string, fn func(ctx context.Context) error) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// 打点：notice_create_failed_total{source_type, reason=panic}
				log.Errorf("notice_create_failed source_type=%v reason=panic detail=%v", sourceType, r)
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), noticeTimeout)
		defer cancel()
		if err := fn(ctx); err != nil {
			// 打点：notice_create_failed_total{source_type, reason=rpc_error}
			log.Errorf("notice_create_failed source_type=%v reason=rpc_error detail=%v", sourceType, err)
		}
	}()
}

// --- 应用（发布/改范围/发新版本/删除等触发场景） ---

// notifyAppPublish 应用发布 / 发新版本（仅 PublishApp 发布流程调用）。
//
// version 恒非空（发布接口校验 required + vX.Y.Z）：受众 = 整个新范围一条消息，
// 文案统一「已上线」。
// 「只改范围不发版本」（UpdateAppVersion）不再触发本函数；范围变更导致的收件人减少不发下线。
func notifyAppPublish(ctx *gin.Context, userId, orgId, appId, appType, newPublishType, version string) {
	// 幂等键设计：消息中心按 (event_id, event_variant) 唯一键去重，键语义决定了去重效果。
	// 若用「目标范围状态」做键（如 app:{id}:public），反复 private↔public 切换时
	// 每次切回 public 都撞同一个键 → 第二次起的「上线」消息被去重吞掉，用户收不到。
	// 所以键必须刻画「这一次发布动作」而非「状态」：时间戳把每次动作区分开。
	// 同一请求内只取一次 → gRPC 重试复用同值仍能去重（重试不算新动作）；
	// 下次发布取新时间戳 → 不撞键。消息永久保留，键空间随历史量增长。
	ts := time.Now().UnixMilli()
	// 查名同步留在主流程：resolveNoticeAppName 依赖 *gin.Context，不能进 goroutine；
	// 且查名是 bff→内部服务的正常毫秒级调用。查名失败则跳过通知（与原逻辑一致）。
	name, err := resolveNoticeAppName(ctx, userId, orgId, appId, appType)
	if err != nil {
		log.Errorf("notice_create_failed source_type=%v reason=resolve_name detail=%v", appType, err)
		return
	}
	safeNotice(appType, func(noticeCtx context.Context) error {
		_, err := notice.CreateAppNotice(noticeCtx, &operate_service.CreateAppNoticeReq{
			AppId:       appId,
			AppType:     appType,
			AppName:     name,
			OrgId:       orgId,
			NewScope:    normalizeAppScope(newPublishType),
			SenderId:    userId,
			SenderOrgId: orgId,
			Version:     version,
			EventId:     fmt.Sprintf("app:%v:%v:%v:%v:%v", appType, appId, newPublishType, version, ts),
		})
		return err
	})
}

// notifyAppDeleted 删除应用本体：原可见者收「已下线」。
// 应用删除仍发下线（与"范围变更不发下线"相反）；受众由删除前查到的 oldPublishType 决定，
// newScope 传空串（orm 走删除分支，oldScope 全体收 offline）。
func notifyAppDeleted(ctx *gin.Context, userId, orgId, appId, appType, appName, oldPublishType string) {
	// 同 notifyAppPublish：时间戳区分每次删除动作，避免删→重建→再删撞键
	ts := time.Now().UnixMilli()
	safeNotice(appType, func(noticeCtx context.Context) error {
		_, err := notice.CreateAppNotice(noticeCtx, &operate_service.CreateAppNoticeReq{
			AppId:       appId,
			AppType:     appType,
			AppName:     appName,
			OrgId:       orgId,
			OldScope:    normalizeAppScope(oldPublishType),
			NewScope:    noticeScopeNone,
			SenderId:    userId,
			SenderOrgId: orgId,
			EventId:     fmt.Sprintf("app:%v:%v:deleted:%v", appType, appId, ts),
		})
		return err
	})
}

// normalizeAppScope 应用 PublishType 已经是目标取值，只做未知值兜底
func normalizeAppScope(publishType string) string {
	switch publishType {
	case noticeScopePublic, noticeScopeOrganization, noticeScopePrivate:
		return publishType
	default:
		return noticeScopeNone
	}
}

// resolveNoticeAppName 按 appType 分支复用 bff 已有详情函数查应用名称
// （各 grpc 发布快照响应均不含名称，必须单独查一次）。
func resolveNoticeAppName(ctx *gin.Context, userId, orgId, appId, appType string) (string, error) {
	switch appType {
	case constant.AppTypeAgent:
		resp, err := GetAssistantInfo(ctx, userId, orgId, request.AssistantIdRequest{AssistantId: appId}, false)
		if err != nil {
			return "", err
		}
		return resp.Name, nil
	case constant.AppTypeRag:
		resp, err := rag.GetRagDetail(ctx.Request.Context(), &rag_service.RagDetailReq{
			RagId:    appId,
			Identity: &rag_service.Identity{UserId: userId, OrgId: orgId},
		})
		if err != nil {
			return "", err
		}
		if resp.BriefConfig == nil {
			return "", fmt.Errorf("rag %v brief config nil", appId)
		}
		return resp.BriefConfig.Name, nil
	case constant.AppTypeWorkflow, constant.AppTypeChatflow:
		resp, err := ListWorkflowByIDs(ctx, "", []string{appId})
		if err != nil {
			return "", err
		}
		if resp == nil || len(resp.Workflows) == 0 {
			return "", fmt.Errorf("workflow %v not found", appId)
		}
		return resp.Workflows[0].Name, nil
	case constant.AppTypeSkill:
		resp, err := GetCustomSkill(ctx, userId, orgId, appId)
		if err != nil {
			return "", err
		}
		return resp.Name, nil
	default:
		return "", fmt.Errorf("unsupported appType %v", appType)
	}
}

// appPublishTypeOf 删除应用前查一次发布范围（DeleteApp 之后记录已消失）
func appPublishTypeOf(ctx *gin.Context, appId, appType string) string {
	resp, err := app.GetAppInfo(ctx.Request.Context(), &app_service.GetAppInfoReq{
		AppId:   appId,
		AppType: appType,
	})
	if err != nil {
		log.Errorf("notice: get app info (%v/%v) failed: %v", appType, appId, err)
		return noticeScopeNone
	}
	return resp.PublishType
}

// --- 模型（导入/删除/停用启用） ---

// notifyModelScopeChange 模型导入 / 删除 / 停用启用。
// 停用等效 (scope→"")、启用等效 (""→scope)，代入同一差分公式。
func notifyModelScopeChange(ctx *gin.Context, userId, orgId, modelId, modelName, oldScope, newScope, eventSuffix string) {
	// 时间戳区分每次动作：停用→启用→停用→启用 等反复操作不再撞键
	ts := time.Now().UnixMilli()
	safeNotice(noticeResourceModel, func(noticeCtx context.Context) error {
		_, err := notice.CreateModelNotice(noticeCtx, &operate_service.CreateModelNoticeReq{
			ModelId:     modelId,
			ModelName:   modelName,
			OrgId:       orgId,
			OldScope:    oldScope,
			NewScope:    newScope,
			SenderId:    userId,
			SenderOrgId: orgId,
			EventId:     fmt.Sprintf("model:%v:%v:%v", modelId, eventSuffix, ts),
		})
		return err
	})
}

// normalizeModelScope 模型 ScopeType（1私有/2公开/3组织）→ 统一 scope 字符串
func normalizeModelScope(scopeType string) string {
	switch scopeType {
	case config.ModelScopeTypePublic:
		return noticeScopePublic
	case config.ModelScopeTypeOrg:
		return noticeScopeOrganization
	case config.ModelScopeTypePrivate:
		return noticeScopePrivate
	default:
		return noticeScopeNone
	}
}

// fetchModelSnapshot 删除/停用启用前必须先查旧状态：删除是硬删、状态改完旧值就没了。
// 低频管理操作，查改之间的并发窗口可接受。
func fetchModelSnapshot(ctx *gin.Context, userId, orgId, modelId string) *model_service.ModelInfo {
	resp, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{
		ModelId: modelId,
		UserId:  userId,
		OrgId:   orgId,
	})
	if err != nil {
		log.Errorf("notice: get model (%v) failed: %v", modelId, err)
		return nil
	}
	return resp
}

// --- 知识库（共享/取消共享/改权限/删库/文档增删，名单型） ---

// notifyKnowledgeDelta 知识库名单型消息的统一入口。
//
// eventPrefix 是调用方提供的「事件语义前缀」（如 "knowledge:12:shared:3"），
// 函数内部追加本次请求的时间戳拼成完整幂等键——只刻画语义、不刻画「这一次动作」
// 的前缀会让反复共享/取消共享等操作撞键被吞，时间戳补上「哪一次」这一维。
// 同一请求内只取一次时间戳，gRPC 重试复用同值仍能去重。
func notifyKnowledgeDelta(ctx *gin.Context, userId, orgId, knowledgeId, knowledgeName string,
	gained, lost, changed []*operate_service.NoticeUserOrgPair,
	variant, changedDetail, eventPrefix string) {
	if len(gained) == 0 && len(lost) == 0 && len(changed) == 0 {
		return
	}
	eventId := fmt.Sprintf("%v:%v", eventPrefix, time.Now().UnixMilli())
	safeNotice(noticeResourceKnowledge, func(noticeCtx context.Context) error {
		_, err := notice.CreateKnowledgeNotice(noticeCtx, &operate_service.CreateKnowledgeNoticeReq{
			KnowledgeId:   knowledgeId,
			KnowledgeName: knowledgeName,
			Gained:        gained,
			Lost:          lost,
			Changed:       changed,
			SenderId:      userId,
			SenderOrgId:   orgId,
			ChangedDetail: changedDetail,
			Variant:       variant,
			EventId:       eventId,
		})
		return err
	})
}

// resolveKnowledgeName 查知识库名称（消息文案需要）。失败时退化为 ID，不阻塞消息生成。
func resolveKnowledgeName(ctx *gin.Context, userId, orgId, knowledgeId string) string {
	resp, err := knowledgeBase.SelectKnowledgeDetailById(ctx.Request.Context(), &knowledgebase_service.KnowledgeDetailSelectReq{
		KnowledgeId: knowledgeId,
		UserId:      userId,
		OrgId:       orgId,
	})
	if err != nil {
		log.Errorf("notice: resolve knowledge name (%v) failed: %v", knowledgeId, err)
		return knowledgeId
	}
	return resp.Name
}

// listKnowledgeSharedPairs 查知识库的全量共享名单。
//
// ⚠ 删库必须在删除**之前**调用：删库是异步清权限，事后查不到名单。
// 文档增删等不清权限的场景虽无此限制，但也需先取全量名单作为通知受众。
func listKnowledgeSharedPairs(ctx *gin.Context, userId, orgId, knowledgeId string) []*operate_service.NoticeUserOrgPair {
	list := listKnowledgePermissions(ctx, userId, orgId, knowledgeId)
	ret := make([]*operate_service.NoticeUserOrgPair, 0, len(list))
	for _, info := range list {
		ret = append(ret, &operate_service.NoticeUserOrgPair{UserId: info.UserId, OrgId: info.OrgId})
	}
	return ret
}

// findKnowledgePermissionUser 按 permissionId 反查被操作的二元组。
// 取消共享与转让管理员都只能凭 permissionId 反查旧方（userId, orgId），
// 必须在执行删除/转让 RPC 之前调用，否则权限记录已被改写、无从反查。
func findKnowledgePermissionUser(ctx *gin.Context, userId, orgId, knowledgeId, permissionId string) *operate_service.NoticeUserOrgPair {
	for _, info := range listKnowledgePermissions(ctx, userId, orgId, knowledgeId) {
		if info.PermissionId == permissionId {
			return &operate_service.NoticeUserOrgPair{UserId: info.UserId, OrgId: info.OrgId}
		}
	}
	return nil
}

func listKnowledgePermissions(ctx *gin.Context, userId, orgId, knowledgeId string) []*knowledgebase_permission_service.KnowledgeUserInfo {
	resp, err := knowledgeBasePermission.SelectKnowledgeUserPermission(ctx.Request.Context(),
		&knowledgebase_permission_service.KnowledgeUserPermissionReq{
			KnowledgeId: knowledgeId,
			UserId:      userId,
			OrgId:       orgId,
		})
	if err != nil {
		log.Errorf("notice: list knowledge (%v) permissions failed: %v", knowledgeId, err)
		return nil
	}
	return resp.KnowledgeUserList
}

func knowledgePermissionLabel(permissionType int32) string {
	if label, ok := knowledgePermissionLabels[permissionType]; ok {
		return label
	}
	return "查看"
}
