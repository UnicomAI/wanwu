package openapi2skill

import (
	"fmt"
	"regexp"
	"strings"
)

// nonAlnumRe matches sequences of non-letter, non-number, non-hyphen characters.
var nonAlnumRe = regexp.MustCompile(`[^\p{L}\p{N}-]+`)
var multiDashRe = regexp.MustCompile(`-{2,}`)
var trimDashRe = regexp.MustCompile(`^-+|-+$`)

// toFileName converts a name to a safe filename by replacing non-alphanumeric
// characters with hyphens.
func toFileName(name string) string {
	sanitized := nonAlnumRe.ReplaceAllString(name, "-")
	sanitized = multiDashRe.ReplaceAllString(sanitized, "-")
	sanitized = trimDashRe.ReplaceAllString(sanitized, "")
	if sanitized == "" {
		return "unnamed"
	}
	return strings.ToLower(sanitized)
}

// =============================================================================
// Renderer
// =============================================================================

// Renderer converts SkillDocument IR to markdown strings.
type Renderer struct{}

// NewRenderer creates a new Renderer.
func NewRenderer() *Renderer {
	return &Renderer{}
}

// RenderSkill renders the top-level SKILL.md content.
func (r *Renderer) RenderSkill(doc SkillDocument) string {
	var b strings.Builder

	totalOps := 0
	for _, res := range doc.Resources {
		totalOps += len(res.Operations)
	}
	totalSchemas := 0
	for _, g := range doc.SchemaGroups {
		totalSchemas += len(g.Schemas)
	}

	// Frontmatter
	b.WriteString("---\n")
	fmt.Fprintf(&b, "name: %s\n", doc.Meta.Name)
	desc := doc.Meta.Description
	if desc != "" {
		fmt.Fprintf(&b, "description: %s. Use when working with the %s or when the user needs to interact with this API.\n", desc, doc.Meta.Title)
	} else {
		fmt.Fprintf(&b, "description: Use when working with the %s or when the user needs to interact with this API.\n", doc.Meta.Title)
	}
	if doc.Meta.License != nil {
		licStr := doc.Meta.License.Name
		if doc.Meta.License.URL != "" {
			licStr += fmt.Sprintf(" (%s)", doc.Meta.License.URL)
		}
		fmt.Fprintf(&b, "license: %s\n", licStr)
	}
	b.WriteString("metadata:\n")
	fmt.Fprintf(&b, "  api-version: \"%s\"\n", doc.Meta.Version)
	fmt.Fprintf(&b, "  openapi-version: \"%s\"\n", doc.Meta.OpenAPIVersion)
	if doc.Meta.Contact != "" {
		fmt.Fprintf(&b, "  contact: \"%s\"\n", doc.Meta.Contact)
	}
	b.WriteString("---\n\n")

	// Title
	fmt.Fprintf(&b, "# %s\n\n", doc.Meta.Title)

	if doc.Meta.Description != "" {
		b.WriteString(doc.Meta.Description + "\n\n")
	}

	// How to Use
	b.WriteString("## How to Use This Skill\n\n")
	b.WriteString("This API documentation is split into multiple files for on-demand loading.\n\n")
	b.WriteString("**Directory structure:**\n```\nreferences/\n")
	fmt.Fprintf(&b, "├── resources/      # %d resource index files\n", len(doc.Resources))
	fmt.Fprintf(&b, "├── operations/     # %d operation detail files\n", totalOps)
	fmt.Fprintf(&b, "└── schemas/        # %d schema groups, %d schema files\n```\n\n", len(doc.SchemaGroups), totalSchemas)
	b.WriteString("**Navigation flow:**\n")
	b.WriteString("1. Find the resource you need in the list below\n")
	b.WriteString("2. Read `references/resources/<resource>.md` to see available operations\n")
	b.WriteString("3. Read `references/operations/<operation>.md` for full details\n")
	b.WriteString("4. If an operation references a schema, read `references/schemas/<prefix>/<schema>.md`\n")

	// Base URL
	if len(doc.Meta.Servers) > 0 {
		b.WriteString("\n## Base URL\n\n")
		for _, server := range doc.Meta.Servers {
			line := fmt.Sprintf("- `%s`", server.URL)
			if server.Description != "" {
				line += fmt.Sprintf(" - %s", server.Description)
			}
			b.WriteString(line + "\n")
		}
	}

	// Authentication
	if len(doc.Meta.SecuritySchemes) > 0 {
		b.WriteString("\n## Authentication\n\nSupported methods: ")
		schemeStrs := make([]string, len(doc.Meta.SecuritySchemes))
		for i, s := range doc.Meta.SecuritySchemes {
			schemeStrs[i] = fmt.Sprintf("**%s**", s)
		}
		b.WriteString(strings.Join(schemeStrs, ", "))
		b.WriteString(". See `references/authentication.md` for details.\n")
	}

	// Resources
	b.WriteString("\n## Resources\n\n")
	for _, resource := range doc.Resources {
		resDesc := ""
		if resource.Description != "" {
			d := resource.Description
			if len(d) > 50 {
				d = d[:50]
			}
			resDesc = fmt.Sprintf(" - %s", d)
		}
		fmt.Fprintf(&b, "- **%s** → `references/resources/%s.md` (%d ops)%s\n",
			resource.Tag, toFileName(resource.Tag), len(resource.Operations), resDesc)
	}

	return b.String()
}

// RenderResource renders a resource index markdown.
func (r *Renderer) RenderResource(doc ResourceDocument) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", doc.Tag)

	if doc.Description != "" {
		b.WriteString(doc.Description + "\n\n")
	}

	b.WriteString("## Operations\n\n")
	b.WriteString("| Method | Path | Summary | Details |\n")
	b.WriteString("|--------|------|---------|----------|\n")

	for _, op := range doc.Operations {
		summary := op.Summary
		fmt.Fprintf(&b, "| %s | `%s` | %s | [View](../operations/%s.md) |\n",
			op.Method, op.Path, summary, toFileName(op.OperationID))
	}

	return b.String()
}

// RenderOperation renders a single operation detail markdown.
func (r *Renderer) RenderOperation(doc OperationDocument) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s %s\n\n", doc.Method, doc.Path)
	fmt.Fprintf(&b, "**Resource:** [%s](../resources/%s.md)\n", doc.Tag, toFileName(doc.Tag))

	if doc.Summary != "" {
		fmt.Fprintf(&b, "\n**%s**\n", doc.Summary)
	}
	fmt.Fprintf(&b, "\n**Operation ID:** `%s`\n", doc.OperationID)

	if doc.Deprecated {
		b.WriteString("\n⚠️ **Deprecated**\n")
	}

	if doc.Description != "" && doc.Description != doc.Summary {
		fmt.Fprintf(&b, "\n%s\n", doc.Description)
	}

	// Parameters
	if len(doc.Parameters) > 0 {
		b.WriteString("\n## Parameters\n\n")
		b.WriteString("| Name | In | Type | Required | Description |\n")
		b.WriteString("|------|------|------|----------|-------------|\n")
		for _, param := range doc.Parameters {
			reqStr := "No"
			if param.Required {
				reqStr = "Yes"
			}
			fmt.Fprintf(&b, "| `%s` | %s | %s | %s | %s |\n",
				param.Name, param.In, param.Type, reqStr, param.Description)
		}
	}

	// Request Body
	if doc.RequestBody != nil {
		b.WriteString("\n## Request Body\n\n")
		if doc.RequestBody.Description != "" {
			b.WriteString(doc.RequestBody.Description + "\n\n")
		}
		if doc.RequestBody.Required {
			b.WriteString("**Required:** Yes\n\n")
		}
		ctStrs := make([]string, len(doc.RequestBody.ContentTypes))
		for i, ct := range doc.RequestBody.ContentTypes {
			ctStrs[i] = fmt.Sprintf("`%s`", ct)
		}
		fmt.Fprintf(&b, "**Content Types:** %s\n\n", strings.Join(ctStrs, ", "))

		if doc.RequestBody.Schema != nil && doc.RequestBody.Schema.Ref != "" {
			refName := strings.TrimSuffix(doc.RequestBody.Schema.Ref, "[]")
			prefix := extractSchemaPrefix(refName)
			isArray := strings.HasSuffix(doc.RequestBody.Schema.Ref, "[]")
			arrPrefix := ""
			if isArray {
				arrPrefix = "Array of "
			}
			fmt.Fprintf(&b, "**Schema:** %s[%s](../schemas/%s/%s.md)\n\n",
				arrPrefix, refName, toFileName(prefix), toFileName(refName))
		} else if doc.RequestBody.Schema != nil && doc.RequestBody.Schema.Inline != nil {
			inline := doc.RequestBody.Schema.Inline
			fmt.Fprintf(&b, "**Schema:** %s\n\n", inline.Name)
			if inline.Type == SchemaTypeObject && len(inline.Fields) > 0 {
				b.WriteString("| Field | Type | Required | Description |\n")
				b.WriteString("|-------|------|----------|-------------|\n")
				for _, field := range inline.Fields {
					typ := field.Type
					if field.Schema != nil && field.Schema.Ref != "" {
						p := extractSchemaPrefix(field.Schema.Ref)
						typ = fmt.Sprintf("[%s](../schemas/%s/%s.md)", field.Schema.Ref, toFileName(p), toFileName(field.Schema.Ref))
					}
					reqStr := "No"
					if field.Required {
						reqStr = "Yes"
					}
					fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", field.Name, typ, reqStr, field.Description)
				}
				b.WriteString("\n")
			}
		}
	}

	// Responses
	if len(doc.Responses) > 0 {
		b.WriteString("\n## Responses\n\n")
		b.WriteString("| Status | Description |\n")
		b.WriteString("|--------|-------------|\n")
		for _, res := range doc.Responses {
			fmt.Fprintf(&b, "| %s | %s |\n", res.Status, res.Description)
		}

		for _, res := range doc.Responses {
			if (res.Status == "200" || res.Status == "201") && res.Schema != nil {
				if res.Schema.Ref != "" {
					refName := strings.TrimSuffix(res.Schema.Ref, "[]")
					prefix := extractSchemaPrefix(refName)
					isArray := strings.HasSuffix(res.Schema.Ref, "[]")
					arrPrefix := ""
					if isArray {
						arrPrefix = "Array of "
					}
					fmt.Fprintf(&b, "\n**Success Response Schema:**\n\n%s[%s](../schemas/%s/%s.md)\n",
						arrPrefix, refName, toFileName(prefix), toFileName(refName))
				} else if res.Schema.Inline != nil {
					inline := res.Schema.Inline
					fmt.Fprintf(&b, "\n**Success Response Schema:** %s\n\n", inline.Name)
					if inline.Type == SchemaTypeObject && len(inline.Fields) > 0 {
						b.WriteString("| Field | Type | Required | Description |\n")
						b.WriteString("|-------|------|----------|-------------|\n")
						for _, field := range inline.Fields {
							typ := field.Type
							if field.Schema != nil && field.Schema.Ref != "" {
								p := extractSchemaPrefix(field.Schema.Ref)
								typ = fmt.Sprintf("[%s](../schemas/%s/%s.md)", field.Schema.Ref, toFileName(p), toFileName(field.Schema.Ref))
							}
							reqStr := "No"
							if field.Required {
								reqStr = "Yes"
							}
							fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", field.Name, typ, reqStr, field.Description)
						}
					}
				}
				break
			}
		}
	}

	// Security
	if len(doc.Security) > 0 {
		b.WriteString("\n## Security\n\n")
		for _, sec := range doc.Security {
			line := fmt.Sprintf("- **%s**", sec.Name)
			if len(sec.Scopes) > 0 {
				line += fmt.Sprintf(": %s", strings.Join(sec.Scopes, ", "))
			}
			b.WriteString(line + "\n")
		}
	}

	return b.String()
}

// RenderSchema renders a single schema detail markdown.
func (r *Renderer) RenderSchema(doc SchemaDocument) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", doc.Name)

	if doc.Description != "" {
		b.WriteString(doc.Description + "\n\n")
	}
	fmt.Fprintf(&b, "**Type:** %s\n\n", doc.Type)

	switch doc.Type {
	case SchemaTypeObject:
		if len(doc.Fields) > 0 {
			b.WriteString("## Fields\n\n")
			b.WriteString("| Field | Type | Required | Description |\n")
			b.WriteString("|-------|------|----------|-------------|\n")
			for _, field := range doc.Fields {
				typ := field.Type
				if field.Schema != nil && field.Schema.Ref != "" {
					typ = fmt.Sprintf("[%s](%s.md)", field.Schema.Ref, toFileName(field.Schema.Ref))
				}
				reqStr := "No"
				if field.Required {
					reqStr = "Yes"
				}
				fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", field.Name, typ, reqStr, field.Description)
			}

			// Nested Fields
			var nestedFields []FieldDocument
			for _, f := range doc.Fields {
				if len(f.NestedFields) > 0 {
					nestedFields = append(nestedFields, f)
				}
			}
			if len(nestedFields) > 0 {
				b.WriteString("\n## Nested Fields\n\n")
				for _, field := range nestedFields {
					fmt.Fprintf(&b, "### `%s`\n\n", field.Name)
					b.WriteString("| Field | Type | Required | Description |\n")
					b.WriteString("|-------|------|----------|-------------|\n")
					for _, nested := range field.NestedFields {
						typ := nested.Type
						if nested.Schema != nil && nested.Schema.Ref != "" {
							typ = fmt.Sprintf("[%s](%s.md)", nested.Schema.Ref, toFileName(nested.Schema.Ref))
						}
						reqStr := "No"
						if nested.Required {
							reqStr = "Yes"
						}
						fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", nested.Name, typ, reqStr, nested.Description)
					}

					// Deep nested
					for _, nested := range field.NestedFields {
						if len(nested.NestedFields) > 0 {
							fmt.Fprintf(&b, "\n#### `%s.%s`\n\n", field.Name, nested.Name)
							b.WriteString("| Field | Type | Required | Description |\n")
							b.WriteString("|-------|------|----------|-------------|\n")
							for _, deep := range nested.NestedFields {
								typ := deep.Type
								if deep.Schema != nil && deep.Schema.Ref != "" {
									typ = fmt.Sprintf("[%s](%s.md)", deep.Schema.Ref, toFileName(deep.Schema.Ref))
								}
								reqStr := "No"
								if deep.Required {
									reqStr = "Yes"
								}
								fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", deep.Name, typ, reqStr, deep.Description)
							}
						}
					}
				}
			}
		}

	case SchemaTypeEnum:
		if len(doc.EnumValues) > 0 {
			b.WriteString("## Values\n\n")
			for _, v := range doc.EnumValues {
				fmt.Fprintf(&b, "- `%v`\n", v)
			}
		}

	default:
		if len(doc.Composition) > 0 {
			b.WriteString("## Composition\n\n")
			for _, item := range doc.Composition {
				if item.Ref != "" {
					fmt.Fprintf(&b, "- [%s](%s.md)\n", item.Ref, toFileName(item.Ref))
				} else {
					b.WriteString("- (inline schema)\n")
				}
			}
		} else if doc.Type == SchemaTypeArray && doc.Items != nil {
			if doc.Items.Ref != "" {
				fmt.Fprintf(&b, "Array of [%s](%s.md)\n", doc.Items.Ref, toFileName(doc.Items.Ref))
			} else {
				b.WriteString("Array of object\n")
			}
		}
	}

	return b.String()
}

// RenderSchemaIndex renders the schema group index markdown.
func (r *Renderer) RenderSchemaIndex(group SchemaGroupDocument) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s Schemas\n\n", group.Prefix)
	fmt.Fprintf(&b, "%d schemas in this group.\n\n", len(group.Schemas))
	b.WriteString("| Schema | Type | Description |\n")
	b.WriteString("|--------|------|-------------|\n")
	for _, schema := range group.Schemas {
		schemaDesc := schema.Description
		if len(schemaDesc) > 50 {
			schemaDesc = schemaDesc[:50]
		}
		fmt.Fprintf(&b, "| [%s](%s.md) | %s | %s |\n",
			schema.Name, toFileName(schema.Name), schema.Type, schemaDesc)
	}

	return b.String()
}

// RenderAuthentication renders the authentication schemes markdown.
func (r *Renderer) RenderAuthentication(schemes []AuthSchemeDocument) string {
	var b strings.Builder

	b.WriteString("# Authentication\n\n")
	b.WriteString("This document describes the authentication methods supported by this API.\n\n")

	for _, scheme := range schemes {
		fmt.Fprintf(&b, "## %s\n\n", scheme.Name)
		fmt.Fprintf(&b, "**Type:** %s\n\n", scheme.Type)

		if scheme.Description != "" {
			b.WriteString(scheme.Description + "\n\n")
		}

		switch scheme.Type {
		case "apiKey":
			fmt.Fprintf(&b, "- **In:** %s\n", scheme.In)
			if scheme.APIKeyName != "" {
				fmt.Fprintf(&b, "- **Name:** %s\n", scheme.APIKeyName)
			}
			b.WriteString("\n")
		case "http":
			fmt.Fprintf(&b, "- **Scheme:** %s\n", scheme.Scheme)
			if scheme.BearerFormat != "" {
				fmt.Fprintf(&b, "- **Bearer Format:** %s\n", scheme.BearerFormat)
			}
			b.WriteString("\n")
		case "oauth2":
			for _, flow := range scheme.Flows {
				fmt.Fprintf(&b, "**%s flow:**\n", flow.Name)
				if flow.AuthorizationURL != "" {
					fmt.Fprintf(&b, "- Authorization URL: %s\n", flow.AuthorizationURL)
				}
				if flow.TokenURL != "" {
					fmt.Fprintf(&b, "- Token URL: %s\n", flow.TokenURL)
				}
				if len(flow.Scopes) > 0 {
					b.WriteString("- Scopes:\n")
					for scope, desc := range flow.Scopes {
						fmt.Fprintf(&b, "  - `%s`: %s\n", scope, desc)
					}
				}
				b.WriteString("\n")
			}
		case "openIdConnect":
			fmt.Fprintf(&b, "- **OpenID Connect URL:** %s\n\n", scheme.OpenIDConnectURL)
		}
	}

	return b.String()
}

// =============================================================================
// Case strategy helpers
// =============================================================================

// ApplyCaseStrategyLowercase merges schema groups by case-insensitive prefix
// and lowercases names, disambiguating collisions with numeric suffixes.
func ApplyCaseStrategyLowercase(doc SkillDocument) SkillDocument {
	mergedMap := make(map[string]*SchemaGroupDocument)
	var groupOrder []string

	for _, group := range doc.SchemaGroups {
		key := strings.ToLower(group.Prefix)
		if existing, ok := mergedMap[key]; ok {
			existing.Schemas = append(existing.Schemas, group.Schemas...)
		} else {
			mergedMap[key] = &SchemaGroupDocument{
				Prefix:  key,
				Schemas: append([]SchemaDocument{}, group.Schemas...),
			}
			groupOrder = append(groupOrder, key)
		}
	}

	var newGroups []SchemaGroupDocument
	for _, key := range groupOrder {
		group := mergedMap[key]
		usedNames := make(map[string]bool)
		renamedSchemas := make([]SchemaDocument, 0, len(group.Schemas))
		for _, schema := range group.Schemas {
			baseName := toFileName(schema.Name)
			baseName = strings.ToLower(baseName)
			finalName := baseName
			counter := 2
			for usedNames[finalName] {
				finalName = fmt.Sprintf("%s-%d", baseName, counter)
				counter++
			}
			usedNames[finalName] = true
			schema.Name = finalName
			renamedSchemas = append(renamedSchemas, schema)
		}
		group.Schemas = renamedSchemas
		newGroups = append(newGroups, *group)
	}

	doc.SchemaGroups = newGroups
	return doc
}
