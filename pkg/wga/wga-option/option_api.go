package wga_option

import (
	"github.com/UnicomAI/wanwu/pkg/wga/internal/config"
	"github.com/UnicomAI/wanwu/pkg/wga/internal/option"
	"github.com/cloudwego/eino/adk"
)

type Option = option.Option
type ModelConfig = option.ModelConfig
type ToolConfig = option.ToolConfig
type ExtraTool = option.ExtraTool
type Skill = option.Skill
type SkillVariable = config.SkillVariable
type MCP = option.MCP
type RunSession = option.RunSession

type CheckResult = option.CheckResult

type ToolCategoryInfo = config.ToolCategory

type SystemMessageStrategy = option.SystemMessageStrategy

const SystemMessageStrategyMerge = option.SystemMessageStrategyMerge

// WithModelConfig 设置模型配置。
func WithModelConfig(model ModelConfig) Option {
	return option.WithModelConfig(model)
}

// WithToolConfig 添加工具配置，工具标题不能重复。
func WithToolConfig(tool ToolConfig) Option {
	return option.WithToolConfig(tool)
}

// WithExtraTool 添加额外工具（非配置文件中的工具）。
// 工具标题不能与配置文件中的工具重复，也不能与已添加的额外工具重复。
func WithExtraTool(tool ExtraTool) Option {
	return option.WithExtraTool(tool)
}

// WithMCP 添加 MCP 服务器。
func WithMCP(mcp MCP) Option {
	return option.WithMCP(mcp)
}

// WithSkill 添加技能（非配置文件中的技能）。
func WithSkill(skill Skill) Option {
	return option.WithSkill(skill)
}

// WithInputDir 设置输入目录。
// 输入目录的内容会在执行前复制到沙箱工作目录。
// 支持 "/." 后缀：如 "/path/to/dir/." 表示复制目录内容而非目录本身。
func WithInputDir(inputDir string) Option {
	return option.WithInputDir(inputDir)
}

// WithOutputDir 设置输出目录。
// 沙箱工作目录的内容会在执行后复制到该目录。
// 注意：隐藏文件（以 "." 开头）不会被复制。
func WithOutputDir(outputDir string) Option {
	return option.WithOutputDir(outputDir)
}

// WithRunSession 设置执行会话标识（ThreadID 和 RunID）。
func WithRunSession(session RunSession) Option {
	return option.WithRunSession(session)
}

// WithMessages 设置历史消息。
func WithMessages(messages []adk.Message) Option {
	return option.WithMessages(messages)
}

// WithInstruction 设置运行时动态指令，覆盖配置文件中的 prompt.md。
func WithInstruction(instruction string) Option {
	return option.WithInstruction(instruction)
}

// WithOverallTask 设置运行时动态整体任务。
func WithOverallTask(overallTask string) Option {
	return option.WithOverallTask(overallTask)
}

// WithEnableHumanInTheLoop 设置是否启用人机交互。
// enableCustom 为可选参数，设置是否允许用户自定义回答。
func WithEnableHumanInTheLoop(enable bool, enableCustom ...bool) Option {
	return option.WithEnableHumanInTheLoop(enable, enableCustom...)
}

func WithSystemMessageStrategy(strategy SystemMessageStrategy) Option {
	return option.WithSystemMessageStrategy(strategy)
}
