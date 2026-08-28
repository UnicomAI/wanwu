package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	wga_option "github.com/UnicomAI/wanwu/pkg/wga/wga-option"
	"github.com/gin-gonic/gin"
)

const wgaKnowledgeSearchOpenAPITemplate = `{
  "openapi": "3.0.1",
  "info": {
    "title": "知识库检索",
    "version": "1.0.0",
    "description": "从配置的知识库中检索相关信息。如果用户的任务需要获取一些信息，即使只有1%的可能性在知识库中，也一定要调用此工具进行检索。",
    "x-wanwu-type": "knowledge"
  },
  "servers": [
    {
      "url": "http://bff-service:6668/callback/v1"
    }
  ],
  "paths": {
    "/wga/rag/search-knowledge-base": {
      "post": {
        "tags": ["knowledge"],
        "summary": "知识库检索",
        "description": "根据问题从知识库中检索相关文档片段，返回与问题最相关的结果。X-Uid和knowledgeIdList为固定值，调用时只需传入question参数即可。",
        "operationId": "wgaKnowledgeSearch",
        "parameters": [
          {
            "name": "X-Uid",
            "in": "header",
            "required": true,
            "schema": {
              "type": "string",
              "default": {{.UserId | tojson}}
            },
            "description": {{.XUidDesc | tojson}}
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["knowledgeIdList", "question"],
                "properties": {
                  "knowledgeIdList": {
                    "type": "array",
                    "items": {"type": "string"},
                    "default": {{.KnowledgeIdList | tojson}},
                    "description": {{.KnowledgeIdListDesc | tojson}}
                  },
                  "question": {
                    "type": "string",
                    "description": "检索问题"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "检索结果",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "code": {"type": "integer"},
                    "msg": {"type": "string"},
                    "data": {
                      "type": "object",
                      "properties": {
                        "searchList": {
                          "type": "array",
                          "items": {
                            "type": "object",
                            "properties": {
                              "title": {"type": "string"},
                              "snippet": {"type": "string"},
                              "kb_name": {"type": "string"},
                              "score": {"type": "number"}
                            }
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`

func renderWgaKnowledgeSearchSchema(userId string, knowledgeIdList []string) ([]byte, error) {
	tmpl, err := template.New("wgaKnowledgeSearch").Funcs(template.FuncMap{
		"tojson": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
	}).Parse(wgaKnowledgeSearchOpenAPITemplate)
	if err != nil {
		return nil, fmt.Errorf("parse wga knowledge search openapi template err: %w", err)
	}

	knowledgeIdListJSON, _ := json.Marshal(knowledgeIdList)

	data := struct {
		UserId              string
		KnowledgeIdList     []string
		XUidDesc            string
		KnowledgeIdListDesc string
	}{
		UserId:              userId,
		KnowledgeIdList:     knowledgeIdList,
		XUidDesc:            fmt.Sprintf("用户ID，固定值：%s", userId),
		KnowledgeIdListDesc: fmt.Sprintf("知识库ID列表，固定值：%s", string(knowledgeIdListJSON)),
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute wga knowledge search openapi template err: %w", err)
	}
	return []byte(buf.String()), nil
}

// --- internal wga knowledge ---

// checkWgaKnowledgeConfig 校验wga Knowledge配置（用于更新配置）
func checkWgaKnowledgeConfig(ctx *gin.Context, userId, orgId string, knowledgeList []*assistant_service.WgaConfigKnowledge) error {
	if len(knowledgeList) == 0 {
		return nil
	}

	knowledgeIds := make([]string, 0, len(knowledgeList))
	for _, k := range knowledgeList {
		knowledgeIds = append(knowledgeIds, k.KnowledgeId)
	}

	validIds, err := getValidKnowledgeIds(ctx, userId, orgId, knowledgeIds)
	if err != nil {
		return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, "knowledge not found")
	}

	for _, k := range knowledgeList {
		if !validIds[k.KnowledgeId] {
			return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("knowledge not found: %s", k.KnowledgeId))
		}
	}
	return nil
}

func buildWgaKnowledgeOptions(ctx *gin.Context, userId, orgId, threadId, runId string, knowledgeList []*assistant_service.WgaConfigKnowledge) ([]wga_option.Option, error) {
	if len(knowledgeList) == 0 {
		return nil, nil
	}

	knowledgeIdList := make([]string, 0, len(knowledgeList))
	for _, k := range knowledgeList {
		knowledgeIdList = append(knowledgeIdList, k.KnowledgeId)
	}

	schemaData, err := renderWgaKnowledgeSearchSchema(userId, knowledgeIdList)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("render wga knowledge search schema err: %v", err))
	}

	doc, err := openapi3_util.LoadFromData(context.Background(), schemaData)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("load wga knowledge search schema err: %v", err))
	}

	return []wga_option.Option{
		wga_option.WithExtraTool(wga_option.ExtraTool{
			OpenAPI3Schema: doc,
		}),
	}, nil
}

// getValidKnowledgeIds 批量获取有效的Knowledge ID映射
func getValidKnowledgeIds(ctx *gin.Context, userId, orgId string, knowledgeIds []string) (map[string]bool, error) {
	if len(knowledgeIds) == 0 {
		return make(map[string]bool), nil
	}
	resp, err := knowledgeBase.SelectKnowledgeListByIdList(ctx.Request.Context(), &knowledgebase_service.BatchKnowledgeSelectReq{
		UserId:          userId,
		KnowledgeIdList: knowledgeIds,
	})
	if err != nil {
		return nil, err
	}
	validIds := make(map[string]bool)
	for _, k := range resp.KnowledgeList {
		validIds[k.KnowledgeId] = true
	}
	return validIds, nil
}
