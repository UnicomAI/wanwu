package constant

// app type
const (
	AppTypeAgent     = "agent"     // agent
	AppTypeRag       = "rag"       // Text Q&A
	AppTypeWorkflow  = "workflow"  // Workflow
	AppTypeChatflow  = "chatflow"  // conversation flow
	AppTypeMCPServer = "mcpserver" // mcp server
)

// app publish type
const (
	AppPublishPublic       = "public"       // System public release
	AppPublishOrganization = "organization" // Organization publishes publicly
	AppPublishPrivate      = "private"      // Post privately
)

// tool type
const (
	ToolTypeBuiltIn = "builtin" // Built-in tools
	ToolTypeCustom  = "custom"  // Custom tools
)

// mcp type
const (
	MCPTypeMCP       = "mcp"       // mcp
	MCPTypeMCPServer = "mcpserver" // mcp server
)

// mcp server tool type
const (
	MCPServerToolTypeCustomTool  = "custom"  // Custom tools
	MCPServerToolTypeBuiltInTool = "builtin" // Built-in tools
	MCPServerToolTypeOpenAPI     = "openapi" // openapi imported by user
)
