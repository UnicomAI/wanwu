// Package config 提供智能体配置的类型定义和加载功能。
package config

import (
	"context"
	"fmt"
	"os"
	"path"
	"slices"

	"github.com/UnicomAI/wanwu/pkg/log"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/viper"
)

// ============================================================================
// 类型 - 公开
// ============================================================================

// Agent 智能体配置。
type Agent struct {
	ID             string          `json:"id"`
	Type           AgentType       `json:"type"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Configure      AgentConfigure  `json:"configure"`
	Prompt         string          `json:"prompt"`
	ToolCategories []*ToolCategory `json:"tool_categories"`
	Skills         []Skill         `json:"skills"`
	SubAgents      []*Agent        `json:"sub_agents"`
}

// AgentConfigure 智能体配置项。
type AgentConfigure struct {
	MaxIterations  int            `json:"max_iterations" mapstructure:"max_iterations"`
	EnableThinking bool           `json:"enable_thinking" mapstructure:"enable_thinking"`
	Sandbox        *SandboxConfig `json:"sandbox" mapstructure:"sandbox"`
}

// SandboxConfig 沙箱配置。
type SandboxConfig struct {
	Type      string `json:"type" mapstructure:"type"`
	Host      string `json:"host" mapstructure:"host"`
	ImageName string `json:"image_name" mapstructure:"image_name"`
}

// SkillVariable 技能变量配置。
type SkillVariable struct {
	Name          string `json:"name" mapstructure:"name"`                   // 变量名称
	Description   string `json:"description" mapstructure:"description"`     // 变量描述
	VariableKey   string `json:"variableKey" mapstructure:"variableKey"`     // 变量键
	VariableValue string `json:"variableValue" mapstructure:"variableValue"` // 变量值
}

// Skill 技能配置。
type Skill struct {
	Dir       string          `json:"dir"`       // skill 目录路径（相对程序运行目录）
	Variables []SkillVariable `json:"variables"` // 用户自定义变量
}

// ToolCategory 工具类别配置。
type ToolCategory struct {
	Category  ToolCategoryType      `json:"category"`
	Condition ToolCategoryCondition `json:"condition"`
	Tools     []*Tool               `json:"tools"`
}

// Tool 工具配置。
type Tool struct {
	Doc          *openapi3.T     `json:"-"`             // OpenAPI schema
	SchemaPath   string          `json:"-"`             // schema 文件路径（相对程序运行目录）
	AuthRequired bool            `json:"auth_required"` // 是否需要认证
	Operations   []ToolOperation `json:"operations"`    // 允许的操作
}

// ToolOperation 工具操作配置。
type ToolOperation struct {
	OperationID    string `json:"operation_id" mapstructure:"operation_id"`
	ReturnDirectly bool   `json:"return_directly" mapstructure:"return_directly"`
}

// ============================================================================
// 类型 - 私有（配置解析用）
// ============================================================================

type all struct {
	Agents []agentTemplate `json:"agents" mapstructure:"agents"`
}

type agentTemplate struct {
	RelativePath string `json:"relative_path" mapstructure:"relative_path"`
}

type agentConfig struct {
	ID                 string          `json:"id" mapstructure:"id"`
	Type               AgentType       `json:"type" mapstructure:"type"`
	Name               string          `json:"name" mapstructure:"name"`
	Description        string          `json:"description" mapstructure:"description"`
	Configure          AgentConfigure  `json:"configure" mapstructure:"configure"`
	PromptRelativePath string          `json:"prompt_relative_path" mapstructure:"prompt_relative_path"`
	ToolCategories     []toolCategory  `json:"tool_categories" mapstructure:"tool_categories"`
	Skills             []skillConfig   `json:"skills" mapstructure:"skills"`
	SubAgents          []agentTemplate `json:"sub_agents" mapstructure:"sub_agents"`
}

type skillConfig struct {
	Dir       string          `json:"dir" mapstructure:"dir"`
	Variables []SkillVariable `json:"variables" mapstructure:"variables"`
}

type toolCategory struct {
	Category  ToolCategoryType      `json:"category" mapstructure:"category"`
	Condition ToolCategoryCondition `json:"condition" mapstructure:"condition"`
	Tools     []toolTemplate        `json:"tools" mapstructure:"tools"`
}

type toolTemplate struct {
	Path         string          `json:"path" mapstructure:"path"`
	AuthRequired bool            `json:"auth_required" mapstructure:"auth_required"`
	Operations   []ToolOperation `json:"operations" mapstructure:"operations"`
}

// ============================================================================
// 公开函数
// ============================================================================

// LoadAgents 从配置文件加载智能体配置。
// configPath 为配置文件路径，支持 YAML 格式。
func LoadAgents(ctx context.Context, configPath string) ([]*Agent, error) {
	cfg := &all{}
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	baseDir := path.Dir(configPath)
	var agents []*Agent
	for _, at := range cfg.Agents {
		agent, err := at.load(ctx, baseDir, "")
		if err != nil {
			return nil, err
		}
		for _, a := range agents {
			if a.ID == agent.ID {
				return nil, fmt.Errorf("load agent [%v(%v)] already exist", agent.ID, agent.Type)
			}
		}
		agents = append(agents, agent)
	}
	return agents, nil
}

// ============================================================================
// 私有方法
// ============================================================================

func (at *agentTemplate) load(ctx context.Context, baseDir, classPrefix string) (*Agent, error) {
	configPath := path.Join(baseDir, at.RelativePath)
	cfg := &agentConfig{}
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("load agent (%v) err: %v", configPath, err)
	}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal agent (%v) err: %v", configPath, err)
	}
	if cfg.ID == "" {
		return nil, fmt.Errorf("load agent (%v) id empty", configPath)
	}
	if cfg.Name == "" {
		return nil, fmt.Errorf("load agent (%v) name empty", configPath)
	}

	agent := &Agent{
		ID:          cfg.ID,
		Type:        cfg.Type,
		Name:        cfg.Name,
		Description: cfg.Description,
		Configure:   cfg.Configure,
	}
	log.Debugf("[wga][config] %vload agent [%v(%v)], %v", classPrefix, agent.ID, agent.Type, configPath)

	if err := at.loadPrompt(cfg, configPath, agent); err != nil {
		return nil, err
	}
	if err := at.loadTools(ctx, cfg, configPath, classPrefix, agent); err != nil {
		return nil, err
	}
	at.loadSkills(cfg, agent)
	if err := at.loadSubAgents(ctx, cfg, configPath, classPrefix, agent); err != nil {
		return nil, err
	}

	return agent, nil
}

func (at *agentTemplate) loadPrompt(cfg *agentConfig, configPath string, agent *Agent) error {
	if cfg.PromptRelativePath == "" {
		return nil
	}
	promptPath := path.Join(path.Dir(configPath), cfg.PromptRelativePath)
	b, err := os.ReadFile(promptPath)
	if err != nil {
		return fmt.Errorf("load agent (%v) read prompt (%v) err: %v", configPath, promptPath, err)
	}
	agent.Prompt = string(b)
	return nil
}

func (at *agentTemplate) loadTools(ctx context.Context, cfg *agentConfig, configPath, classPrefix string, agent *Agent) error {
	var tools []string
	var categories []string
	var operationIDs []string

	for _, tc := range cfg.ToolCategories {
		if slices.Contains(categories, string(tc.Category)) {
			return fmt.Errorf("load agent (%v), tool category (%v) already exist", configPath, tc.Category)
		}
		categories = append(categories, string(tc.Category))

		category := &ToolCategory{
			Category:  tc.Category,
			Condition: tc.Condition,
		}
		for _, tt := range tc.Tools {
			tool, err := tt.load(ctx, classPrefix+"  ")
			if err != nil {
				return fmt.Errorf("load agent (%v) err: %v", configPath, err)
			}
			if slices.Contains(tools, tool.Doc.Info.Title) {
				return fmt.Errorf("load agent (%v), tool (%v) already exist", configPath, tool.Doc.Info.Title)
			}
			tools = append(tools, tool.Doc.Info.Title)

			for _, op := range tool.Operations {
				if slices.Contains(operationIDs, op.OperationID) {
					return fmt.Errorf("load agent (%v), tool operation (%v) already exist", configPath, op.OperationID)
				}
				operationIDs = append(operationIDs, op.OperationID)
			}
			category.Tools = append(category.Tools, tool)
		}
		agent.ToolCategories = append(agent.ToolCategories, category)
	}
	return nil
}

func (at *agentTemplate) loadSkills(cfg *agentConfig, agent *Agent) {
	for _, sc := range cfg.Skills {
		agent.Skills = append(agent.Skills, Skill(sc))
	}
}

func (at *agentTemplate) loadSubAgents(ctx context.Context, cfg *agentConfig, configPath, classPrefix string, agent *Agent) error {
	for _, subAt := range cfg.SubAgents {
		subAgent, err := subAt.load(ctx, path.Dir(configPath), classPrefix+"  ")
		if err != nil {
			return fmt.Errorf("load agent (%v) err: %v", configPath, err)
		}
		for _, sa := range agent.SubAgents {
			if sa.ID == subAgent.ID {
				return fmt.Errorf("load agent (%v), sub agent (%v) already exist", configPath, subAgent.ID)
			}
		}
		agent.SubAgents = append(agent.SubAgents, subAgent)
	}
	return nil
}

func (tt *toolTemplate) load(ctx context.Context, classPrefix string) (*Tool, error) {
	schema, err := os.ReadFile(tt.Path)
	if err != nil {
		return nil, err
	}
	doc, err := openapi3_util.LoadFromData(ctx, schema)
	if err != nil {
		return nil, fmt.Errorf("load tool (%v) err: %v", tt.Path, err)
	}
	if len(tt.Operations) == 0 {
		return nil, fmt.Errorf("load tool (%v) operations empty", tt.Path)
	}

	var operationIDs []string
	for _, op := range tt.Operations {
		if slices.Contains(operationIDs, op.OperationID) {
			return nil, fmt.Errorf("load tool (%v) operation (%v) duplicate", tt.Path, op.OperationID)
		}
		operationIDs = append(operationIDs, op.OperationID)

		var exist bool
		for _, pathItem := range doc.Paths {
			for _, operation := range pathItem.Operations() {
				if operation.OperationID == op.OperationID {
					exist = true
					break
				}
			}
			if exist {
				break
			}
		}
		if !exist {
			return nil, fmt.Errorf("load tool (%v) operation (%v) not exist", tt.Path, op.OperationID)
		}
	}

	log.Debugf("[wga][config] %vload tool %v, %v", classPrefix, tt.Operations, tt.Path)
	return &Tool{
		Doc:          doc,
		SchemaPath:   tt.Path,
		AuthRequired: tt.AuthRequired,
		Operations:   tt.Operations,
	}, nil
}
