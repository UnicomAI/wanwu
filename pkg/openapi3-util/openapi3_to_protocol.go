package openapi3_util

import (
	"context"
	"fmt"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/getkin/kin-openapi/openapi3"
)

func Schema2ProtocolTools(ctx context.Context, schema []byte) ([]*protocol.Tool, error) {
	doc, err := LoadFromData(ctx, schema)
	if err != nil {
		return nil, err
	}
	var rets []*protocol.Tool
	for _, pathItem := range doc.Paths.Map() {
		for _, operation := range pathItem.Operations() {
			rets = append(rets, Operation2ProtocolTool(operation))
		}
	}
	return rets, nil
}

func Schema2ProtocolTool(ctx context.Context, schema []byte, operationID string) (*protocol.Tool, error) {
	doc, err := LoadFromData(ctx, schema)
	if err != nil {
		return nil, err
	}
	return Doc2ProtocolTool(doc, operationID)
}

func Doc2ProtocolTools(doc *openapi3.T) ([]*protocol.Tool, error) {
	var rets []*protocol.Tool
	for _, pathItem := range doc.Paths.Map() {
		for _, operation := range pathItem.Operations() {
			rets = append(rets, Operation2ProtocolTool(operation))
		}
	}
	return rets, nil
}

func Doc2ProtocolTool(doc *openapi3.T, operationID string) (*protocol.Tool, error) {
	var exist bool
	var ret *protocol.Tool
	for _, pathItem := range doc.Paths.Map() {
		for _, operation := range pathItem.Operations() {
			if operation.OperationID != operationID {
				continue
			}
			exist = true
			ret = Operation2ProtocolTool(operation)
			break
		}
	}
	if !exist {
		return nil, fmt.Errorf("opentionID(%v) not found", operationID)
	}
	return ret, nil
}

func Operation2ProtocolTool(operation *openapi3.Operation) *protocol.Tool {
	ret := &protocol.Tool{
		Name:        operation.OperationID,
		Description: operation.Description,
		InputSchema: protocol.InputSchema{
			Type:       protocol.Object,
			Properties: make(map[string]*protocol.Property),
		},
	}
	// Process description and ensure it is not empty
	if ret.Description == "" {
		if operation.Summary != "" {
			ret.Description = operation.Summary
		} else {
			ret.Description = operation.OperationID
		}
	}
	// Parse path parameters, query parameters, header parameters, etc.
	if operation.Parameters != nil {
		properties, requireds := Parameters2ProtocolProperties(operation.Parameters)
		for field, property := range properties {
			ret.InputSchema.Properties[field] = property
		}
		ret.InputSchema.Required = append(ret.InputSchema.Required, requireds...)
	}
	// Parse request body
	if operation.RequestBody != nil && operation.RequestBody.Value != nil {
		for _, mediaType := range operation.RequestBody.Value.Content {
			if mediaType.Schema != nil && mediaType.Schema.Value != nil {
				properties := SchemaProperties2ProtocolProperties(mediaType.Schema.Value.Properties)
				for field, property := range properties {
					ret.InputSchema.Properties[field] = property
				}
				ret.InputSchema.Required = append(ret.InputSchema.Required, mediaType.Schema.Value.Required...)
			}
		}
	}
	return ret
}

func Parameters2ProtocolProperties(parameters openapi3.Parameters) (map[string]*protocol.Property, []string) {
	if parameters == nil {
		return nil, nil
	}

	rets := make(map[string]*protocol.Property)
	var requireds []string
	for _, param := range parameters {
		if param.Value == nil {
			continue
		}

		propType := ParameterType2ProtocolDataType(param.Value)
		ret := &protocol.Property{
			Type:        propType,
			Description: param.Value.Description,
		}
		switch propType {
		case protocol.ObjectT:
			if param.Value.Schema != nil && param.Value.Schema.Value != nil {
				ret.Properties = SchemaProperties2ProtocolProperties(param.Value.Schema.Value.Properties)
				ret.Required = param.Value.Schema.Value.Required
			}
		case protocol.Array:
			if param.Value.Schema != nil && param.Value.Schema.Value != nil && param.Value.Schema.Value.Items != nil && param.Value.Schema.Value.Items.Value != nil {
				ret.Items = &protocol.Property{
					Type:        SchemaType2ProtocolDataType(param.Value.Schema.Value.Items.Value),
					Description: param.Value.Schema.Value.Items.Value.Description,
					Properties:  SchemaProperties2ProtocolProperties(param.Value.Schema.Value.Items.Value.Properties),
					Required:    param.Value.Schema.Value.Items.Value.Required,
				}
			}
		default:
		}

		field := param.Value.In + "-" + param.Value.Name
		rets[field] = ret
		if param.Value.Required {
			requireds = append(requireds, field)
		}
	}
	return rets, requireds
}

func SchemaProperties2ProtocolProperties(properties openapi3.Schemas) map[string]*protocol.Property {
	if properties == nil {
		return nil
	}

	rets := make(map[string]*protocol.Property)
	for propName, propSchema := range properties {
		if propSchema.Value == nil {
			continue
		}

		propType := SchemaType2ProtocolDataType(propSchema.Value)
		ret := &protocol.Property{
			Type:        propType,
			Description: propSchema.Value.Description,
			Properties:  SchemaProperties2ProtocolProperties(propSchema.Value.Properties),
			Required:    propSchema.Value.Required,
		}
		switch propType {
		case protocol.Array:
			if propSchema.Value.Items != nil && propSchema.Value.Items.Value != nil {
				ret.Items = &protocol.Property{
					Type:        SchemaType2ProtocolDataType(propSchema.Value.Items.Value),
					Description: propSchema.Value.Items.Value.Description,
					Properties:  SchemaProperties2ProtocolProperties(propSchema.Value.Items.Value.Properties),
					Required:    propSchema.Value.Items.Value.Required,
				}
			}
		default:
		}

		rets[propName] = ret
	}
	return rets
}

// ParameterType2ProtocolDataType gets parameter type
func ParameterType2ProtocolDataType(param *openapi3.Parameter) protocol.DataType {
	if param.Schema != nil && param.Schema.Value != nil {
		return SchemaType2ProtocolDataType(param.Schema.Value)
	}
	return protocol.String
}

// SchemaType2ProtocolDataType gets the type of schema
func SchemaType2ProtocolDataType(schema *openapi3.Schema) protocol.DataType {
	if schema.Type != nil {
		// Check for concrete types in type slices
		if schema.Type.Is("object") {
			return protocol.ObjectT
		} else if schema.Type.Is("array") {
			return protocol.Array
		} else if schema.Type.Is("string") {
			return protocol.String
		} else if schema.Type.Is("number") {
			return protocol.Number
		} else if schema.Type.Is("integer") {
			return protocol.Integer
		} else if schema.Type.Is("boolean") {
			return protocol.Boolean
		}

		// If there are multiple types, return the first one
		if len(*schema.Type) > 0 {
			return protocol.DataType((*schema.Type)[0])
		}
	}

	if len(schema.AnyOf) > 0 {
		return "anyOf"
	}
	if len(schema.AllOf) > 0 {
		return "allOf"
	}
	if len(schema.OneOf) > 0 {
		return "oneOf"
	}

	if schema.Format != "" {
		return protocol.DataType(schema.Format)
	}

	return protocol.String
}
