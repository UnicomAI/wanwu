// Package option 提供智能体运行选项的内部实现。
package option

import (
	"context"
	"fmt"

	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/wga/internal/config"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
)

// ============================================================================
// 类型 - 配置
// ============================================================================

// ModelConfig 模型配置。
type ModelConfig struct {
	Provider     string               // 提供商标识
	ProviderName string               // 提供商显示名称
	BaseURL      string               // API 基础地址
	APIKey       string               // API 密钥
	Model        string               // 模型标识
	ModelName    string               // 模型显示名称
	Params       *mp_common.LLMParams // 模型参数
}

// ToolConfig 工具配置。
type ToolConfig struct {
	Title   string                  // 工具标题（对应 OpenAPI schema 的 info.title）
	APIAuth *util.ApiAuthWebRequest // API 认证配置
}

// ExtraTool 额外工具配置（非配置文件中的工具）。
type ExtraTool struct {
	OpenAPI3Schema *openapi3.T             // OpenAPI 3.0 schema（必须）
	APIAuth        *util.ApiAuthWebRequest // API 认证（可选）
}

// Skill 技能配置。
type Skill struct {
	Dir       string                 `json:"dir"`       // skill 目录路径（相对程序运行目录）
	Variables []config.SkillVariable `json:"variables"` // 用户自定义变量
}

// MCP MCP 服务器配置。
type MCP struct {
	Name        string                  // MCP 名称
	URL         string                  // MCP SSE/STREAMABLE 服务器地址
	Description string                  // MCP 描述
	ApiAuth     *util.ApiAuthWebRequest // API 认证配置（可选）
	Headers     map[string]string       // 自定义请求头（可选）
}

// WorkspaceConfig 工作空间配置。
type WorkspaceConfig struct {
	InputDir  string // 输入目录路径
	OutputDir string // 输出目录路径
}

// RunSession 执行会话标识。
type RunSession struct {
	ThreadID string // 对话会话 ID
	RunID    string // 执行请求 ID
}

// ============================================================================
// Option/Options
// ============================================================================

// Option 选项接口。
type Option interface {
	apply(*Options) error
}

// optionFunc 选项函数。
type optionFunc func(*Options) error

func (f optionFunc) apply(opts *Options) error {
	return f(opts)
}

// Options 智能体运行选项。
type Options struct {
	RunSession                 RunSession            // 执行会话标识
	Workspace                  WorkspaceConfig       // 工作空间配置
	Model                      ModelConfig           // 模型配置
	Tools                      []ToolConfig          // 工具配置列表（配置文件工具的认证）
	ExtraTools                 []ExtraTool           // 额外工具列表（运行时传入）
	Skills                     []Skill               // 技能列表（运行时传入）
	MCPs                       []MCP                 // MCP 服务器列表
	Messages                   []adk.Message         // 历史消息 + 当前问题（最后一条 User 消息）
	Instruction                string                // 运行时动态指令（覆盖配置文件 prompt.md）
	OverallTask                string                // 运行时动态整体任务（用于子智能体）
	EnableHumanInTheLoop       bool                  // 是否启用人机交互
	EnableHumanInTheLoopCustom bool                  // 是否允许用户自定义回答
	SystemMessageStrategy      SystemMessageStrategy // system 消息处理策略，默认不处理
}

// Apply 应用选项。
// 如果 ThreadID 或 RunID 为空，自动生成 UUID。
func (options *Options) Apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt.apply(options); err != nil {
			return err
		}
	}
	if options.RunSession.ThreadID == "" {
		options.RunSession.ThreadID = uuid.New().String()
	}
	if options.RunSession.RunID == "" {
		options.RunSession.RunID = uuid.New().String()
	}
	return nil
}

// ============================================================================
// Check
// ============================================================================

// CheckResult 条件检查结果。
type CheckResult struct {
	ToolCategories []CheckToolCategory // 工具类别检查结果
}

// CheckToolCategory 工具类别检查结果。
type CheckToolCategory struct {
	Category  string      // 工具类别类型
	Condition string      // 工具类别条件
	Meet      bool        // 是否满足条件
	Tools     []CheckTool // 工具检查结果
}

// CheckTool 工具检查结果。
type CheckTool struct {
	Title string // 工具标题
	Meet  bool   // 是否满足条件
}

// CheckModelConfig 检查模型配置是否有效。
func (options *Options) CheckModelConfig() error {
	return options.checkModel()
}

// CheckMessages 检查消息配置是否有效。
func (options *Options) CheckMessages() error {
	if len(options.Messages) == 0 {
		return fmt.Errorf("messages is empty")
	}
	lastMsg := options.Messages[len(options.Messages)-1]
	if lastMsg.Role != schema.User {
		return fmt.Errorf("last message must be user message, got %s", lastMsg.Role)
	}
	return nil
}

// CheckTools 检查工具配置（包括配置文件工具条件和额外工具冲突检查）。
func (options *Options) CheckTools(cfg *config.Agent) ([]CheckToolCategory, error) {
	categories := cfg.CollectToolCategories()
	if err := options.checkExtraToolsConflict(categories); err != nil {
		return nil, err
	}
	return options.checkToolCategories(categories)
}

// ============================================================================
// 选项函数
// ============================================================================

// WithModelConfig 设置模型配置。
func WithModelConfig(model ModelConfig) Option {
	return optionFunc(func(opts *Options) error {
		opts.Model = model
		return nil
	})
}

// WithToolConfig 添加工具配置，工具标题不能重复。
func WithToolConfig(tool ToolConfig) Option {
	return optionFunc(func(opts *Options) error {
		if tool.APIAuth != nil {
			if err := tool.APIAuth.Check(); err != nil {
				return fmt.Errorf("tool (%v) check auth err: %v", tool.Title, err)
			}
		}
		for _, toolOpt := range opts.Tools {
			if toolOpt.Title == tool.Title {
				return fmt.Errorf("tool (%v) already exist", tool.Title)
			}
		}
		opts.Tools = append(opts.Tools, tool)
		return nil
	})
}

// WithExtraTool 添加额外工具（非配置文件中的工具）。
// 工具标题不能与配置文件中的工具重复，也不能与已添加的额外工具重复。
func WithExtraTool(tool ExtraTool) Option {
	return optionFunc(func(opts *Options) error {
		if tool.OpenAPI3Schema == nil {
			return fmt.Errorf("extra tool schema is required")
		}
		if tool.OpenAPI3Schema.Info == nil || tool.OpenAPI3Schema.Info.Title == "" {
			return fmt.Errorf("extra tool schema must have title")
		}
		if err := openapi3_util.ValidateDoc(context.Background(), tool.OpenAPI3Schema); err != nil {
			return fmt.Errorf("extra tool schema invalid: %w", err)
		}
		if tool.APIAuth != nil {
			if err := tool.APIAuth.Check(); err != nil {
				return fmt.Errorf("extra tool (%s) check auth err: %w", tool.OpenAPI3Schema.Info.Title, err)
			}
		}
		opts.ExtraTools = append(opts.ExtraTools, tool)
		return nil
	})
}

// WithMCP 添加 MCP 服务器。
func WithMCP(mcp MCP) Option {
	return optionFunc(func(opts *Options) error {
		if mcp.Name == "" {
			return fmt.Errorf("mcp name is required")
		}
		if mcp.URL == "" {
			return fmt.Errorf("mcp [%s] url is required", mcp.Name)
		}
		opts.MCPs = append(opts.MCPs, mcp)
		return nil
	})
}

// WithSkill 添加技能（非配置文件中的技能）。
func WithSkill(skill Skill) Option {
	return optionFunc(func(opts *Options) error {
		if skill.Dir == "" {
			return fmt.Errorf("skill dir is required")
		}
		// 验证变量
		seen := make(map[string]bool)
		for _, v := range skill.Variables {
			if v.Name == "" {
				return fmt.Errorf("skill [%s] variable name is required", skill.Dir)
			}
			if seen[v.Name] {
				return fmt.Errorf("skill [%s] variable name [%s] duplicate", skill.Dir, v.Name)
			}
			if v.VariableKey == "" {
				return fmt.Errorf("skill [%s] variable [%s] key is required", skill.Dir, v.Name)
			}
			if v.VariableValue == "" {
				return fmt.Errorf("skill [%s] variable [%s] value is required", skill.Dir, v.Name)
			}
			seen[v.Name] = true
		}
		opts.Skills = append(opts.Skills, skill)
		return nil
	})
}

// WithInputDir 设置输入目录。
// 输入目录的内容会在执行前复制到沙箱工作目录。
// 支持 "/." 后缀：如 "/path/to/dir/." 表示复制目录内容而非目录本身。
func WithInputDir(inputDir string) Option {
	return optionFunc(func(opts *Options) error {
		opts.Workspace.InputDir = inputDir
		return nil
	})
}

// WithOutputDir 设置输出目录。
// 沙箱工作目录的内容会在执行后复制到该目录。
// 注意：隐藏文件（以 "." 开头）不会被复制。
func WithOutputDir(outputDir string) Option {
	return optionFunc(func(opts *Options) error {
		opts.Workspace.OutputDir = outputDir
		return nil
	})
}

// WithRunSession 设置执行会话标识（ThreadID 和 RunID）。
func WithRunSession(session RunSession) Option {
	return optionFunc(func(opts *Options) error {
		opts.RunSession = session
		return nil
	})
}

// WithMessages 设置消息列表，最后一条消息必须是 User 消息。
func WithMessages(messages []adk.Message) Option {
	return optionFunc(func(opts *Options) error {
		opts.Messages = append(opts.Messages, messages...)
		return nil
	})
}

// WithInstruction 设置运行时动态指令，覆盖配置文件中的 prompt.md。
func WithInstruction(instruction string) Option {
	return optionFunc(func(opts *Options) error {
		opts.Instruction = instruction
		return nil
	})
}

// WithOverallTask 设置运行时动态整体任务。
func WithOverallTask(overallTask string) Option {
	return optionFunc(func(opts *Options) error {
		opts.OverallTask = overallTask
		return nil
	})
}

// WithEnableHumanInTheLoop 设置是否启用人机交互。
// enableCustom 为可选参数，设置是否允许用户自定义回答。
func WithEnableHumanInTheLoop(enable bool, enableCustom ...bool) Option {
	return optionFunc(func(opts *Options) error {
		opts.EnableHumanInTheLoop = enable
		if len(enableCustom) > 0 {
			opts.EnableHumanInTheLoopCustom = enableCustom[0]
		}
		return nil
	})
}
