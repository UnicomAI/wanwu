package constant

// app type
const (
	AppTypeAgent     = "agent"     // 智能体 [EN] agent
	AppTypeRag       = "rag"       // 文本问答 [EN] Text Q&A
	AppTypeWorkflow  = "workflow"  // 工作流 [EN] Workflow
	AppTypeChatflow  = "chatflow"  // 对话流 [EN] conversation flow
	AppTypeMCPServer = "mcpserver" // mcp server
)

// app publish type
const (
	AppPublishPublic       = "public"       // 系统公开发布 [EN] System public release
	AppPublishOrganization = "organization" // 组织公开发布 [EN] Organization publishes publicly
	AppPublishPrivate      = "private"      // 私密发布 [EN] Post privately
)

// tool type
const (
	ToolTypeBuiltIn = "builtin" // 内置工具 [EN] Built-in tools
	ToolTypeCustom  = "custom"  // 自定义工具 [EN] Custom tools
)

// mcp type
const (
	MCPTypeMCP       = "mcp"       // mcp
	MCPTypeMCPServer = "mcpserver" // mcp server
)

// mcp server tool type
const (
	MCPServerToolTypeCustomTool  = "custom"  // 自定义工具 [EN] Custom tools
	MCPServerToolTypeBuiltInTool = "builtin" // 内置工具 [EN] Built-in tools
	MCPServerToolTypeOpenAPI     = "openapi" // 用户导入的openapi [EN] openapi imported by user
)
