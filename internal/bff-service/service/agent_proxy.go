package service

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

const (
	proxyAgentEventType = 0
)

const proxyAgentChatOpenAPITemplate = `{
  "openapi": "3.0.1",
  "info": {
    "title": {{.Title | tojson}},
    "version": {{.Version | tojson}},
    "description": {{.Description | tojson}},
    "x-wanwu-type": "agent"
  },
  "servers": [
    {
      "url": "http://bff-service:6668/callback/v1"
    }
  ],
  "paths": {
    "/agent/{{.AssistantId}}/chat": {
      "post": {
        "tags": ["callback"],
        "summary": {{.Description | tojson}},
        "description": {{.DescWithParams | tojson}},
        "operationId": "agentChatProxy",
        "requestBody": {
          "required": true,
          "description": "智能体问答请求体",
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["input"],
                "properties": {
                  "input": {
                    "type": "string",
                    "description": "用户输入的提问内容"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "返回智能体回答的文本",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "code": {
                      "type": "integer",
                      "format": "int64",
                      "description": "响应状态码，0表示成功"
                    },
                    "data": {
                      "type": "string",
                      "nullable": true,
                      "description": "成功时为智能体回答的文本，失败时为null"
                    },
                    "msg": {
                      "type": "string",
                      "description": "响应消息，成功时为空，失败时包含错误描述"
                    }
                  }
                },
                "example": {
                  "code": 0,
                  "data": "智能体回答的文本内容...",
                  "msg": ""
                }
              }
            }
          },
          "400": {
            "description": "请求参数错误（如缺少必填字段 input）",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "code": {
                      "type": "integer",
                      "format": "int64",
                      "description": "响应状态码，0表示成功"
                    },
                    "data": {
                      "type": "string",
                      "nullable": true,
                      "description": "成功时为智能体回答的文本，失败时为null"
                    },
                    "msg": {
                      "type": "string",
                      "description": "响应消息，成功时为空，失败时包含错误描述"
                    }
                  }
                },
                "example": {
                  "code": 400,
                  "data": null,
                  "msg": "invalid parameter"
                }
              }
            }
          },
          "500": {
            "description": "服务内部错误（如下游 agent-service 不可达）",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "code": {
                      "type": "integer",
                      "format": "int64",
                      "description": "响应状态码，0表示成功"
                    },
                    "data": {
                      "type": "string",
                      "nullable": true,
                      "description": "成功时为智能体回答的文本，失败时为null"
                    },
                    "msg": {
                      "type": "string",
                      "description": "响应消息，成功时为空，失败时包含错误描述"
                    }
                  }
                },
                "example": {
                  "code": 500,
                  "data": null,
                  "msg": "agent proxy stream request failed"
                }
              }
            }
          }
        }
      }
    }
  }
}`

func AgentChatProxy(ctx *gin.Context, assistantId string, req *request.AgentChatProxyReq) (string, error) {
	url := fmt.Sprintf("http://%s/agent/chat", config.Cfg().AgentService.Host)

	agentReq := map[string]interface{}{
		"assistantId": util.MustU32(assistantId),
		"input":       req.Input,
		"stream":      true,
		// "uploadFile":  req.UploadFile,
	}

	retCh, errCh := agentStreamProxy(ctx.Request.Context(), url, agentReq)
	if err := <-errCh; err != nil {
		return "", grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
	}

	var aggregatedResponse strings.Builder
	for data := range retCh {
		aggregatedResponse.WriteString(data)
	}

	return aggregatedResponse.String(), nil
}

func agentStreamProxy(ctx context.Context, url string, req map[string]interface{}) (<-chan string, <-chan error) {
	ret := make(chan string, 1024)
	errCh := make(chan error, 1)

	go func() {
		defer util.PrintPanicStack()
		defer close(ret)
		defer close(errCh)

		var resp *resty.Response
		var err error

		request := trace_util.NewResty(ctx).
			SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
			R().
			SetContext(ctx).
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetBody(req).
			SetDoNotParseResponse(true)

		resp, err = request.Post(url)
		if err != nil {
			wrappedErr := fmt.Errorf("agent proxy stream post request failed | url: %s | error: %v", url, err)
			log.Errorf("%v", wrappedErr.Error())
			errCh <- wrappedErr
			return
		}
		defer func() {
			if resp != nil && resp.RawResponse != nil {
				_ = resp.RawResponse.Body.Close()
			}
		}()

		if resp.StatusCode() >= 300 {
			b, err := io.ReadAll(resp.RawResponse.Body)
			if err != nil {
				wrappedErr := fmt.Errorf("agent proxy stream read response body failed | url: %s: %w", url, err)
				log.Errorf("%v", wrappedErr)
				errCh <- wrappedErr
				return
			}
			wrappedErr := fmt.Errorf("agent proxy stream request failed | url: %s | status: %d | message: %s", url, resp.StatusCode(), string(b))
			log.Errorf("%v", wrappedErr.Error())
			errCh <- wrappedErr
			return
		}

		close(errCh)

		scan := util.NewScanner(resp.RawResponse.Body)
		for scan.Scan() {
			sseData := scan.Text()
			data := parseAgentSSEData(sseData)
			if data == "" {
				continue
			}

			select {
			case ret <- data:
			case <-ctx.Done():
				log.Warnf("agent proxy stream ctx canceled | url: %s", url)
				return
			}
		}

		if scanErr := scan.Err(); scanErr != nil {
			log.Errorf("agent proxy stream scan err | url: %s | error: %v", url, scanErr)
		}
	}()

	return ret, errCh
}

func parseAgentSSEData(sseData string) string {
	sseData = strings.TrimSpace(sseData)
	if sseData == "" || !strings.HasPrefix(sseData, "data:") {
		return ""
	}

	dataStr := strings.TrimPrefix(sseData, "data:")
	dataStr = strings.TrimSpace(dataStr)
	if dataStr == "" {
		return ""
	}

	log.Infof("agent proxy response data: %s", dataStr)

	var agentResp response.AgentProxyChatResp
	if err := json.Unmarshal([]byte(dataStr), &agentResp); err != nil {
		log.Errorf("unmarshal agent response error: %v, data: %s", err, dataStr)
		return ""
	}

	if agentResp.EventType == proxyAgentEventType {
		return agentResp.Response
	}

	return ""
}

func renderAgentChatProxySchema(assistantId, name, desc string) ([]byte, error) {
	tmpl, err := template.New("agentChatProxy").Funcs(template.FuncMap{
		"tojson": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
	}).Parse(proxyAgentChatOpenAPITemplate)
	if err != nil {
		return nil, fmt.Errorf("parse agent chat proxy openapi template err: %w", err)
	}
	data := struct {
		Title          string
		Version        string
		Description    string
		DescWithParams string
		AssistantId    string
	}{
		Title:          name,
		Version:        "1.0.0",
		Description:    desc,
		DescWithParams: fmt.Sprintf("%s。请求参数：input（string，必填）为用户提问内容。成功时返回智能体回答的文本（code=0, data为回答字符串, msg为空）；失败时 code 非零，msg 包含错误信息。", desc),
		AssistantId:    assistantId,
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute agent chat proxy openapi template err: %w", err)
	}
	return []byte(buf.String()), nil
}
