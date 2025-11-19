package service

import (
	knowledgebase_splitter_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-splitter-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/gin-gonic/gin"
)

// SelectKnowledgeSplitterList 查询知识库分隔符列表 [EN] SelectKnowledgeSplitterList queries the knowledge base separator list
func SelectKnowledgeSplitterList(ctx *gin.Context, userId, orgId string, req *request.GetKnowledgeSplitterReq) (*response.KnowledgeSplitterListResp, error) {
	resp, err := knowledgeBaseSplitter.SelectKnowledgeSplitterList(ctx.Request.Context(), &knowledgebase_splitter_service.KnowledgeSplitterSelectReq{
		UserId:       userId,
		OrgId:        orgId,
		SplitterName: req.SplitterName,
	})
	if err != nil {
		return nil, err
	}
	return buildKnowledgeSplitterList(resp), nil
}

// CreateKnowledgeSplitter 创建知识库分隔符 [EN] CreateKnowledgeSplitter creates a knowledge base separator
func CreateKnowledgeSplitter(ctx *gin.Context, userId, orgId string, r *request.CreateKnowledgeSplitterReq) error {
	_, err := knowledgeBaseSplitter.CreateKnowledgeSplitter(ctx.Request.Context(), &knowledgebase_splitter_service.CreateKnowledgeSplitterReq{
		SplitterValue: r.SplitterValue,
		SplitterName:  r.SplitterName,
		UserId:        userId,
		OrgId:         orgId,
	})
	return err
}

// UpdateKnowledgeSplitter 更新知识库分隔符 [EN] UpdateKnowledgeSplitter updates the knowledge base separator
func UpdateKnowledgeSplitter(ctx *gin.Context, userId, orgId string, r *request.UpdateKnowledgeSplitterReq) error {
	_, err := knowledgeBaseSplitter.UpdateKnowledgeSplitter(ctx.Request.Context(), &knowledgebase_splitter_service.UpdateKnowledgeSplitterReq{
		SplitterName:  r.SplitterName,
		SplitterValue: r.SplitterValue,
		SplitterId:    r.SplitterId,
		UserId:        userId,
		OrgId:         orgId,
	})
	return err
}

// DeleteKnowledgeSplitter 删除知识库分隔符 [EN] DeleteKnowledgeSplitter deletes the knowledge base separator
func DeleteKnowledgeSplitter(ctx *gin.Context, userId, orgId string, r *request.DeleteKnowledgeSplitterReq) error {
	_, err := knowledgeBaseSplitter.DeleteKnowledgeSplitter(ctx.Request.Context(), &knowledgebase_splitter_service.DeleteKnowledgeSplitterReq{
		SplitterId: r.SplitterId,
		UserId:     userId,
		OrgId:      orgId,
	})
	return err
}

// buildKnowledgeSplitterList 构造知识库分隔符结果 [EN] buildKnowledgeSplitterList constructs the knowledge base separator result
func buildKnowledgeSplitterList(knowledgeSplitterListResp *knowledgebase_splitter_service.KnowledgeSplitterSelectListResp) *response.KnowledgeSplitterListResp {
	if knowledgeSplitterListResp == nil || len(knowledgeSplitterListResp.KnowledgeSplitterList) == 0 {
		return &response.KnowledgeSplitterListResp{
			KnowledgeSplitterList: make([]*response.KnowledgeSplitter, 0),
		}
	}
	var knowledgeSplitterList []*response.KnowledgeSplitter
	for _, knowledgeSplitter := range knowledgeSplitterListResp.KnowledgeSplitterList {
		knowledgeSplitterList = append(knowledgeSplitterList, &response.KnowledgeSplitter{
			SplitterId:    knowledgeSplitter.SplitterId,
			SplitterName:  knowledgeSplitter.SplitterName,
			SplitterValue: knowledgeSplitter.SplitterValue,
			Type:          knowledgeSplitter.Type,
		})
	}
	return &response.KnowledgeSplitterListResp{
		KnowledgeSplitterList: knowledgeSplitterList,
	}
}
