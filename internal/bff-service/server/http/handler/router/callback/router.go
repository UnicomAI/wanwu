package callback

import (
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/callback"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func Register(callbackAPI *gin.RouterGroup) {
	// callback
	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId", http.MethodGet, callback.GetModelById, "Get model by modelId")
	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId/chat/completions", http.MethodPost, callback.ModelChatCompletions, "Model Chat Completions")
	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId/embeddings", http.MethodPost, callback.ModelEmbeddings, "Model Embeddings")
	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId/rerank", http.MethodPost, callback.ModelRerank, "Model rerank")
	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId/ocr", http.MethodPost, callback.ModelOcr, "Model ocr")
	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId/gui", http.MethodPost, callback.ModelGui, "Model gui")
	mid.Sub("callback").Reg(callbackAPI, "/model/:modelId/pdf-parser", http.MethodPost, callback.ModelPdfParser, "Model PDF document parsing")
	// workflow
	mid.Sub("callback").Reg(callbackAPI, "/workflow/list", http.MethodGet, callback.GetWorkflowList, "Get Workflow by userId and spaceId")
	mid.Sub("callback").Reg(callbackAPI, "/workflow/tool/square", http.MethodGet, callback.GetWorkflowSquareTool, "Get built-in tool details")
	mid.Sub("callback").Reg(callbackAPI, "/workflow/tool/custom", http.MethodGet, callback.GetWorkflowCustomTool, "Get custom tool details")
	// rag bff proxy
	mid.Sub("callback").Reg(callbackAPI, "/rag/search-knowledge-base", http.MethodPost, callback.SearchKnowledgeBase, "Search knowledge base list (hit test)")
	mid.Sub("callback").Reg(callbackAPI, "/rag/knowledge/stream/search", http.MethodPost, callback.KnowledgeStreamSearch, "Get authorized knowledge base list by knowledge base id and current user id")
}
