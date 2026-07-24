package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/UnicomAI/wanwu/internal/bff-service/config"
)

// SkillCategory constants for wanwu API categories.
const (
	SkillCategoryAgent    = "agent"
	SkillCategoryWorkflow = "workflow"
	SkillCategoryRAG      = "rag"
)

// wanwuOpenAPITemplates maps category names to their OpenAPI JSON templates.
// The templates use Go text/template syntax with {{.UUID}} as the placeholder
// for the application UUID that gets embedded into the API spec.
var wanwuOpenAPITemplates = map[string]string{
	// Agent API: create conversation + chat. These two APIs are related:
	// 1. Call agentCreateConversation to get a conversation_id
	// 2. Use that conversation_id in agentChat for multi-turn dialogue
	// 3. If conversation_id is omitted in agentChat, it performs a single-turn conversation
	SkillCategoryAgent: agentOpenAPITemplate,

	// Workflow API: run workflow + upload file
	SkillCategoryWorkflow: workflowOpenAPITemplate,

	// RAG API: chat with knowledge base
	SkillCategoryRAG: ragOpenAPITemplate,
}

// renderWanwuOpenAPISpec renders the OpenAPI JSON spec for the given category,
// with the uuid and metadata embedded into the spec. Returns the raw JSON bytes.
func renderWanwuOpenAPISpec(category, uuid, name, desc string) ([]byte, error) {
	tmpl, ok := wanwuOpenAPITemplates[category]
	if !ok {
		return nil, fmt.Errorf("unsupported wanwu API category: %s (supported: %s, %s, %s)", category, SkillCategoryAgent, SkillCategoryWorkflow, SkillCategoryRAG)
	}

	t, err := template.New(category).Funcs(template.FuncMap{
		"tojson": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
	}).Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("parse %s openapi template err: %w", category, err)
	}

	data := struct {
		UUID      string
		Name      string
		Desc      string
		ServerURL string
		Category  string
	}{
		UUID:      uuid,
		Name:      name,
		Desc:      desc,
		ServerURL: config.Cfg().Server.WebBaseUrl,
		Category:  category,
	}
	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute %s openapi template err: %w", category, err)
	}
	return []byte(buf.String()), nil
}

const agentOpenAPITemplate = `{
  "openapi": "3.0.1",
  "info": {
    "title": "{{.Name}}",
    "version": "1.0.0",
    "description": "{{.Desc}}智能体对话与创建会话接口。包含两个关联接口：先调用创建对话接口获取conversation_id，再在智能体对话接口中传入conversation_id进行多轮对话；若不传conversation_id则为单轮对话。",
    "x-wanwu-type": "{{.Category}}"
  },
  "servers": [
    {
      "url": "{{.ServerURL}}",
      "description": "服务地址"
    }
  ],
  "paths": {
    "/service/api/openapi/v1/agent/conversation": {
      "post": {
        "operationId": "agentCreateConversation",
        "summary": "创建对话",
        "description": "创建一个新的智能体对话会话，返回conversation_id，后续对话接口需携带此ID。该接口与智能体对话接口配合使用：先调用本接口获取conversation_id，再将该ID传入智能体对话接口实现多轮对话。",
        "tags": ["智能体"],
        "security": [{"BearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["uuid"],
                "properties": {
                  "uuid": {
                    "type": "string",
                    "description": "智能体UUID，固定值为 {{.UUID}}",
                    "default": "{{.UUID}}"
                  },
                  "title": {
                    "type": "string",
                    "description": "对话标题"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "成功响应",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "code": {"type": "integer", "description": "状态码，0表示成功"},
                    "msg": {"type": "string", "description": "提示信息"},
                    "data": {
                      "type": "object",
                      "properties": {
                        "conversation_id": {"type": "string", "description": "会话ID，用于后续对话接口传参"}
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/service/api/openapi/v1/agent/chat": {
      "post": {
        "operationId": "agentChat",
        "summary": "智能体对话",
        "description": "向智能体发起问答。支持流式和非流式两种模式。先通过创建对话接口获取conversation_id，再传入本接口实现多轮对话；不传conversation_id则为单轮对话，无历史上下文。如需在对话中携带文件，请先调用文件上传API获取文件信息后再传入。",
        "tags": ["智能体"],
        "security": [{"BearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["uuid", "query"],
                "properties": {
                  "uuid": {
                    "type": "string",
                    "description": "智能体UUID，固定值为 {{.UUID}}",
                    "default": "{{.UUID}}"
                  },
                  "conversation_id": {
                    "type": "string",
                    "description": "会话ID，由创建对话接口返回；不传时为单轮对话，无历史上下文"
                  },
                  "query": {
                    "type": "string",
                    "description": "用户输入的问题或指令"
                  },
                  "stream": {
                    "type": "boolean",
                    "description": "是否流式返回，默认为false（非流式）"
                  },
                  "file_info": {
                    "type": "array",
                    "description": "随对话携带的文件列表，文件信息需通过文件上传API提前获取",
                    "items": {
                      "type": "object",
                      "properties": {
                        "fileName": {"type": "string", "description": "上传后的文件名称"},
                        "fileSize": {"type": "integer", "description": "文件大小（字节）"},
                        "fileUrl": {"type": "string", "description": "文件完整访问路径"}
                      }
                    }
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "成功响应",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "code": {"type": "integer", "description": "状态码，0表示成功"},
                    "message": {"type": "string", "description": "提示信息"},
                    "response": {"type": "string", "description": "模型回答文本"},
                    "search_list": {
                      "type": "array",
                      "description": "知识库检索结果列表",
                      "items": {
                        "type": "object",
                        "properties": {
                          "kb_name": {"type": "string", "description": "知识库名称"},
                          "snippet": {"type": "string", "description": "命中的知识片段内容"},
                          "title": {"type": "string", "description": "来源文件标题"}
                        }
                      }
                    },
                    "usage": {
                      "type": "object",
                      "description": "Token用量统计",
                      "properties": {
                        "prompt_tokens": {"type": "integer", "description": "输入Token数"},
                        "completion_tokens": {"type": "integer", "description": "输出Token数"},
                        "total_tokens": {"type": "integer", "description": "总Token数"}
                      }
                    },
                    "finish": {"type": "integer", "description": "仅流式返回。0-未结束，1-正常结束，2-超长度截断，3-异常结束，4-命中安全护栏"}
                  }
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "API Key",
        "description": "通过 Authorization: Bearer {API Key} 进行鉴权"
      }
    }
  },
  "security": [{"BearerAuth": []}]
}`

const workflowOpenAPITemplate = `{
  "openapi": "3.0.1",
  "info": {
    "title": "{{.Name}}",
    "version": "1.0.0",
    "description": "{{.Desc}}工作流运行与文件上传接口。如果需要上传文件，先通过文件上传接口上传文件获取链接，再在工作流运行接口中将链接作为参数传入。",
    "x-wanwu-type": "{{.Category}}"
  },
  "servers": [
    {
      "url": "{{.ServerURL}}",
      "description": "服务地址"
    }
  ],
  "paths": {
    "/service/api/openapi/v1/workflow/run": {
      "post": {
        "operationId": "workflowRun",
        "summary": "工作流运行",
        "description": "执行指定的工作流并返回运行结果。可通过parameters传入工作流所需的参数，包括通过文件上传接口获取的文件链接。",
        "tags": ["工作流"],
        "security": [{"BearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["uuid", "parameters"],
                "properties": {
                  "uuid": {
                    "type": "string",
                    "description": "工作流唯一ID，固定值为 {{.UUID}}",
                    "default": "{{.UUID}}"
                  },
                  "parameters": {
                    "type": "object",
                    "description": "工作流执行所需的参数对象",
                    "additionalProperties": {}
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "成功响应，返回工作流输出（JSON对象，取决于工作流）",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "additionalProperties": {}
                }
              }
            }
          }
        }
      }
    },
    "/service/api/openapi/v1/workflow/file/upload": {
      "post": {
        "operationId": "workflowFileUpload",
        "summary": "工作流文件上传",
        "description": "上传文件到工作流，返回文件链接。上传成功后，可将返回的文件链接作为工作流运行接口中parameters的值传入。",
        "tags": ["工作流"],
        "security": [{"BearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "multipart/form-data": {
              "schema": {
                "type": "object",
                "required": ["file"],
                "properties": {
                  "file": {
                    "type": "string",
                    "format": "binary",
                    "description": "需要上传的文件"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "成功响应，返回文件链接",
            "content": {
              "application/json": {
                "schema": {
                  "type": "string",
                  "description": "上传文件在OSS中的链接"
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "API Key",
        "description": "通过 Authorization: Bearer {API Key} 进行鉴权"
      }
    }
  },
  "security": [{"BearerAuth": []}]
}`

const ragOpenAPITemplate = `{
  "openapi": "3.0.1",
  "info": {
    "title": "{{.Name}}",
    "version": "1.0.0",
    "description": "{{.Desc}}基于知识库的RAG问答接口，适合用于翻译、文章写作、总结等文本生成场景。支持流式和非流式两种模式。",
    "x-wanwu-type": "{{.Category}}"
  },
  "servers": [
    {
      "url": "{{.ServerURL}}",
      "description": "服务地址"
    }
  ],
  "paths": {
    "/service/api/openapi/v1/rag/chat": {
      "post": {
        "operationId": "ragChat",
        "summary": "知识问答",
        "description": "基于知识库的RAG问答接口。支持流式和非流式两种模式，通过stream参数控制。",
        "tags": ["知识问答"],
        "security": [{"BearerAuth": []}],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["uuid", "query"],
                "properties": {
                  "uuid": {
                    "type": "string",
                    "description": "文本问答唯一ID，固定值为 {{.UUID}}",
                    "default": "{{.UUID}}"
                  },
                  "query": {
                    "type": "string",
                    "description": "用户提出的问题或提示语"
                  },
                  "stream": {
                    "type": "boolean",
                    "description": "是否以流式接口的形式返回数据，默认为false"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "成功响应",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "code": {"type": "integer", "description": "状态码，0表示成功"},
                    "message": {"type": "string", "description": "提示信息"},
                    "msg_id": {"type": "string", "description": "提示信息ID"},
                    "data": {
                      "type": "object",
                      "properties": {
                        "output": {"type": "string", "description": "当前响应文本内容片段"},
                        "searchList": {
                          "type": "array",
                          "description": "知识增强搜索结果",
                          "items": {
                            "type": "object",
                            "properties": {
                              "kb_name": {"type": "string", "description": "知识库名字"},
                              "snippet": {"type": "string", "description": "知识内容片段"},
                              "title": {"type": "string", "description": "文件标题"}
                            }
                          }
                        }
                      }
                    },
                    "history": {
                      "type": "array",
                      "description": "对话历史",
                      "items": {
                        "type": "object",
                        "properties": {
                          "query": {"type": "string", "description": "请求文本"},
                          "response": {"type": "string", "description": "模型响应文本"}
                        }
                      }
                    },
                    "finish": {"type": "integer", "description": "仅流式返回。0-未结束，1-正常结束，2-生成长度导致结束，3-异常结束，4-命中安全护栏结束"}
                  }
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "API Key",
        "description": "通过 Authorization: Bearer {API Key} 进行鉴权"
      }
    }
  },
  "security": [{"BearerAuth": []}]
}`
