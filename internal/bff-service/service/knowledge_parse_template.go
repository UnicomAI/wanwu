package service

import (
	knowledgebase_doc_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-doc-service"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	knowledgebase_template_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-template-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/gin-gonic/gin"
)

func GetParseTemplateList(ctx *gin.Context, userId, orgId string, r *request.ParseTemplateListReq) (*response.ParseTemplateListResp, error) {
	resp, err := knowledgeBaseTemplate.GetParseTemplateList(ctx.Request.Context(), &knowledgebase_template_service.GetParseTemplateListReq{
		DocType:  r.DocType,
		Identity: buildTemplateIdentity(userId, orgId),
	})
	if err != nil {
		return nil, err
	}
	var list []*response.ParseTemplateInfo
	for _, template := range resp.Templates {
		list = append(list, buildParseTemplateInfo(template))
	}
	return &response.ParseTemplateListResp{List: list}, nil
}

func CreateParseTemplate(ctx *gin.Context, userId, orgId string, r *request.CreateParseTemplateReq) error {
	_, err := knowledgeBaseTemplate.CreateParseTemplate(ctx.Request.Context(), &knowledgebase_template_service.CreateParseTemplateReq{
		Name:     r.Name,
		DocType:  r.DocType,
		Config:   buildParseTemplateConfig(&r.ParseTemplateConfig),
		Identity: buildTemplateIdentity(userId, orgId),
	})
	return err
}

func UpdateParseTemplate(ctx *gin.Context, userId, orgId string, r *request.UpdateParseTemplateReq) error {
	_, err := knowledgeBaseTemplate.UpdateParseTemplate(ctx.Request.Context(), &knowledgebase_template_service.UpdateParseTemplateReq{
		TemplateId: r.TemplateId,
		Name:       r.Name,
		Config:     buildParseTemplateConfig(&r.ParseTemplateConfig),
		Identity:   buildTemplateIdentity(userId, orgId),
	})
	return err
}

func DeleteParseTemplate(ctx *gin.Context, userId, orgId string, r *request.DeleteParseTemplateReq) error {
	_, err := knowledgeBaseTemplate.DeleteParseTemplate(ctx.Request.Context(), &knowledgebase_template_service.DeleteParseTemplateReq{
		TemplateId: r.TemplateId,
		Identity:   buildTemplateIdentity(userId, orgId),
	})
	return err
}

func buildTemplateIdentity(userId, orgId string) *knowledgebase_template_service.Identity {
	return &knowledgebase_template_service.Identity{UserId: userId, OrgId: orgId}
}

func buildParseTemplateConfig(config *request.ParseTemplateConfig) *knowledgebase_template_service.ParseTemplateConfig {
	result := &knowledgebase_template_service.ParseTemplateConfig{
		DocAnalyzer:       config.DocAnalyzer,
		OcrModelId:        config.ParserModelId,
		DocPreprocess:     config.DocPreprocess,
		AsrModelId:        config.AsrModelId,
		MultimodalModelId: config.MultimodalModelId,
	}
	if segment := config.DocSegment; segment != nil {
		result.DocSegment = &knowledgebase_template_service.ParseTemplateSegment{
			SegmentMethod:  segment.SegmentMethod,
			SegmentType:    segment.SegmentType,
			Splitter:       segment.Splitter,
			MaxSplitter:    int32(segment.MaxSplitter),
			Overlap:        segment.Overlap,
			SubSplitter:    segment.SubSplitter,
			SubMaxSplitter: int32(segment.SubMaxSplitter),
		}
	}
	return result
}

func buildParseTemplateInfo(template *knowledgebase_template_service.ParseTemplateInfo) *response.ParseTemplateInfo {
	config := template.GetConfig()
	segment := config.GetDocSegment()
	return &response.ParseTemplateInfo{
		TemplateId: template.TemplateId,
		Name:       template.Name,
		DocType:    template.DocType,
		DocSegment: &response.DocSegmentParam{
			SegmentMethod:  segment.GetSegmentMethod(),
			SegmentType:    segment.GetSegmentType(),
			Splitter:       segment.GetSplitter(),
			MaxSplitter:    int(segment.GetMaxSplitter()),
			Overlap:        segment.GetOverlap(),
			SubSplitter:    segment.GetSubSplitter(),
			SubMaxSplitter: int(segment.GetSubMaxSplitter()),
		},
		DocAnalyzer:       config.GetDocAnalyzer(),
		ParserModelId:     config.GetOcrModelId(),
		AsrModelId:        config.GetAsrModelId(),
		MultimodalModelId: config.GetMultimodalModelId(),
		DocPreprocess:     config.GetDocPreprocess(),
		UpdatedAt:         template.UpdatedAt,
		CreatedAt:         template.CreatedAt,
		BuiltIn:           template.BuiltIn,
	}
}

func buildKnowledgeBinds(binds []*request.ParseTemplateBind) []*knowledgebase_service.ParseTemplateBind {
	var result []*knowledgebase_service.ParseTemplateBind
	for _, bind := range binds {
		result = append(result, &knowledgebase_service.ParseTemplateBind{
			DocType:    bind.DocType,
			TemplateId: bind.TemplateId,
		})
	}
	return result
}

func buildDocTemplateBinds(binds []*request.ParseTemplateBind) []*knowledgebase_doc_service.ParseTemplateBind {
	var result []*knowledgebase_doc_service.ParseTemplateBind
	for _, bind := range binds {
		result = append(result, &knowledgebase_doc_service.ParseTemplateBind{
			DocType:    bind.DocType,
			TemplateId: bind.TemplateId,
		})
	}
	return result
}

func buildTemplateBindResp(binds []*knowledgebase_service.ParseTemplateBind) []*response.ParseTemplateBind {
	var result []*response.ParseTemplateBind
	for _, bind := range binds {
		result = append(result, &response.ParseTemplateBind{
			DocType:    bind.DocType,
			TemplateId: bind.TemplateId,
		})
	}
	return result
}
