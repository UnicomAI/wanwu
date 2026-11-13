package statistic

import "github.com/UnicomAI/wanwu/pkg/constant"

// ViewScope 统计看板视角（BFF request 校验与 ORM 查询过滤共用）。
const (
	ViewScopePublished = "published" // 我发布的：按发布者 / module_creator / model_creator
	ViewScopeUsed      = "used"      // 我使用的：按调用人 user_id / org_id
)

// SuccessStatusCode 统计成功判定：HTTP 语义 200。
const SuccessStatusCode int64 = 200

// IsSuccess 由 statusCode 反推本次调用是否计入成功（与写入/聚合口径一致）。
func IsSuccess(statusCode int64) bool {
	return statusCode == SuccessStatusCode
}

// StatisticModulesAllowingEmptyAppID 返回应用统计中允许 appId 为空的板块（板块级维度）。
// 写入校验与 sqlopt.WithStatisticAppIDFilter 共用该列表，避免两处白名单漂移。
func StatisticModulesAllowingEmptyAppID() []string {
	return []string{
		constant.BizModuleWGA,
		constant.BizModuleModel,
		constant.BizModuleResourceSkill,
		constant.BizModuleResourceKnowledge,
		constant.BizModuleResourcePrompt,
	}
}

// StatisticModuleAllowsEmptyAppID 应用统计写入/查询：这些板块允许 appId 为空（板块级维度）。
func StatisticModuleAllowsEmptyAppID(module string) bool {
	for _, m := range StatisticModulesAllowingEmptyAppID() {
		if m == module {
			return true
		}
	}
	return false
}
