package openapi2skill

// GroupByStrategy defines how operations are grouped into resources.
type GroupByStrategy string

const (
	GroupByTags GroupByStrategy = "tags"
	GroupByPath GroupByStrategy = "path"
	GroupByAuto GroupByStrategy = "auto"
)

// CaseStrategy defines how to handle case-insensitive filesystem collisions.
type CaseStrategy string

const (
	CaseStrategyLowercase CaseStrategy = "lowercase"
)

// ParserOptions configures the parsing behavior.
type ParserOptions struct {
	// SkillName overrides the generated skill name.
	SkillName string
	// Filter controls which operations are included.
	Filter *ParserFilter
	// GroupBy controls how operations are grouped into resources.
	// Defaults to "auto" (use tags if available, otherwise path).
	GroupBy GroupByStrategy
}

// ParserFilter controls which operations are included in the output.
type ParserFilter struct {
	// IncludeTags only includes operations with these tags.
	IncludeTags []string
	// ExcludeTags excludes operations with these tags.
	ExcludeTags []string
	// ExcludeDeprecated excludes deprecated operations.
	ExcludeDeprecated bool
	// ExcludePaths excludes paths matching these prefixes.
	ExcludePaths []string
}

// ConvertOptions configures the conversion process.
type ConvertOptions struct {
	// OutputDir is the directory to write skill files to.
	OutputDir string
	// Parser options.
	Parser ParserOptions
	// CaseStrategy for handling case-insensitive filesystem collisions.
	CaseStrategy CaseStrategy
}

// =============================================================================
// Intermediate Representation (IR)
// =============================================================================

// SkillDocument is the top-level IR produced by parsing an OpenAPI spec.
type SkillDocument struct {
	Meta         SkillMeta
	Resources    []ResourceDocument
	SchemaGroups []SchemaGroupDocument
	AuthSchemes  []AuthSchemeDocument
}

// SkillMeta contains metadata about the API.
type SkillMeta struct {
	Name            string
	Title           string
	Description     string
	Version         string
	OpenAPIVersion  string
	License         *LicenseDocument
	Contact         string
	Servers         []ServerDocument
	SecuritySchemes []string
}

// LicenseDocument represents API license info.
type LicenseDocument struct {
	Name string
	URL  string
}

// ServerDocument represents an API server.
type ServerDocument struct {
	URL         string
	Description string
}

// ResourceDocument groups operations by tag or path prefix.
type ResourceDocument struct {
	Tag         string
	Description string
	Operations  []OperationDocument
}

// OperationDocument represents a single API operation.
type OperationDocument struct {
	OperationID string
	Path        string
	Method      string
	Tag         string
	Summary     string
	Description string
	Deprecated  bool
	Parameters  []ParameterDocument
	RequestBody *RequestBodyDocument
	Responses   []ResponseDocument
	Security    []SecurityRequirementDocument
}

// ParameterDocument represents an operation parameter.
type ParameterDocument struct {
	Name        string
	In          string // query, header, path, cookie
	Type        string
	Required    bool
	Description string
	Schema      *SchemaRefDocument
}

// RequestBodyDocument represents a request body.
type RequestBodyDocument struct {
	Description  string
	Required     bool
	ContentTypes []string
	Schema       *SchemaRefDocument
}

// ResponseDocument represents an operation response.
type ResponseDocument struct {
	Status      string
	Description string
	Schema      *SchemaRefDocument
}

// SecurityRequirementDocument represents a security requirement.
type SecurityRequirementDocument struct {
	Name   string
	Scopes []string
}

// SchemaGroupDocument groups schemas by naming prefix.
type SchemaGroupDocument struct {
	Prefix  string
	Schemas []SchemaDocument
}

// SchemaDocumentType defines the type of a schema.
type SchemaDocumentType string

const (
	SchemaTypeObject    SchemaDocumentType = "object"
	SchemaTypeArray     SchemaDocumentType = "array"
	SchemaTypeEnum      SchemaDocumentType = "enum"
	SchemaTypePrimitive SchemaDocumentType = "primitive"
	SchemaTypeAllOf     SchemaDocumentType = "allOf"
	SchemaTypeOneOf     SchemaDocumentType = "oneOf"
	SchemaTypeAnyOf     SchemaDocumentType = "anyOf"
)

// SchemaDocument represents a schema definition.
type SchemaDocument struct {
	Name        string
	Type        SchemaDocumentType
	Description string
	Fields      []FieldDocument
	EnumValues  []interface{}
	Composition []SchemaRefDocument
	Items       *SchemaRefDocument
}

// FieldDocument represents a field in an object schema.
type FieldDocument struct {
	Name         string
	Type         string
	Required     bool
	Description  string
	Schema       *SchemaRefDocument
	NestedFields []FieldDocument
}

// SchemaRefDocument represents a reference to a schema or an inline schema.
type SchemaRefDocument struct {
	Ref    string
	Inline *SchemaDocument
}

// AuthSchemeDocument represents an authentication scheme.
type AuthSchemeDocument struct {
	Name             string
	Type             string
	Description      string
	In               string // for apiKey: header, query, cookie
	APIKeyName       string // for apiKey
	Scheme           string // for http
	BearerFormat     string // for http
	Flows            []OAuthFlowDocument
	OpenIDConnectURL string // for openIdConnect
}

// OAuthFlowDocument represents an OAuth flow.
type OAuthFlowDocument struct {
	Name             string
	AuthorizationURL string
	TokenURL         string
	Scopes           map[string]string
}
