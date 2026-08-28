package openapi3_util

import (
	"context"
	"errors"
	"fmt"
	"slices"

	url_util "github.com/UnicomAI/wanwu/pkg/url-util"
	"github.com/getkin/kin-openapi/openapi3"
)

type Auth struct {
	Type  string `json:"type" validate:"omitempty,oneof='none' 'apiKey'"` //
	In    string `json:"in" validate:"omitempty,oneof='header' 'query'"`  // header or query
	Name  string `json:"name"`
	Value string `json:"value"`
}

func LoadFromData(ctx context.Context, data []byte) (*openapi3.T, error) {
	doc, err := openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		return nil, err
	}
	if err = ValidateDoc(ctx, doc); err != nil {
		return nil, err
	}
	return doc, err
}

func ValidateSchema(ctx context.Context, data []byte) error {
	_, err := LoadFromData(ctx, data)
	if err != nil {
		return err
	}
	return nil
}

func ValidateDoc(ctx context.Context, doc *openapi3.T) error {
	if doc == nil {
		return errors.New("schema nil")
	}
	// check info
	if doc.Info == nil {
		return errors.New("schema info empty")
	}
	// check servers
	if len(doc.Servers) == 0 {
		return errors.New("schema servers empty")
	}
	// check operationId
	for path, pathItem := range doc.Paths {
		for method, operation := range pathItem.Operations() {
			if operation.OperationID == "" {
				return fmt.Errorf("schema path(%v) method(%v) operationId empty", path, method)
			}
		}
	}
	// validate server URLs for SSRF protection
	for _, server := range doc.Servers {
		if server.URL != "" {
			if err := url_util.ValidateURL(server.URL); err != nil {
				return fmt.Errorf("invalid server URL %q: %w", server.URL, err)
			}
		}
	}
	return doc.Validate(ctx)
}

func FilterSchemaOperations(ctx context.Context, data []byte, operationIDs []string) ([]byte, error) {
	doc, err := LoadFromData(ctx, data)
	if err != nil {
		return nil, err
	}
	return FilterDocOperations(doc, operationIDs).MarshalJSON()
}

func FilterDocOperations(doc *openapi3.T, operationIDs []string) *openapi3.T {
	paths := doc.Paths
	doc.Paths = nil
	for path, pathItem := range paths {
		for method, operation := range pathItem.Operations() {
			if slices.Contains(operationIDs, operation.OperationID) {
				doc.AddOperation(path, method, operation)
			}
		}
	}
	return doc
}
