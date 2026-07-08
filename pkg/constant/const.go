package constant

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
	AppTypeAgent     = "agent"     // 智能体
	AppTypeRag       = "rag"       // 文本问答
	AppTypeWorkflow  = "workflow"  // 工作流
	AppTypeChatflow  = "chatflow"  // 对话流
	AppTypeSkill     = "skill"     // Skill
	AppTypeMCPServer = "mcpserver" // mcp server
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
	AppStatisticSourceWeb     = "web"
	AppStatisticSourceOpenAPI = "openapi"
	AppStatisticSourceWebUrl  = "webURL"
	AppStatisticSourceDraft   = "draft" // 应用的草稿版本不统计
)

// biz module
const (
	BizModuleWGA               = "wga"       // 通用智能体
	BizModuleModel             = "model"     // 模型
	BizModuleResourceKnowledge = "knowledge" // 知识库
	BizModuleResourceMCP       = "mcp"       // MCP
	BizModuleResourceTool      = "tool"      // 插件工具
	BizModuleResourcePrompt    = "prompt"    // 提示词
	BizModuleResourceSkill     = "skill"     // Skills
	BizModuleResourceSafety    = "safety"    // 安全护栏
	BizModuleAppRag            = "rag"       // 知识问答
	BizModuleAppWorkflow       = "workflow"  // 工作流
	BizModuleAppAgent          = "agent"     // 智能体
)
