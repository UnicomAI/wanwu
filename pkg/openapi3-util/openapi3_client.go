package openapi3_util

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

type Client struct {
	httpClient *http.Client
	doc        *openapi3.T
}

type RequestParams struct {
	HeaderParams map[string]string
	PathParams   map[string]interface{}
	QueryParams  map[string]interface{}
	BodyParams   map[string]interface{}
}

func NewClient(ctx context.Context, schema []byte) (*Client, error) {
	doc, err := LoadFromData(ctx, schema)
	if err != nil {
		return nil, err
	}
	return NewClientByDoc(doc), nil
}

func NewClientByDoc(doc *openapi3.T) *Client {
	return &Client{
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				DialContext: (&net.Dialer{
					Timeout:   time.Minute, // 连接超时时间 [EN] Connection timeout
					KeepAlive: time.Minute, // 连接保持活跃的时间 [EN] How long the connection remains active
				}).DialContext,
				ResponseHeaderTimeout: time.Minute,
			},
			Timeout: time.Minute,
		},
		doc: doc,
	}
}

func (c *Client) DoRequestByMethodPath(ctx context.Context, method, path string, params *RequestParams) (interface{}, error) {
	var exist bool
	contentType := "application/json"
	for currPath, pathItem := range c.doc.Paths.Map() {
		if currPath != path {
			continue
		}
		for currMethod, operation := range pathItem.Operations() {
			if currMethod == method {
				exist = true
				if operation.RequestBody != nil && operation.RequestBody.Value != nil {
					for ct := range operation.RequestBody.Value.Content {
						contentType = ct
						break
					}
				}
				break
			}
		}
		if exist {
			break
		}
	}
	if !exist {
		return nil, fmt.Errorf("method(%v) path(%v) not found", method, path)
	}

	baseURL, err := c.doc.Servers.BasePath()
	if err != nil {
		return nil, fmt.Errorf("get base url err: %v", err)
	}
	return executeRequest(ctx, c.httpClient, baseURL, method, path, contentType, params)
}

func (c *Client) DoRequestByOperationID(ctx context.Context, operationID string, params *RequestParams) (interface{}, error) {
	var baseURL, method, path, contentType string
	for currPath, pathItem := range c.doc.Paths.Map() {
		for currMethod, operation := range pathItem.Operations() {
			if operation.OperationID == operationID {
				method = currMethod
				path = currPath
				if operation.RequestBody != nil && operation.RequestBody.Value != nil {
					for ct := range operation.RequestBody.Value.Content {
						contentType = ct
						break
					}
				}
				break
			}
		}
		if method != "" || path != "" {
			break
		}
	}
	if method == "" || path == "" {
		return nil, fmt.Errorf("operationId(%v) not found", operationID)
	}
	if len(c.doc.Servers) > 0 {
		baseURL = c.doc.Servers[rand.Intn(len(c.doc.Servers))].URL
	} else {
		return nil, errors.New("get base url empty")
	}
	return executeRequest(ctx, c.httpClient, baseURL, method, path, contentType, params)
}

func executeRequest(
	ctx context.Context,
	httpClient *http.Client,
	baseURL string,
	method string,
	path string,
	contentType string,
	params *RequestParams,
) (interface{}, error) {

	// path
	specPath := path
	var err error
	if params != nil {
		specPath, err = buildPathWithParams(specPath, params.PathParams)
		if err != nil {
			return nil, err
		}
	}

	// base + path
	fullPath, err := url.JoinPath(baseURL, specPath)
	if err != nil {
		return nil, err
	}

	// query
	fullURL := fullPath
	if params != nil {
		fullURL = buildPathWithQuery(fullURL, params.QueryParams)
	}

	// body
	var body io.Reader
	if params != nil && params.BodyParams != nil {
		switch contentType {
		case "application/x-www-form-urlencoded":
			formData := url.Values{}
			for key, value := range params.BodyParams {
				formData.Add(key, fmt.Sprintf("%v", value))
			}
			body = strings.NewReader(formData.Encode())
		case "multipart/form-data":
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)
			for key, value := range params.BodyParams {
				if err := writer.WriteField(key, fmt.Sprintf("%v", value)); err != nil {
					_ = writer.Close()
					return nil, err
				}
			}
			body = &buf
			contentType = writer.FormDataContentType() // 获取包含boundary的完整Content-Type [EN] Get the complete Content-Type including boundary
			if err := writer.Close(); err != nil {
				return nil, err
			}
		default:
			// application/json
			jsonData, err := json.Marshal(params.BodyParams)
			if err != nil {
				return nil, err
			}
			body = bytes.NewBuffer(jsonData)
			contentType = "application/json"
		}
	}

	// req
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}

	// header
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if params != nil {
		for key, value := range params.HeaderParams {
			req.Header.Set(key, value)
		}
	}

	// execute
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("execute method(%v) url(%v) http status(%v)", method, fullURL, resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return string(respBody), nil
	}

	return result, nil
}

func buildPathWithParams(path string, pathParams map[string]interface{}) (string, error) {
	specPath := path

	// 替换路径参数 [EN] Replace path parameters
	for paramName, paramValue := range pathParams {
		placeholder := "{" + paramName + "}"
		if !strings.Contains(specPath, placeholder) {
			return "", fmt.Errorf("path parameter(%v) not found in path(%v)", paramName, path)
		}
		specPath = strings.ReplaceAll(specPath, placeholder, fmt.Sprintf("%v", paramValue))
	}

	return specPath, nil
}

func buildPathWithQuery(path string, queryParams map[string]interface{}) string {
	fullPath := path
	if len(queryParams) > 0 {
		query := url.Values{}
		for key, value := range queryParams {
			query.Add(key, fmt.Sprintf("%v", value))
		}
		fullPath = fullPath + "?" + query.Encode()
	}
	return fullPath
}
