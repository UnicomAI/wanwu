// Package config 提供智能体配置的类型定义和加载功能。
package config

// AgentType 智能体类型。
type AgentType string

const (
	// 原子智能体
	AgentTypeReAct   AgentType = "react"   // ReAct 模式智能体
	AgentTypeSandbox AgentType = "sandbox" // 沙箱执行智能体

	// 组合智能体
	AgentTypeSequential AgentType = "sequential" // 顺序执行
	AgentTypeLoop       AgentType = "loop"       // 循环执行
	AgentTypeParallel   AgentType = "parallel"   // 并行执行
	AgentTypeSupervisor AgentType = "supervisor" // 监督者模式
)

// ToolCategoryType 工具类别类型。
// 类型值为 i18n 键。
type ToolCategoryType string

const (
	ToolCategoryTypeSearch ToolCategoryType = "wga_tool_category_search"
)

// ToolCategoryCondition 工具类别条件。
//
//	对于用户动态配置某个智能体的工具集合：
//		 - 智能体运行，需要智能体的所有工具类别都要满足对应条件
//
//	对于用户动态配置某个智能体的每个工具：
//		 - 工具无需鉴权，则该工具默认配置完成，可以直接使用
//		 - 工具需要鉴权，则该工具需要用户提供鉴权完成配置，才能被使用
type ToolCategoryCondition string

const (
	ToolCategoryConditionNone     = "none"     // 无需检查，该类别下的工具都是可选项
	ToolCategoryConditionOptional = "optional" // 该类别下至少有一个工具完成配置
	ToolCategoryConditionRequired = "required" // 该类别下所有工具完成配置
)
