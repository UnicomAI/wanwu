package openapi2skill

import (
	"regexp"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

var httpMethods = []string{"GET", "PUT", "POST", "DELETE", "OPTIONS", "HEAD", "PATCH", "TRACE"}

var versionPrefixRe = regexp.MustCompile(`^/(api/)?(v\d+/)?`)

// schemaPrefixRe matches a PascalCase prefix (e.g., "Pet" from "PetFood").
var schemaPrefixRe = regexp.MustCompile(`^([A-Z][a-z]+)`)

// underscorePrefixRe matches text before the first underscore.
var underscorePrefixRe = regexp.MustCompile(`^([^_]+)`)

// Parse converts an OpenAPI 3 document into a SkillDocument IR.
func Parse(doc *openapi3.T, opts ParserOptions) SkillDocument {
	filter := opts.Filter
	if filter == nil {
		filter = &ParserFilter{}
	}
	groupBy := opts.GroupBy
	if groupBy == "" {
		groupBy = GroupByAuto
	}

	meta := parseMeta(doc, opts.SkillName)
	resources := parseResources(doc, filter, groupBy)
	schemaGroups := parseSchemaGroups(doc)
	authSchemes := parseAuthSchemes(doc)

	return SkillDocument{
		Meta:         meta,
		Resources:    resources,
		SchemaGroups: schemaGroups,
		AuthSchemes:  authSchemes,
	}
}

func parseMeta(doc *openapi3.T, skillName string) SkillMeta {
	description := ""
	if doc.Info != nil {
		description = doc.Info.Description
	}

	name := skillName
	if name == "" && doc.Info != nil {
		name = toFileName(doc.Info.Title)
	}
	name = strings.ToLower(name)

	var license *LicenseDocument
	if doc.Info != nil && doc.Info.License != nil {
		license = &LicenseDocument{
			Name: doc.Info.License.Name,
			URL:  doc.Info.License.URL,
		}
	}

	var contact string
	if doc.Info != nil && doc.Info.Contact != nil {
		contact = doc.Info.Contact.Email
	}

	var servers []ServerDocument
	for _, s := range doc.Servers {
		servers = append(servers, ServerDocument{
			URL:         s.URL,
			Description: s.Description,
		})
	}

	var securitySchemes []string
	if doc.Components != nil && doc.Components.SecuritySchemes != nil {
		for k := range doc.Components.SecuritySchemes {
			securitySchemes = append(securitySchemes, k)
		}
	}

	openAPIVersion := doc.OpenAPI
	if openAPIVersion == "" {
		openAPIVersion = "3.0.0"
	}

	version := ""
	if doc.Info != nil {
		version = doc.Info.Version
	}

	return SkillMeta{
		Name:            name,
		Title:           doc.Info.Title,
		Description:     description,
		Version:         version,
		OpenAPIVersion:  openAPIVersion,
		License:         license,
		Contact:         contact,
		Servers:         servers,
		SecuritySchemes: securitySchemes,
	}
}

func parseResources(doc *openapi3.T, filter *ParserFilter, groupBy GroupByStrategy) []ResourceDocument {
	tagDescriptions := make(map[string]string)
	for _, tag := range doc.Tags {
		tagDescriptions[tag.Name] = tag.Description
	}

	resourceMap := make(map[string]*ResourceDocument)
	var resourceOrder []string

	for path, pathItem := range doc.Paths {
		if pathItem == nil {
			continue
		}
		if isPathExcluded(path, filter) {
			continue
		}

		for _, method := range httpMethods {
			operation := getOperation(pathItem, method)
			if operation == nil {
				continue
			}
			if filter.ExcludeDeprecated && operation.Deprecated {
				continue
			}

			resourceNames := getResourceNames(path, operation, groupBy)
			for _, resourceName := range resourceNames {
				if !isTagIncluded(resourceName, filter) {
					continue
				}
				if _, exists := resourceMap[resourceName]; !exists {
					resourceMap[resourceName] = &ResourceDocument{
						Tag:         resourceName,
						Description: tagDescriptions[resourceName],
					}
					resourceOrder = append(resourceOrder, resourceName)
				}
				opDoc := parseOperation(path, method, operation, resourceName)
				resourceMap[resourceName].Operations = append(resourceMap[resourceName].Operations, opDoc)
			}
		}
	}

	resources := make([]ResourceDocument, 0, len(resourceOrder))
	for _, name := range resourceOrder {
		resources = append(resources, *resourceMap[name])
	}

	// Sort by operation count descending.
	sortResourcesByOpCount(resources)

	return resources
}

func getOperation(pathItem *openapi3.PathItem, method string) *openapi3.Operation {
	switch method {
	case "GET":
		return pathItem.Get
	case "PUT":
		return pathItem.Put
	case "POST":
		return pathItem.Post
	case "DELETE":
		return pathItem.Delete
	case "OPTIONS":
		return pathItem.Options
	case "HEAD":
		return pathItem.Head
	case "PATCH":
		return pathItem.Patch
	case "TRACE":
		return pathItem.Trace
	default:
		return nil
	}
}

func getResourceNames(path string, operation *openapi3.Operation, groupBy GroupByStrategy) []string {
	switch groupBy {
	case GroupByTags:
		if len(operation.Tags) > 0 {
			return operation.Tags
		}
		return []string{"default"}
	case GroupByPath:
		return []string{extractResourceFromPath(path)}
	case GroupByAuto:
		if len(operation.Tags) > 0 {
			return operation.Tags
		}
		return []string{extractResourceFromPath(path)}
	default:
		return []string{"default"}
	}
}

func extractResourceFromPath(path string) string {
	stripped := versionPrefixRe.ReplaceAllString(path, "/")
	segments := segments(stripped)
	if len(segments) == 0 {
		return "default"
	}
	first := segments[0]
	if strings.HasPrefix(first, "{") {
		return "default"
	}
	return first
}

func segments(path string) []string {
	parts := strings.Split(path, "/")
	var result []string
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func parseOperation(path, method string, operation *openapi3.Operation, tag string) OperationDocument {
	opID := operation.OperationID
	if opID == "" {
		opID = method + "-" + strings.ReplaceAll(path, "/", "-")
	}

	var securityReqs openapi3.SecurityRequirements
	if operation.Security != nil {
		securityReqs = *operation.Security
	}

	return OperationDocument{
		OperationID: opID,
		Path:        path,
		Method:      method,
		Tag:         tag,
		Summary:     operation.Summary,
		Description: operation.Description,
		Deprecated:  operation.Deprecated,
		Parameters:  parseParameters(operation.Parameters),
		RequestBody: parseRequestBody(operation.RequestBody),
		Responses:   parseResponses(operation.Responses),
		Security:    parseSecurity(securityReqs),
	}
}

func parseParameters(params openapi3.Parameters) []ParameterDocument {
	var result []ParameterDocument
	for _, p := range params {
		if p == nil || p.Value == nil {
			continue
		}
		v := p.Value
		pd := ParameterDocument{
			Name:        v.Name,
			In:          v.In,
			Type:        getSchemaTypeFromRef(v.Schema),
			Required:    v.Required,
			Description: v.Description,
		}
		if v.Schema != nil && v.Schema.Value != nil {
			sr := parseSchemaRef(v.Schema.Value)
			pd.Schema = &sr
		}
		result = append(result, pd)
	}
	return result
}

func parseRequestBody(reqBody *openapi3.RequestBodyRef) *RequestBodyDocument {
	if reqBody == nil || reqBody.Value == nil {
		return nil
	}
	v := reqBody.Value

	var contentTypes []string
	var schema *SchemaRefDocument
	for ct, media := range v.Content {
		contentTypes = append(contentTypes, ct)
		if schema == nil && media != nil && media.Schema != nil && media.Schema.Value != nil {
			sr := parseSchemaRef(media.Schema.Value)
			schema = &sr
		}
	}

	return &RequestBodyDocument{
		Description:  v.Description,
		Required:     v.Required,
		ContentTypes: contentTypes,
		Schema:       schema,
	}
}

func parseResponses(responses openapi3.Responses) []ResponseDocument {
	var result []ResponseDocument
	for status, resp := range responses {
		if resp == nil {
			continue
		}
		if resp.Ref != "" {
			result = append(result, ResponseDocument{
				Status:      status,
				Description: "(reference)",
			})
			continue
		}
		if resp.Value == nil {
			continue
		}
		rd := ResponseDocument{
			Status: status,
		}
		if resp.Value.Description != nil {
			rd.Description = *resp.Value.Description
		}
		if media, ok := resp.Value.Content["application/json"]; ok && media != nil && media.Schema != nil && media.Schema.Value != nil {
			sr := parseSchemaRef(media.Schema.Value)
			rd.Schema = &sr
		}
		result = append(result, rd)
	}
	return result
}

func parseSecurity(security openapi3.SecurityRequirements) []SecurityRequirementDocument {
	var result []SecurityRequirementDocument
	for _, req := range security {
		for name, scopes := range req {
			result = append(result, SecurityRequirementDocument{
				Name:   name,
				Scopes: scopes,
			})
		}
	}
	return result
}

func parseSchemaGroups(doc *openapi3.T) []SchemaGroupDocument {
	if doc.Components == nil || doc.Components.Schemas == nil {
		return nil
	}

	groupMap := make(map[string]*SchemaGroupDocument)
	var groupOrder []string

	for name, schemaRef := range doc.Components.Schemas {
		prefix := extractSchemaPrefix(name)
		if _, exists := groupMap[prefix]; !exists {
			groupMap[prefix] = &SchemaGroupDocument{Prefix: prefix}
			groupOrder = append(groupOrder, prefix)
		}
		sd := parseSchema(name, schemaRef)
		groupMap[prefix].Schemas = append(groupMap[prefix].Schemas, sd)
	}

	result := make([]SchemaGroupDocument, 0, len(groupOrder))
	for _, prefix := range groupOrder {
		result = append(result, *groupMap[prefix])
	}
	return result
}

func parseSchema(name string, schemaRef *openapi3.SchemaRef) SchemaDocument {
	if schemaRef.Ref != "" {
		return SchemaDocument{
			Name:        name,
			Type:        SchemaTypeObject,
			Description: "Reference: " + schemaRef.Ref,
		}
	}
	schema := schemaRef.Value
	if schema == nil {
		return SchemaDocument{Name: name, Type: SchemaTypePrimitive}
	}

	schemaType := getSchemaDocType(schema)
	doc := SchemaDocument{
		Name:        name,
		Type:        schemaType,
		Description: schema.Description,
	}

	switch schemaType {
	case SchemaTypeObject:
		if schema.Properties != nil {
			doc.Fields = parseFields(schema)
		}
	case SchemaTypeEnum:
		if schema.Enum != nil {
			doc.EnumValues = append(doc.EnumValues, schema.Enum...)
		}
	case SchemaTypeAllOf, SchemaTypeOneOf, SchemaTypeAnyOf:
		var composite []*openapi3.SchemaRef
		switch schemaType {
		case SchemaTypeAllOf:
			composite = schema.AllOf
		case SchemaTypeOneOf:
			composite = schema.OneOf
		case SchemaTypeAnyOf:
			composite = schema.AnyOf
		}
		for _, item := range composite {
			if item == nil {
				continue
			}
			if item.Value != nil {
				sr := parseSchemaRef(item.Value)
				doc.Composition = append(doc.Composition, sr)
			} else if item.Ref != "" {
				doc.Composition = append(doc.Composition, SchemaRefDocument{Ref: getRefName(item.Ref)})
			}
		}
	case SchemaTypeArray:
		if schema.Items != nil {
			if schema.Items.Value != nil {
				sr := parseSchemaRef(schema.Items.Value)
				doc.Items = &sr
			} else if schema.Items.Ref != "" {
				doc.Items = &SchemaRefDocument{Ref: getRefName(schema.Items.Ref) + "[]"}
			}
		}
	}

	return doc
}

func parseFields(schema *openapi3.Schema) []FieldDocument {
	requiredSet := make(map[string]bool)
	for _, r := range schema.Required {
		requiredSet[r] = true
	}

	var fields []FieldDocument
	for propName, propRef := range schema.Properties {
		if propRef == nil {
			continue
		}
		isRequired := requiredSet[propName]
		fd := parseField(propName, propRef, isRequired)
		fields = append(fields, fd)
	}
	return fields
}

func parseField(name string, schemaRef *openapi3.SchemaRef, isRequired bool) FieldDocument {
	fd := FieldDocument{
		Name:     name,
		Type:     getSchemaTypeFromRef(schemaRef),
		Required: isRequired,
	}

	if schemaRef.Ref != "" {
		fd.Schema = &SchemaRefDocument{Ref: getRefName(schemaRef.Ref)}
		return fd
	}

	if schemaRef.Value == nil {
		return fd
	}

	schema := schemaRef.Value
	fd.Description = schema.Description

	// Handle nested inline objects.
	if schema.Type == "object" && len(schema.Properties) > 0 {
		fd.NestedFields = parseFields(schema)
	}

	// Handle array of inline objects.
	if schema.Type == "array" && schema.Items != nil && schema.Items.Value != nil {
		items := schema.Items.Value
		if items.Type == "object" && len(items.Properties) > 0 {
			fd.NestedFields = parseFields(items)
		}
	}

	if schemaRef.Value != nil {
		sr := parseSchemaRef(schemaRef.Value)
		fd.Schema = &sr
	}

	return fd
}

func parseSchemaRef(schema *openapi3.Schema) SchemaRefDocument {
	// Handle array with reference items.
	if schema.Type == "array" && schema.Items != nil {
		if schema.Items.Ref != "" {
			return SchemaRefDocument{Ref: getRefName(schema.Items.Ref) + "[]"}
		}
	}

	// Inline schema.
	sd := SchemaDocument{
		Name:        "(inline)",
		Type:        getSchemaDocType(schema),
		Description: schema.Description,
	}

	switch sd.Type {
	case SchemaTypeObject:
		if schema.Properties != nil {
			sd.Fields = parseFields(schema)
		}
	case SchemaTypeEnum:
		if schema.Enum != nil {
			sd.EnumValues = append(sd.EnumValues, schema.Enum...)
		}
	case SchemaTypeArray:
		if schema.Items != nil && schema.Items.Value != nil {
			sr := parseSchemaRef(schema.Items.Value)
			sd.Items = &sr
		}
	}

	return SchemaRefDocument{Inline: &sd}
}

func parseAuthSchemes(doc *openapi3.T) []AuthSchemeDocument {
	if doc.Components == nil || doc.Components.SecuritySchemes == nil {
		return nil
	}

	var schemes []AuthSchemeDocument
	for name, schemeRef := range doc.Components.SecuritySchemes {
		if schemeRef == nil || schemeRef.Value == nil {
			continue
		}
		scheme := schemeRef.Value

		sd := AuthSchemeDocument{
			Name:        name,
			Type:        scheme.Type,
			Description: scheme.Description,
		}

		switch scheme.Type {
		case "apiKey":
			sd.In = scheme.In
			sd.APIKeyName = scheme.Name
		case "http":
			sd.Scheme = scheme.Scheme
			sd.BearerFormat = scheme.BearerFormat
		case "oauth2":
			sd.Flows = parseOAuthFlows(scheme.Flows)
		case "openIdConnect":
			sd.OpenIDConnectURL = scheme.OpenIdConnectUrl
		}

		schemes = append(schemes, sd)
	}
	return schemes
}

func parseOAuthFlows(flows *openapi3.OAuthFlows) []OAuthFlowDocument {
	if flows == nil {
		return nil
	}

	var result []OAuthFlowDocument

	if flows.Implicit != nil {
		result = append(result, OAuthFlowDocument{
			Name:             "implicit",
			AuthorizationURL: flows.Implicit.AuthorizationURL,
			Scopes:           flows.Implicit.Scopes,
		})
	}
	if flows.Password != nil {
		result = append(result, OAuthFlowDocument{
			Name:     "password",
			TokenURL: flows.Password.TokenURL,
			Scopes:   flows.Password.Scopes,
		})
	}
	if flows.ClientCredentials != nil {
		result = append(result, OAuthFlowDocument{
			Name:     "clientCredentials",
			TokenURL: flows.ClientCredentials.TokenURL,
			Scopes:   flows.ClientCredentials.Scopes,
		})
	}
	if flows.AuthorizationCode != nil {
		result = append(result, OAuthFlowDocument{
			Name:             "authorizationCode",
			AuthorizationURL: flows.AuthorizationCode.AuthorizationURL,
			TokenURL:         flows.AuthorizationCode.TokenURL,
			Scopes:           flows.AuthorizationCode.Scopes,
		})
	}

	return result
}

// =============================================================================
// Filter helpers
// =============================================================================

func isPathExcluded(path string, filter *ParserFilter) bool {
	for _, pattern := range filter.ExcludePaths {
		if path == pattern || strings.HasPrefix(path, pattern) {
			return true
		}
	}
	return false
}

func isTagIncluded(tag string, filter *ParserFilter) bool {
	if len(filter.IncludeTags) > 0 {
		for _, t := range filter.IncludeTags {
			if t == tag {
				return true
			}
		}
		return false
	}
	if len(filter.ExcludeTags) > 0 {
		for _, t := range filter.ExcludeTags {
			if t == tag {
				return false
			}
		}
	}
	return true
}

// =============================================================================
// Utility helpers
// =============================================================================

func getRefName(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ref
}

func extractSchemaPrefix(name string) string {
	if m := schemaPrefixRe.FindStringSubmatch(name); len(m) > 1 {
		return m[1]
	}
	if m := underscorePrefixRe.FindStringSubmatch(name); len(m) > 1 {
		return m[1]
	}
	return "Other"
}

func getSchemaTypeFromRef(schemaRef *openapi3.SchemaRef) string {
	if schemaRef == nil {
		return "any"
	}
	if schemaRef.Ref != "" {
		return getRefName(schemaRef.Ref)
	}
	if schemaRef.Value == nil {
		return "any"
	}
	return getSchemaType(schemaRef.Value)
}

func getSchemaType(s *openapi3.Schema) string {
	if s == nil {
		return "any"
	}

	if len(s.Enum) > 0 {
		vals := make([]string, 0, 3)
		for i, v := range s.Enum {
			if i >= 3 {
				vals = append(vals, "...")
				break
			}
			vals = append(vals, toString(v))
		}
		return "enum: " + strings.Join(vals, ", ")
	}

	if s.Type == "array" && s.Items != nil {
		if s.Items.Ref != "" {
			return getRefName(s.Items.Ref) + "[]"
		}
		if s.Items.Value != nil {
			return s.Items.Value.Type + "[]"
		}
		return "any[]"
	}

	typ := s.Type
	if typ == "" {
		typ = "any"
	}
	if s.Format != "" {
		typ += " (" + s.Format + ")"
	}
	return typ
}

func getSchemaDocType(schema *openapi3.Schema) SchemaDocumentType {
	if len(schema.Enum) > 0 {
		return SchemaTypeEnum
	}
	if len(schema.AllOf) > 0 {
		return SchemaTypeAllOf
	}
	if len(schema.OneOf) > 0 {
		return SchemaTypeOneOf
	}
	if len(schema.AnyOf) > 0 {
		return SchemaTypeAnyOf
	}
	if schema.Type == "array" {
		return SchemaTypeArray
	}
	if schema.Type == "object" || len(schema.Properties) > 0 {
		return SchemaTypeObject
	}
	return SchemaTypePrimitive
}

func toString(v interface{}) string {
	if v == nil {
		return "null"
	}
	switch val := v.(type) {
	case string:
		return val
	default:
		s, ok := v.(string)
		if ok {
			return s
		}
		return ""
	}
}

func sortResourcesByOpCount(resources []ResourceDocument) {
	sort.Slice(resources, func(i, j int) bool {
		return len(resources[i].Operations) > len(resources[j].Operations)
	})
}
