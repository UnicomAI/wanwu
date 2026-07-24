package callback

import (
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

// WgaRagSearchKnowledgeBase
//
//	@Tags			callback
//	@Summary		WGA知识库检索
//	@Description	WGA专用知识库检索接口
//	@Accept			json
//	@Produce		json
//	@Param			X-uid	header		string									true	"用户ID"
//	@Param			data	body		request.WgaRagSearchKnowledgeBaseReq	true	"WGA知识库检索请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/wga/rag/search-knowledge-base [post]
func WgaRagSearchKnowledgeBase(ctx *gin.Context) {
	userId := ctx.GetHeader("X-uid")
	if userId == "" {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "callback: empty X-uid"))
		return
	}
	var req request.WgaRagSearchKnowledgeBaseReq
	if !gin_util.Bind(ctx, &req) {
		return
	}

	fullReq := &request.RagSearchKnowledgeBaseReq{
		UserId:          userId,
		KnowledgeIdList: req.KnowledgeIdList,
		Question:        req.Question,
		TopK:            5,
		Threshold:       0.4,
		RetrieveMethod:  "hybrid_search",
		RerankMod:       "weighted_score",
		Weight: &request.WeightParams{
			VectorWeight: 0.2,
			TextWeight:   0.8,
		},
	}

	resp, err := service.RagKnowledgeHit(ctx, fullReq)
	gin_util.Response(ctx, resp, err)
}
