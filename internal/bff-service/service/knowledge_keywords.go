package service

import (
	knowledgebase_keywords_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-keywords-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/gin-gonic/gin"
)

func GetKnowledgeKeywordsList(ctx *gin.Context, userId, orgId string, r *request.ListKeywordsReq) (*response.GetKnowledgeKeywordListResp, error) {
	resp, err := knowledgeBaseKeywords.GetKnowledgeKeywordsList(ctx.Request.Context(), &knowledgebase_keywords_service.GetKnowledgeKeywordsListReq{
		PageSize: int32(r.PageSize),
		PageNum:  int32(r.PageNum),
		Name:     r.Name,
		Identity: &knowledgebase_keywords_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return &response.GetKnowledgeKeywordListResp{}, err
	}
	var list []*response.KeywordsInfo
	for _, v := range resp.Keywords {
		keyword := &response.KeywordsInfo{
			Id:                 v.Id,
			Name:               v.Name,
			Alias:              v.Alias,
			KnowledgeBaseIds:   v.KnowledgeBaseIds,
			KnowledgeBaseNames: v.KnowledgeBaseNames,
			UpdatedAt:          v.UpdatedAt,
		}
		list = append(list, keyword)
	}
	return &response.GetKnowledgeKeywordListResp{
		List:     list,
		Total:    resp.Total,
		PageNum:  resp.PageNum,
		PageSize: resp.PageSize,
	}, nil
}

func CreateKnowledgeKeywords(ctx *gin.Context, userId, orgId string, r *request.CreateKeywordsReq) error {
	_, err := knowledgeBaseKeywords.CreateKnowledgeKeywords(ctx.Request.Context(), &knowledgebase_keywords_service.CreateKnowledgeKeywordsReq{
		Name:             r.Name,
		Alias:            r.Alias,
		KnowledgeBaseIds: r.KnowledgeBaseIds,
		Identity: &knowledgebase_keywords_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func GetKnowledgeKeywordsDetail(ctx *gin.Context, userId, orgId string, r *request.GetKeywordsDetailReq) (*response.KeywordsInfo, error) {
	resp, err := knowledgeBaseKeywords.GetKnowledgeKeywordsDetail(ctx.Request.Context(), &knowledgebase_keywords_service.GetKnowledgeKeywordsDetailReq{
		Id: r.Id,
		Identity: &knowledgebase_keywords_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return &response.KeywordsInfo{}, err
	}
	return &response.KeywordsInfo{
		Id:                 resp.Detail.Id,
		Name:               resp.Detail.Name,
		Alias:              resp.Detail.Alias,
		KnowledgeBaseIds:   resp.Detail.KnowledgeBaseIds,
		KnowledgeBaseNames: resp.Detail.KnowledgeBaseNames,
		UpdatedAt:          resp.Detail.UpdatedAt,
	}, nil
}

func UpdateKnowledgeKeywords(ctx *gin.Context, userId, orgId string, r *request.UpdateKeywordsReq) error {
	_, err := knowledgeBaseKeywords.UpdateKnowledgeKeywords(ctx.Request.Context(), &knowledgebase_keywords_service.UpdateKnowledgeKeywordsReq{
		Id: r.Id,
		Detail: &knowledgebase_keywords_service.CreateKnowledgeKeywordsReq{
			Name:             r.Name,
			Alias:            r.Alias,
			KnowledgeBaseIds: r.KnowledgeBaseIds,
			Identity: &knowledgebase_keywords_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func DeleteDocCategoryKeywords(ctx *gin.Context, userId, orgId string, r *request.DeleteKeywordsReq) error {
	// 删除前校验关键词归属：detail 接口已按 (org,user) 鉴权，非本租户的关键词会返回错误，
	// 以此拦截跨租户删除（DeleteKnowledgeKeywords 请求自身不携带调用方身份）。
	if _, err := knowledgeBaseKeywords.GetKnowledgeKeywordsDetail(ctx.Request.Context(), &knowledgebase_keywords_service.GetKnowledgeKeywordsDetailReq{
		Id: r.Id,
		Identity: &knowledgebase_keywords_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	}); err != nil {
		return err
	}
	_, err := knowledgeBaseKeywords.DeleteKnowledgeKeywords(ctx.Request.Context(), &knowledgebase_keywords_service.DeleteKnowledgeKeywordsReq{
		Id: r.Id,
	})
	if err != nil {
		return err
	}
	return nil
}
