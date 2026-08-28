package trace_util

import (
	"context"
	"fmt"

	"github.com/UnicomAI/wanwu/pkg/constant"
)

// TraceExtra 统计维度 key，由 BFF 中间件写入，record 阶段只读。
const (
	TraceExtraSource             = "source"
	TraceExtraModule             = "module"
	TraceExtraModuleResourceID   = "moduleResourceId"
	TraceExtraModuleResourceType = "moduleResourceType"
	TraceExtraModuleCreatorUser  = "moduleCreatorUserId"
	TraceExtraModuleCreatorOrg   = "moduleCreatorOrgId"
	TraceExtraClientID           = "clientId"
)

// StatisticContext 模型统计上下文（从 TraceInfo 只读解析）。
type StatisticContext struct {
	TraceID string
	// 调用人（与 App/API Key 统计一致；表字段 user_id/org_id）
	UserID string
	OrgID  string

	Source   string
	Module   string
	AppID    string
	AppType  string
	APIKey   string
	APIKeyID string

	MethodPath string

	ModuleCreatorUserID string
	ModuleCreatorOrgID  string
}

// ParseStatisticContext 从 TraceInfo 读取统计维度并校验，不做 path 推断。
func ParseStatisticContext(ctx context.Context) (*StatisticContext, error) {
	traceInfo, err := GetTraceUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("get trace user: %w", err)
	}
	if traceInfo == nil {
		return nil, fmt.Errorf("trace info not found for traceId %s", GetTraceID(ctx))
	}

	extra := traceInfo.GetTraceExtra()
	source := extra[TraceExtraSource]
	module := extra[TraceExtraModule]
	moduleCreatorUserID := extra[TraceExtraModuleCreatorUser]
	moduleCreatorOrgID := extra[TraceExtraModuleCreatorOrg]

	if source == "" {
		return nil, fmt.Errorf("trace extra %q is required", TraceExtraSource)
	}
	if module == "" {
		return nil, fmt.Errorf("trace extra %q is required", TraceExtraModule)
	}
	if moduleCreatorUserID == "" || moduleCreatorOrgID == "" {
		return nil, fmt.Errorf("trace extra moduleCreatorUserId/moduleCreatorOrgId is required")
	}

	clientID := extra[TraceExtraClientID]
	// OpenAPI 保留 MethodPath（与 APIKeyRecord / StatisticApiKey 对齐）；其余 source 不写。
	methodPath := ""
	if source == constant.BizSourceOpenAPI {
		methodPath = traceInfo.GetTraceApi().GetApiPath()
	}

	// proto-generated getter 在字段为 nil 时安全返回零值，无需额外 nil 检查。
	traceUser := traceInfo.GetTraceUser()
	stat := &StatisticContext{
		TraceID:             GetTraceID(ctx),
		Source:              source,
		Module:              module,
		AppID:               extra[TraceExtraModuleResourceID],
		AppType:             extra[TraceExtraModuleResourceType],
		APIKey:              traceUser.GetApiKey(),
		APIKeyID:            traceUser.GetApiKeyId(),
		MethodPath:          methodPath,
		ModuleCreatorUserID: moduleCreatorUserID,
		ModuleCreatorOrgID:  moduleCreatorOrgID,
	}

	switch source {
	case constant.BizSourceWebUrl:
		// webURL 无登录用户：调用人记 app 创建人（与 AppRecord 一致）。
		// X-Client-ID 仅作匿名访客标识校验，不写入 user_id（UUID 不是平台用户）。
		if clientID == "" {
			return nil, fmt.Errorf("webURL source requires X-Client-ID")
		}
		stat.UserID, stat.OrgID = moduleCreatorUserID, moduleCreatorOrgID
	default:
		// OpenAPI / web 等：调用人取 TraceUser（OpenAPI 即 API Key 创建人）。
		stat.UserID = traceUser.GetUserId()
		stat.OrgID = traceUser.GetOrgId()
		if stat.UserID == "" || stat.OrgID == "" {
			return nil, fmt.Errorf("trace user userId/orgId is required for source %q", source)
		}
	}
	return stat, nil
}
