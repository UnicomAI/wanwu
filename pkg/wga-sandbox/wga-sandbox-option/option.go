// Package wga_sandbox_option 提供 wga_sandbox 的选项配置。
package wga_sandbox_option

import (
	"context"
	"fmt"

	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
)

// ============================================================================
// 常量
// ============================================================================

const (
	SandboxTypeReuse   SandboxType = "reuse"   // 复用容器模式
	SandboxTypeOneshot SandboxType = "oneshot" // 一次性容器模式
)

const (
	RunnerTypeOpencode      RunnerType = "opencode"        // opencode 智能体（默认）
	RunnerTypeEinoChatModel RunnerType = "eino-chat-model" // eino-agent ChatModel 智能体
)

const (
	sandboxAPIPort  = 8080 // sandbox API 端口
	opencodeAPIPort = 4096 // opencode API 端口
	einoAPIPort     = 4097 // eino-agent API 端口
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

// Tool 工具配置。
type Tool struct {
	OpenAPI3Schema *openapi3.T // OpenAPI 3.0 schema 文档
	OperationIDs   []string    // 允许的 operations，为空则全部允许
	APIAuth        *openapi3_util.Auth
	Name           string // 工具名称，从 schema 的 info.title 自动读取
}

// SkillVariable 技能变量配置。
type SkillVariable struct {
	Name          string `json:"name"`          // 变量名称
	Description   string `json:"description"`   // 变量描述
	VariableKey   string `json:"variableKey"`   // 变量键（用于环境变量/模板替换）
	VariableValue string `json:"variableValue"` // 变量值
}

// Skill 技能配置。
type Skill struct {
	Dir       string          `json:"dir"`       // skill 目录路径
	Variables []SkillVariable `json:"variables"` // 用户自定义变量
}

// MCP MCP 服务器配置。
type MCP struct {
	Name        string                  // MCP 名称
	URL         string                  // MCP SSE/STREAMABLE 服务器地址
	Description string                  // MCP 描述
	ApiAuth     *util.ApiAuthWebRequest // API 认证配置（可选）
	Headers     map[string]string       // 自定义请求头（可选）
}

// RunSession 执行会话标识。
type RunSession struct {
	ThreadID string // 对话会话 ID
	RunID    string // 执行请求 ID
}

// SandboxType 沙箱类型。
type SandboxType string

// RunnerType 运行器类型。
type RunnerType string

// SystemMessageStrategy system 消息处理策略。
type SystemMessageStrategy string

const (
	SystemMessageStrategyNone  SystemMessageStrategy = ""      // 不处理（默认）
	SystemMessageStrategyMerge SystemMessageStrategy = "merge" // 合并所有 system 消息到第一位
)

// ============================================================================
// 类型 - SandboxConfig
// ============================================================================

// SandboxConfig 沙箱配置。
type SandboxConfig struct {
	sandboxType SandboxType
	host        string // localhost 或容器名
	imageName   string // oneshot 模式用
}

func (c SandboxConfig) Type() SandboxType {
	return c.sandboxType
}

func (c SandboxConfig) Host() string {
	return c.host
}

func (c SandboxConfig) ImageName() string {
	return c.imageName
}

func (c SandboxConfig) APIEndpoint() string {
	return fmt.Sprintf("http://%s:%d", c.host, sandboxAPIPort)
}

func (c SandboxConfig) OpencodeEndpoint() string {
	return fmt.Sprintf("http://%s:%d", c.host, opencodeAPIPort)
}

func (c SandboxConfig) EinoEndpoint() string {
	return fmt.Sprintf("http://%s:%d", c.host, einoAPIPort)
}

// ============================================================================
// 类型 - Option/RunOption
// ============================================================================

// Option 选项接口。
type Option interface {
	apply(*RunOption) error
}

// OptionFunc 选项函数。
type OptionFunc func(*RunOption) error

func (f OptionFunc) apply(opts *RunOption) error {
	return f(opts)
}

// RunOption 运行选项。
type RunOption struct {
	RunSession                 RunSession
	ModelConfig                ModelConfig
	Sandbox                    SandboxConfig
	RunnerType                 RunnerType
	Instruction                string
	OverallTask                string
	InputDir                   string
	OutputDir                  string
	Skills                     []Skill
	Tools                      []Tool
	MCPs                       []MCP                 // MCP 服务器列表
	Messages                   []adk.Message         // 历史消息 + 当前问题（最后一条 User 消息）
	SystemMessageStrategy      SystemMessageStrategy // system 消息处理策略，默认不处理
	EnableThinking             bool
	EnableHumanInTheLoop       bool // 是否启用人机交互
	EnableHumanInTheLoopCustom bool // 是否允许用户自定义回答
	SkipCleanup                bool
	AgentName                  string
	TraceContext               map[string]string // W3C trace 传播头（traceparent, tracestate, baggage），用于跨进程 trace 传播
}

func (o *RunOption) Apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt.apply(o); err != nil {
			return err
		}
	}
	if o.RunSession.ThreadID == "" {
		o.RunSession.ThreadID = uuid.New().String()
	}
	if o.RunSession.RunID == "" {
		o.RunSession.RunID = uuid.New().String()
	}
	if o.Sandbox.Type() == "" {
		o.Sandbox.sandboxType = SandboxTypeReuse
	}
	if o.Sandbox.Host() == "" {
		return fmt.Errorf("sandbox requires host")
	}
	if o.Sandbox.Type() == SandboxTypeOneshot && o.Sandbox.ImageName() == "" {
		return fmt.Errorf("oneshot sandbox requires image name")
	}
	if len(o.Messages) == 0 {
		return fmt.Errorf("messages is empty")
	}
	lastMsg := o.Messages[len(o.Messages)-1]
	if lastMsg.Role != schema.User {
		return fmt.Errorf("last message must be user message, got %s", lastMsg.Role)
	}
	return nil
}

// ============================================================================
// 构造函数
// ============================================================================

// SandboxReuse 创建复用容器模式的沙箱配置。
func SandboxReuse(host string) SandboxConfig {
	return SandboxConfig{
		sandboxType: SandboxTypeReuse,
		host:        host,
	}
}

// SandboxOneshot 创建一次性容器模式的沙箱配置。
func SandboxOneshot(imageName string) SandboxConfig {
	return SandboxConfig{
		sandboxType: SandboxTypeOneshot,
		imageName:   imageName,
	}
}

// ============================================================================
// 选项函数
// ============================================================================

func WithModelConfig(config ModelConfig) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.ModelConfig = config
		return nil
	})
}

func WithRunSession(session RunSession) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.RunSession = session
		return nil
	})
}

func WithSandbox(cfg SandboxConfig) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.Sandbox = cfg
		return nil
	})
}

func WithRunnerType(runnerType RunnerType) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.RunnerType = runnerType
		return nil
	})
}

func WithInstruction(instruction string) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.Instruction = instruction
		return nil
	})
}

func WithOverallTask(overallTask string) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.OverallTask = overallTask
		return nil
	})
}

// WithInputDir 设置输入目录。
// 输入目录的内容会在执行前复制到沙箱工作目录。
// 支持 "/." 后缀：如 "/path/to/dir/." 表示复制目录内容而非目录本身。
func WithInputDir(inputDir string) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.InputDir = inputDir
		return nil
	})
}

// WithOutputDir 设置输出目录。
// 沙箱工作目录的内容会在执行后复制到该目录。
// 注意：隐藏文件（以 "." 开头）不会被复制。
func WithOutputDir(outputDir string) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.OutputDir = outputDir
		return nil
	})
}

// WithMessages 设置消息列表，最后一条消息必须是 User 消息。
func WithMessages(messages []adk.Message) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.Messages = append(opts.Messages, messages...)
		return nil
	})
}

func WithSystemMessageStrategy(strategy SystemMessageStrategy) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.SystemMessageStrategy = strategy
		return nil
	})
}

func WithSkills(skills []Skill) Option {
	return OptionFunc(func(opts *RunOption) error {
		for _, skill := range skills {
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
		}
		opts.Skills = skills
		return nil
	})
}

func WithMCPs(mcps []MCP) Option {
	return OptionFunc(func(opts *RunOption) error {
		for _, mcp := range mcps {
			if mcp.Name == "" {
				return fmt.Errorf("mcp name is required")
			}
			if mcp.URL == "" {
				return fmt.Errorf("mcp [%s] url is required", mcp.Name)
			}
		}
		opts.MCPs = mcps
		return nil
	})
}

func WithTools(tools []Tool) Option {
	return OptionFunc(func(opts *RunOption) error {
		ctx := context.Background()
		for i := range tools {
			if tools[i].OpenAPI3Schema == nil {
				return fmt.Errorf("tool schema is required")
			}
			if err := openapi3_util.ValidateDoc(ctx, tools[i].OpenAPI3Schema); err != nil {
				return fmt.Errorf("invalid tool schema: %w", err)
			}
			if tools[i].APIAuth != nil && tools[i].APIAuth.Type != "none" && tools[i].APIAuth.Value == "" {
				return fmt.Errorf("tool [%s] auth value is empty", tools[i].Name)
			}
			if tools[i].Name == "" {
				tools[i].Name = tools[i].OpenAPI3Schema.Info.Title
			}
			if len(tools[i].OperationIDs) > 0 {
				tools[i].OpenAPI3Schema = openapi3_util.FilterDocOperations(tools[i].OpenAPI3Schema, tools[i].OperationIDs)
			}
			opts.Tools = append(opts.Tools, tools[i])
		}
		return nil
	})
}

func WithEnableThinking(enable bool) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.EnableThinking = enable
		return nil
	})
}

func WithSkipCleanup(skip bool) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.SkipCleanup = skip
		return nil
	})
}

func WithAgentName(name string) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.AgentName = name
		return nil
	})
}

func WithEnableHumanInTheLoop(enable bool, enableCustom ...bool) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.EnableHumanInTheLoop = enable
		if len(enableCustom) > 0 {
			opts.EnableHumanInTheLoopCustom = enableCustom[0]
		}
		return nil
	})
}

// WithTraceContext 设置 W3C trace 传播头，用于跨进程（如 sandbox 容器）trace 上下文传播。
func WithTraceContext(headers map[string]string) Option {
	return OptionFunc(func(opts *RunOption) error {
		opts.TraceContext = headers
		return nil
	})
}
