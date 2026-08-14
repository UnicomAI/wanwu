package constant

import (
	"unicode"
	"unicode/utf8"
)

// openapi type
const (
	OpenAPITypeChatflow  = "chatflow"  // 对话问答
	OpenAPITypeWorkflow  = "workflow"  // 工作流
	OpenAPITypeAgent     = "agent"     // 智能体
	OpenAPITypeRag       = "rag"       // 文本问答
	OpenAPITypeKnowledge = "knowledge" // 知识库
	OpenAPITypeModel     = "model"     // 可用模型列表查询
	OpenAPITypeWGA       = "wga"       // 通用智能体
)

// app type
const (
	AppTypeAgent           = "agent"           // 智能体
	AppTypeRag             = "rag"             // 文本问答
	AppTypeWorkflow        = "workflow"        // 工作流
	AppTypeChatflow        = "chatflow"        // 对话流
	AppTypeSkill           = "skill"           // Skill
	AppTypeMCPServer       = "mcpserver"       // mcp server
	AppTypeDigitalEmployee = "digitalemployee" // 数字员工
)

// app publish type
const (
	AppPublishPublic       = "public"       // 系统公开发布
	AppPublishOrganization = "organization" // 组织公开发布
	AppPublishPrivate      = "private"      // 私密发布
)

// tool type
const (
	ToolTypeBuiltIn = "builtin" // 内置工具
	ToolTypeCustom  = "custom"  // 自定义工具
)

// mcp type
const (
	MCPTypeMCP       = "mcp"       // mcp
	MCPTypeMCPServer = "mcpserver" // mcp server
)

// mcp server tool type
const (
	MCPServerToolTypeCustomTool  = "custom"  // 自定义工具
	MCPServerToolTypeBuiltInTool = "builtin" // 内置工具
	MCPServerToolTypeOpenAPI     = "openapi" // 用户导入的openapi
)

// mcp transport type
const (
	MCPTransportSSE        = "sse"
	MCPTransportStreamable = "streamable"
)

// agent category
const (
	AgentCategorySingle = 1
	AgentCategoryMulti  = 2
)

// conversation type
const (
	ConversationTypeWebURL    = "openurl"   // openurl
	ConversationTypePublished = "published" // 已发布
	ConversationTypeDraft     = "draft"     // 草稿
	ConversationTypeOpenAPI   = "openapi"   // openapi
)

// skill type
const (
	SkillTypeBuiltIn  = "builtin"  // 内置技能
	SkillTypeCustom   = "custom"   // 自定义技能
	SkillTypeAcquired = "acquired" // 添加的技能
)

// safety type
const (
	SensitiveTableTypeGlobal   = "global"   // 全局敏感词表
	SensitiveTableTypePersonal = "personal" // 个人敏感词表
)

// knowledge type
const (
	KnowledgeBase       = 0 // 文本知识库
	KnowledgeQA         = 1 // 问答库
	KnowledgeMultiModal = 2 // 多模态知识库
)

// app statistic source
const (
	BizSourceWeb     = "web"
	BizSourceOpenAPI = "openapi"
	BizSourceWebUrl  = "webURL"
)

// biz module
const (
	BizModuleWGA                = "wga"             // 通用智能体
	BizModuleModel              = "model"           // 模型
	BizModuleResourceKnowledge  = "knowledge"       // 知识库
	BizModuleResourceMCP        = "mcp"             // MCP
	BizModuleResourceTool       = "tool"            // 插件工具
	BizModuleResourcePrompt     = "prompt"          // 提示词
	BizModuleResourceSkill      = "skill"           // Skills
	BizModuleResourceSafety     = "safety"          // 安全护栏
	BizModuleAppRag             = "rag"             // 知识问答
	BizModuleAppWorkflow        = "workflow"        // 工作流
	BizModuleAppAgent           = "agent"           // 智能体
	BizModuleAppDigitalEmployee = "digitalemployee" // 数字员工
)

// StatisticModuleAllowsEmptyAppID 应用统计 V2 写入/查询：这些板块允许 appId 为空（板块级维度）。
func StatisticModuleAllowsEmptyAppID(module string) bool {
	switch module {
	case BizModuleWGA, BizModuleModel, BizModuleResourceSkill, BizModuleResourceKnowledge, BizModuleResourcePrompt:
		return true
	default:
		return false
	}
}

// BizModuleName 返回业务板块中文展示名；未知 code 原样返回。
func BizModuleName(module string) string {
	switch module {
	case BizModuleWGA:
		return "通用智能体"
	case BizModuleModel:
		return "模型体验"
	case BizModuleResourceKnowledge:
		return "知识库"
	case BizModuleResourceMCP:
		return "MCP"
	case BizModuleResourceTool:
		return "插件工具"
	case BizModuleResourcePrompt:
		return "提示词"
	case BizModuleResourceSkill:
		return "Skill"
	case BizModuleResourceSafety:
		return "安全护栏"
	case BizModuleAppRag:
		return "知识问答"
	case BizModuleAppWorkflow:
		return "工作流"
	case BizModuleAppAgent:
		return "智能体"
	case BizModuleAppDigitalEmployee:
		return "数字员工"
	default:
		return module
	}
}

// BizSourceName 返回调用来源展示名（首字母大写，其余保持原样）。
// web→Web、openapi→Openapi、webURL→WebURL；未知 code 原样返回。
func BizSourceName(source string) string {
	if source == "" {
		return source
	}
	r, size := utf8.DecodeRuneInString(source)
	if r == utf8.RuneError && size == 1 {
		return source
	}
	return string(unicode.ToUpper(r)) + source[size:]
}
