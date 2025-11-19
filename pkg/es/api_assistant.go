package es

import (
	"context"
	"fmt"

	"github.com/UnicomAI/wanwu/pkg/log"
)

var (
	_esAssistant *client
)

func InitAssistant(ctx context.Context, cfg Config) error {
	if _esAssistant != nil {
		return fmt.Errorf("ES assistant客户端已经初始化")
	}
	c, err := newClient(ctx, cfg)
	if err != nil {
		return err
	}
	_esAssistant = c
	return nil
}

func StopAssistant() {
	if _esAssistant != nil {
		_esAssistant.Stop()
		_esAssistant = nil
	}
}

func Assistant() *client {
	return _esAssistant
}

func InitESIndexTemplate(ctx context.Context) error {
	templateName := "conversation_detail_infos_template"

	// 检查模板是否已存在 [EN] Check if the template already exists
	exists, err := Assistant().IndexTemplateExists(ctx, templateName)
	if err != nil {
		return fmt.Errorf("检查ES索引模板失败: %v", err)
	}

	if exists {
		log.Infof("ES索引模板已存在: %s", templateName)
		return nil
	}

	// 创建索引模板 [EN] Create index template
	template := `{
		"index_patterns": [
			"conversation_detail_infos_*"
		],
		"template": {
			"mappings": {
				"properties": {
					"id": {
						"type": "keyword",
						"index": true
					},
					"assistantId": {
						"type": "keyword",
						"index": true
					},
					"conversationId": {
						"type": "keyword",
						"index": true
					},
					"prompt": {
						"type": "text",
						"fields": {
							"keyword": {
								"type": "keyword"
							}
						}
					},
					"sysPrompt": {
						"type": "text",
						"fields": {
							"keyword": {
								"type": "keyword"
							}
						}
					},
					"algPrompt": {
						"type": "text",
						"fields": {
							"keyword": {
								"type": "keyword"
							}
						}
					},
					"requestFileIds": {
						"type": "keyword",
						"index": false
					},
					"requestFileUrls": {
						"type": "keyword",
						"index": false
					},
					"response": {
						"type": "text",
						"fields": {
							"keyword": {
								"type": "keyword"
							}
						}
					},
					"algResponse": {
						"type": "text",
						"fields": {
							"keyword": {
								"type": "keyword"
							}
						}
					},
					"responseFileIds": {
						"type": "keyword",
						"index": false
					},
					"responseFileUrls": {
						"type": "keyword",
						"index": false
					},
					"searchList": {
						"type": "keyword",
						"index": false
					},
					"fileInfo": {
						"type": "object",
						"properties": {
							"fileName": {
								"type": "text",
								"fields": {
									"keyword": {
										"type": "keyword"
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
					"createdBy": {
						"type": "keyword",
						"index": true
					},
					"ts": {
						"type": "date"
					},
					"timestamp": {
						"type": "long"
					},
					"qaType": {
						"type": "integer"
					},
					"createdAt": {
						"type": "date"
					},
					"modelId": {
						"type": "keyword",  
						"fields": {
							"keyword": {
								"type": "keyword"
							}
						}
					},
					"modelVersion": {
						"type": "keyword",  
						"fields": {
							"keyword": {
								"type": "keyword"
							}
						}
					},
					"finish": {
						"type": "integer",
						"index": true
					},
					"fileFormat": {
						"type": "text",
						"index": false
					},
					"fileSize": {
						"type": "long",
						"index": false
					},
					"fileName": {
						"type": "text",
						"index": false
					},
					"videoStatus": {
						"type": "integer"
					},
					"responseId": {
						"type": "keyword",
						"index": true
					}
				}
			}
		}
	}`
	if err := Assistant().CreateIndexTemplate(ctx, templateName, template); err != nil {
		return fmt.Errorf("创建ES索引模板失败: %v", err)
	}

	log.Infof("成功创建ES索引模板: %s", templateName)
	return nil
}
