package service

import (
	knowledgebase_tag_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-tag-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/gin-gonic/gin"
)

// SelectKnowledgeTagList queries the knowledge base tag list, mainly querying all the user's knowledge base tags based on userId.
func SelectKnowledgeTagList(ctx *gin.Context, userId, orgId string, req *request.KnowledgeTagSelectReq) (*response.KnowledgeTagListResp, error) {
	resp, err := knowledgeBaseTag.SelectKnowledgeTagList(ctx.Request.Context(), &knowledgebase_tag_service.KnowledgeTagSelectReq{
		UserId:      userId,
		OrgId:       orgId,
		TagName:     req.TagName,
		KnowledgeId: req.KnowledgeId,
	})
	if err != nil {
		return nil, err
	}
	return buildKnowledgeTagList(resp), nil
}

// SelectKnowledgeTagBindCount queries the number of tag bindings
func SelectKnowledgeTagBindCount(ctx *gin.Context, userId, orgId string, req *request.KnowledgeTagBindCountReq) (*response.TagBindResp, error) {
	resp, err := knowledgeBaseTag.TagBindCount(ctx.Request.Context(), &knowledgebase_tag_service.TagBindCountReq{
		UserId: userId,
		OrgId:  orgId,
		TagId:  req.TagId,
	})
	if err != nil {
		return nil, err
	}
	return &response.TagBindResp{BindCount: resp.BindCount}, nil
}

// CreateKnowledgeTag creates a knowledge base tag
func CreateKnowledgeTag(ctx *gin.Context, userId, orgId string, r *request.CreateKnowledgeTagReq) (*response.CreateKnowledgeTagResp, error) {
	resp, err := knowledgeBaseTag.CreateKnowledgeTag(ctx.Request.Context(), &knowledgebase_tag_service.CreateKnowledgeTagReq{
		TagName: r.TagName,
		UserId:  userId,
		OrgId:   orgId,
	})
	if err != nil {
		return nil, err
	}
	return &response.CreateKnowledgeTagResp{KnowledgeId: resp.TagId}, nil
}

// UpdateKnowledgeTag updates the knowledge base tag
func UpdateKnowledgeTag(ctx *gin.Context, userId, orgId string, r *request.UpdateKnowledgeTagReq) error {
	_, err := knowledgeBaseTag.UpdateKnowledgeTag(ctx.Request.Context(), &knowledgebase_tag_service.UpdateKnowledgeTagReq{
		TagId:   r.TagId,
		TagName: r.TagName,
		UserId:  userId,
		OrgId:   orgId,
	})
	return err
}

// DeleteKnowledgeTag Delete knowledge base tag
func DeleteKnowledgeTag(ctx *gin.Context, userId, orgId string, r *request.DeleteKnowledgeTagReq) error {
	_, err := knowledgeBaseTag.DeleteKnowledgeTag(ctx.Request.Context(), &knowledgebase_tag_service.DeleteKnowledgeTagReq{
		TagId:  r.TagId,
		UserId: userId,
		OrgId:  orgId,
	})
	return err
}

// BindKnowledgeTag binds knowledge base tags
func BindKnowledgeTag(ctx *gin.Context, userId, orgId string, r *request.BindKnowledgeTagReq) error {
	_, err := knowledgeBaseTag.BindKnowledgeTag(ctx.Request.Context(), &knowledgebase_tag_service.BindKnowledgeTagReq{
		KnowledgeId: r.KnowledgeId,
		TagIdList:   r.TagIdList,
		UserId:      userId,
		OrgId:       orgId,
	})
	return err
}

// buildKnowledgeTagList constructs knowledge base tag results
func buildKnowledgeTagList(knowledgeTagListResp *knowledgebase_tag_service.KnowledgeTagSelectListResp) *response.KnowledgeTagListResp {
	if knowledgeTagListResp == nil || len(knowledgeTagListResp.KnowledgeTagList) == 0 {
		return &response.KnowledgeTagListResp{
			KnowledgeTagList: make([]*response.KnowledgeTag, 0),
		}
	}
	var knowledgeTagList []*response.KnowledgeTag
	for _, knowledgeTag := range knowledgeTagListResp.KnowledgeTagList {
		knowledgeTagList = append(knowledgeTagList, &response.KnowledgeTag{
			TagId:    knowledgeTag.TagId,
			TagName:  knowledgeTag.TagName,
			Selected: knowledgeTag.Selected,
		})
	}
	return &response.KnowledgeTagListResp{
		KnowledgeTagList: knowledgeTagList,
	}
}
