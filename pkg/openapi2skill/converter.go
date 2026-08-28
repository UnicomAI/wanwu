package openapi2skill

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
)

// Convert converts an OpenAPI 3 specification into the Agent Skills markdown
// format and writes the output files to the specified directory.
//
// The output directory structure:
//
//	{outputDir}/{skillName}/
//	  SKILL.md
//	  references/
//	    resources/
//	    operations/
//	    schemas/
//	    authentication.md  (if auth schemes exist)
func Convert(ctx context.Context, specData []byte, opts ConvertOptions) error {
	// Load OpenAPI spec.
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(specData)
	if err != nil {
		return fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}
	if err := doc.Validate(ctx); err != nil {
		return fmt.Errorf("invalid OpenAPI spec: %w", err)
	}

	return ConvertDoc(doc, opts)
}

// ConvertDoc converts an already-loaded OpenAPI 3 document into the Agent Skills
// markdown format and writes the output files.
func ConvertDoc(doc *openapi3.T, opts ConvertOptions) error {
	// Parse OpenAPI to IR.
	skillDoc := Parse(doc, opts.Parser)

	// Apply case strategy.
	if opts.CaseStrategy == CaseStrategyLowercase {
		skillDoc = ApplyCaseStrategyLowercase(skillDoc)
	}

	// Write output.
	renderer := NewRenderer()
	return writeSkillOutput(skillDoc, opts.OutputDir, renderer)
}

// ParseToIR parses an OpenAPI 3 spec and returns the SkillDocument IR without
// writing any files. This is useful for programmatic use of the parsed data.
func ParseToIR(ctx context.Context, specData []byte, opts ParserOptions) (SkillDocument, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(specData)
	if err != nil {
		return SkillDocument{}, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}
	if err := doc.Validate(ctx); err != nil {
		return SkillDocument{}, fmt.Errorf("invalid OpenAPI spec: %w", err)
	}

	return Parse(doc, opts), nil
}

// ParseDocToIR parses an already-loaded OpenAPI 3 document into the
// SkillDocument IR without writing any files.
func ParseDocToIR(doc *openapi3.T, opts ParserOptions) SkillDocument {
	return Parse(doc, opts)
}

// =============================================================================
// Output writing
// =============================================================================

func writeSkillOutput(doc SkillDocument, outputDir string, renderer *Renderer) error {
	skillDir := filepath.Join(outputDir, doc.Meta.Name)
	referencesDir := filepath.Join(skillDir, "references")
	resourcesDir := filepath.Join(referencesDir, "resources")
	operationsDir := filepath.Join(referencesDir, "operations")
	schemasDir := filepath.Join(referencesDir, "schemas")

	// Create directories.
	dirs := []string{skillDir, resourcesDir, operationsDir, schemasDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write SKILL.md.
	skillMd := renderer.RenderSkill(doc)
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMd), 0o644); err != nil {
		return fmt.Errorf("failed to write SKILL.md: %w", err)
	}

	// Write resources and operations.
	for _, resource := range doc.Resources {
		fileName := toFileName(resource.Tag)

		// Resource index.
		resourceMd := renderer.RenderResource(resource)
		if err := os.WriteFile(filepath.Join(resourcesDir, fileName+".md"), []byte(resourceMd), 0o644); err != nil {
			return fmt.Errorf("failed to write resource %s: %w", fileName, err)
		}

		// Individual operation files.
		for _, op := range resource.Operations {
			opFileName := toFileName(op.OperationID)
			opMd := renderer.RenderOperation(op)
			if err := os.WriteFile(filepath.Join(operationsDir, opFileName+".md"), []byte(opMd), 0o644); err != nil {
				return fmt.Errorf("failed to write operation %s: %w", opFileName, err)
			}
		}
	}

	// Write schema groups.
	for _, group := range doc.SchemaGroups {
		prefixDir := filepath.Join(schemasDir, toFileName(group.Prefix))
		if err := os.MkdirAll(prefixDir, 0o755); err != nil {
			return fmt.Errorf("failed to create schema directory %s: %w", prefixDir, err)
		}

		// Schema index.
		indexMd := renderer.RenderSchemaIndex(group)
		if err := os.WriteFile(filepath.Join(prefixDir, "_index.md"), []byte(indexMd), 0o644); err != nil {
			return fmt.Errorf("failed to write schema index for %s: %w", group.Prefix, err)
		}

		// Individual schema files.
		for _, schema := range group.Schemas {
			schemaMd := renderer.RenderSchema(schema)
			if err := os.WriteFile(filepath.Join(prefixDir, toFileName(schema.Name)+".md"), []byte(schemaMd), 0o644); err != nil {
				return fmt.Errorf("failed to write schema %s: %w", schema.Name, err)
			}
		}
	}

	// Write authentication.
	if len(doc.AuthSchemes) > 0 {
		authMd := renderer.RenderAuthentication(doc.AuthSchemes)
		if err := os.WriteFile(filepath.Join(referencesDir, "authentication.md"), []byte(authMd), 0o644); err != nil {
			return fmt.Errorf("failed to write authentication.md: %w", err)
		}
	}

	return nil
}
