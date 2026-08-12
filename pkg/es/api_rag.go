package es

import (
	"context"
	"fmt"
	"time"

	"github.com/UnicomAI/wanwu/pkg/log"
)

const (
	// RagChatHistoryIndexPattern 读/删走通配，覆盖所有月份
	RagChatHistoryIndexPattern = "rag_conversation_detail_infos_*"
)

var (
	_esRag *client
)

// RagChatHistoryIndex 按月分索引，写入时用
func RagChatHistoryIndex(t time.Time) string {
	return fmt.Sprintf("rag_conversation_detail_infos_%d%02d", t.Year(), t.Month())
}

func InitRag(ctx context.Context, cfg Config) error {
	if _esRag != nil {
		return fmt.Errorf("ES rag客户端已经初始化")
	}
	c, err := newClient(ctx, cfg)
	if err != nil {
		return err
	}
	_esRag = c
	return nil
}

func StopRag() {
	if _esRag != nil {
		_esRag.Stop()
		_esRag = nil
	}
}

// Rag 返回rag的ES客户端，初始化失败时为nil，调用方需判空降级
func Rag() *client {
	return _esRag
}

func InitRagChatHistoryIndexTemplate(ctx context.Context) error {
	templateName := "rag_conversation_detail_infos_template"
	template := `{
		"index_patterns": [
			"rag_conversation_detail_infos_*"
		],
		"template": {
			"mappings": {
				"properties": {
					"id": {
						"type": "keyword",
						"index": true
					},
					"ragId": {
						"type": "keyword",
						"index": true
					},
					"conversationId": {
						"type": "keyword",
						"index": true
					},
					"publish": {
						"type": "integer"
					},
					"userId": {
						"type": "keyword",
						"index": true
					},
					"orgId": {
						"type": "keyword",
						"index": true
					},
					"prompt": {
						"type": "text",
						"fields": {
							"keyword": {
								"type": "keyword",
								"ignore_above": 256
							}
						}
					},
					"response": {
						"type": "text",
						"fields": {
							"keyword": {
								"type": "keyword",
								"ignore_above": 256
							}
						}
					},
					"reasoningContent": {
						"type": "text",
						"index": false
					},
					"searchList": {
						"type": "text",
						"index": false
					},
					"qaSearchList": {
						"type": "text",
						"index": false
					},
					"errMessage": {
						"type": "text",
						"index": false
					},
					"qaErrMessage": {
						"type": "text",
						"index": false
					},
					"fileInfo": {
						"type": "object",
						"properties": {
							"fileName": {
								"type": "text",
								"fields": {
									"keyword": {
										"type": "keyword",
										"ignore_above": 256
									}
								}
							},
							"fileSize": {
								"type": "long"
							},
							"fileUrl": {
								"type": "text",
								"index": false
							}
						}
					},
					"createdAt": {
						"type": "date"
					},
					"updatedAt": {
						"type": "date"
					},
					"statistic": {
						"type": "object",
						"properties": {
							"startTime": {
								"type": "long"
							},
							"firstTokenLatency": {
								"type": "long"
							},
							"totalCostTime": {
								"type": "long"
							},
							"promptTokens": {
								"type": "long"
							},
							"completionTokens": {
								"type": "long"
							},
							"totalTokens": {
								"type": "long"
							}
						}
					}
				}
			}
		}
	}`
	if err := Rag().CreateIndexTemplate(ctx, templateName, template); err != nil {
		return fmt.Errorf("创建ES索引模板失败: %v", err)
	}

	log.Infof("成功创建ES索引模板: %s", templateName)
	return nil
}
