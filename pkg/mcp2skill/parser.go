package mcp2skill

import (
	"fmt"
	"strings"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
)

// Parse converts a list of MCP protocol.Tools into a SkillDocument IR.
func Parse(tools []*protocol.Tool, skillName string) SkillDocument {
	toolDocs := make([]ToolDocument, 0, len(tools))
	for _, tool := range tools {
		toolDocs = append(toolDocs, parseTool(tool))
	}

	if skillName == "" && len(tools) > 0 {
		skillName = deriveSkillName(tools)
	}

	description := ""
	if len(tools) > 0 {
		description = fmt.Sprintf("MCP skill with %d tool(s): %s", len(tools), toolNamesSummary(tools))
	}

	return SkillDocument{
		Meta: SkillMeta{
			Name:        skillName,
			Description: description,
			ToolCount:   len(tools),
		},
		Tools: toolDocs,
	}
}

func parseTool(tool *protocol.Tool) ToolDocument {
	if tool == nil {
		return ToolDocument{}
	}

	params := parseProperties(tool.InputSchema.Properties, tool.InputSchema.Required)
	required := tool.InputSchema.Required
	if required == nil {
		required = []string{}
	}

	var outputSchema *OutputSchemaDocument
	if tool.OutputSchema.Properties != nil {
		outputSchema = &OutputSchemaDocument{
			Type:       string(tool.OutputSchema.Type),
			Properties: parseProperties(tool.OutputSchema.Properties, tool.OutputSchema.Required),
		}
	}

	var annotations *ToolAnnotationsDocument
	if tool.Annotations != nil {
		annotations = &ToolAnnotationsDocument{
			Title: tool.Annotations.Title,
		}
		if tool.Annotations.ReadOnlyHint != nil {
			annotations.ReadOnlyHint = *tool.Annotations.ReadOnlyHint
		}
		if tool.Annotations.DestructiveHint != nil {
			annotations.DestructiveHint = *tool.Annotations.DestructiveHint
		}
		if tool.Annotations.IdempotentHint != nil {
			annotations.IdempotentHint = *tool.Annotations.IdempotentHint
		}
		if tool.Annotations.OpenWorldHint != nil {
			annotations.OpenWorldHint = *tool.Annotations.OpenWorldHint
		}
	}

	return ToolDocument{
		Name:         tool.Name,
		Description:  tool.Description,
		Parameters:   params,
		Required:     required,
		OutputSchema: outputSchema,
		Annotations:  annotations,
	}
}

// parseProperties recursively converts protocol.Properties map to ParameterDocument slice.
func parseProperties(props map[string]*protocol.Property, required []string) []ParameterDocument {
	if len(props) == 0 {
		return nil
	}

	requiredSet := make(map[string]bool, len(required))
	for _, r := range required {
		requiredSet[r] = true
	}

	result := make([]ParameterDocument, 0, len(props))
	for name, prop := range props {
		if prop == nil {
			continue
		}
		result = append(result, parseProperty(name, prop, requiredSet[name]))
	}
	return result
}

func parseProperty(name string, prop *protocol.Property, isRequired bool) ParameterDocument {
	doc := ParameterDocument{
		Name:        name,
		Type:        propertyTypeToString(prop.Type),
		Description: prop.Description,
		Required:    isRequired,
		Default:     formatDefault(prop.Default),
	}

	// Enum values
	if len(prop.Enum) > 0 {
		doc.Enum = make([]string, len(prop.Enum))
		for i, v := range prop.Enum {
			doc.Enum[i] = fmt.Sprintf("%v", v)
		}
	}

	// Nested object properties
	if len(prop.Properties) > 0 {
		doc.Properties = parseProperties(prop.Properties, prop.Required)
	}

	// Array items
	if prop.Items != nil {
		item := parseProperty("items", prop.Items, false)
		doc.Items = &item
	}

	return doc
}

func propertyTypeToString(pt protocol.PropertyType) string {
	if len(pt) == 0 {
		return "any"
	}
	if len(pt) == 1 {
		return string(pt[0])
	}
	parts := make([]string, len(pt))
	for i, t := range pt {
		parts[i] = string(t)
	}
	return strings.Join(parts, ",")
}

func formatDefault(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func deriveSkillName(tools []*protocol.Tool) string {
	if len(tools) == 0 {
		return "mcp-skill"
	}
	// Use common prefix if available, otherwise use first tool name.
	name := tools[0].Name
	if idx := strings.LastIndex(name, "_"); idx > 0 {
		return toFileName(name[:idx])
	}
	return toFileName(name)
}

func toolNamesSummary(tools []*protocol.Tool) string {
	names := make([]string, 0, len(tools))
	for _, t := range tools {
		if len(names) >= 5 {
			names = append(names, "...")
			break
		}
		names = append(names, t.Name)
	}
	return strings.Join(names, ", ")
}
