package request

import (
	"fmt"

	"github.com/UnicomAI/wanwu/pkg/statistic"
)

const (
	// StatisticFilterAll orgIds/userIds 数组中的哨兵值，表示该维度在角色可见范围内「全部」。
	StatisticFilterAll = "ALL"
)

// StatisticFilter 统计看板组织/用户筛选（嵌入各统计与下拉请求）。
//
// 前端约定：组织/用户筛选下拉框仅对管理员渲染，非管理员页面无此入口，
// 正常不会携带 orgIds/userIds。后端按角色解析（实现见 ResolveStatisticScope）：
//   - 非管理员：必须为空（非空报错兜底），统计固定为 JWT 当前 userId + orgId；
//   - 系统管理员：空（未传/[]）或含 "ALL" 均为全量（置空切片，下游 SQL 跳过该维度过滤）；
//     指定 id 原样传递；
//   - 组织管理员：空（未传/[]）报错"筛选范围内无可用组织或用户"；含 "ALL" 组织按 IAM
//     展开可见范围、用户为已解析组织下的全部用户；指定 id 仅查所列 。
type StatisticFilter struct {
	OrgIds  []string `json:"orgIds" `  // 空=按角色解析（见上方约定）；["ALL"]=可见全部组织；["id",...]=指定组织
	UserIds []string `json:"userIds" ` // 空=按角色解析（见上方约定）；["ALL"]=已解析组织下全部用户；["id",...]=指定用户
}

func (f *StatisticFilter) Check() error { return nil }

func checkStatisticViewScope(viewScope string) error {
	switch viewScope {
	case statistic.ViewScopePublished, statistic.ViewScopeUsed:
		return nil
	default:
		return fmt.Errorf("viewScope must be published or used")
	}
}
