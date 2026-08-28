package mcp2skill

// MCPServerConfig holds connection info for an MCP server.
type MCPServerConfig struct {
	URL           string            // MCP server URL
	TransportType string            // "sse" or "streamable"
	Headers       map[string]string // HTTP headers to include in requests
}

// ConvertOptions configures the conversion process.
type ConvertOptions struct {
	// OutputDir is the directory to write skill files to.
	OutputDir string
	// SkillName overrides the auto-detected skill name.
	SkillName string
	// Description overrides the auto-generated description.
	Description string
	// ServerURL is included in SKILL.md for reference (key params stripped).
	ServerURL string
	// TransportType is included in SKILL.md for reference ("sse" or "streamable").
	TransportType string
	// ApiAuth is the API authentication configuration for credential display.
	ApiAuth *APIAuthConfig
}

// =============================================================================
// Intermediate Representation (IR)
// =============================================================================

// SkillDocument is the top-level IR produced by parsing MCP tools.
type SkillDocument struct {
	Meta       SkillMeta
	Tools      []ToolDocument
	ServerInfo ServerInfoDocument
}

// SkillMeta contains metadata about the skill.
type SkillMeta struct {
	Name        string
	Description string
	ToolCount   int
}

// ServerInfoDocument describes the MCP server connection.
type ServerInfoDocument struct {
	URL           string
	TransportType string
	AuthHeader    string // Description of header-based auth, e.g. "Authorization: Bearer <YOUR_TOKEN>"
}

// ToolDocument represents a single MCP tool in the IR.
type ToolDocument struct {
	Name         string
	Description  string
	Parameters   []ParameterDocument
	Required     []string
	OutputSchema *OutputSchemaDocument
	Annotations  *ToolAnnotationsDocument
}

// OutputSchemaDocument describes the output shape of a tool.
type OutputSchemaDocument struct {
	Type       string
	Properties []ParameterDocument
}

// ToolAnnotationsDocument represents tool annotations.
type ToolAnnotationsDocument struct {
	Title           string
	ReadOnlyHint    bool
	DestructiveHint bool
	IdempotentHint  bool
	OpenWorldHint   bool
}

// ParameterDocument represents a tool parameter.
type ParameterDocument struct {
	Name        string
	Type        string // comma-separated types for union types
	Description string
	Required    bool
	Enum        []string
	Default     string
	Properties  []ParameterDocument // for object type
	Items       *ParameterDocument  // for array type
}
