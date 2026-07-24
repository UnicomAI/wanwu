package service

import (
	"path/filepath"
	"regexp"
	"strings"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

// segmentImageMarkdownRegexp 匹配分段内容中的 markdown 图片语法 ![alt](url) —— 与前端 Md2Img 保持一致
var segmentImageMarkdownRegexp = regexp.MustCompile(`!\[(.*?)\]\(([^)\s]+)(?:\s+"([^"]*)")?\)`)

// segmentImageAllowExtSet openapi 分段图片上传允许的扩展名（对齐前端 uploadImgMd acceptType）
var segmentImageAllowExtSet = map[string]struct{}{
	".png":  {},
	".jpg":  {},
	".jpeg": {},
}

// CreateDocSegmentOpenapi openapi 新增单条文档切片（普通知识库不允许插入图片）
func CreateDocSegmentOpenapi(ctx *gin.Context, userId, orgId string, r *request.CreateDocSegmentReq) error {
	if err := checkSegmentImageContent(ctx, userId, orgId, r.DocId, r.Content); err != nil {
		return err
	}
	return CreateDocSegment(ctx, userId, orgId, r)
}

// UpdateDocSegmentOpenapi openapi 更新文档切片（多模态知识库可插入图片，普通知识库不行）
func UpdateDocSegmentOpenapi(ctx *gin.Context, userId, orgId string, r *request.UpdateDocSegmentReq) error {
	if err := checkSegmentImageContent(ctx, userId, orgId, r.DocId, r.Content); err != nil {
		return err
	}
	return UpdateDocSegment(ctx, userId, orgId, r)
}

// UploadDocSegmentImageOpenapi openapi 上传分段图片，返回 markdown 格式 url，仅多模态知识库可用
func UploadDocSegmentImageOpenapi(ctx *gin.Context, userId, orgId string, knowledgeId string) (*response.RagUploadResponse, error) {
	// 1.仅多模态知识库允许上传图片
	multimodal, err := isMultimodalKnowledge(ctx, userId, orgId, knowledgeId)
	if err != nil {
		return nil, err
	}
	if !multimodal {
		return nil, grpc_util.ErrorStatus(errs.Code_KnowledgeDocSegmentImageNotSupport)
	}
	// 2.校验上传文件为图片格式
	form, err := ctx.MultipartForm()
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_save", err.Error())
	}
	files := form.File["files"]
	if len(files) == 0 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_file_upload_check", "file is empty")
	}
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if _, ok := segmentImageAllowExtSet[ext]; !ok {
			return nil, grpc_util.ErrorStatus(errs.Code_KnowledgeDocSegmentImageTypeInvalid)
		}
	}
	// 3.复用 rag 上传，强制返回 markdown url
	return RagUpload(ctx, userId, orgId, request.RagUploadParams{Markdown: true})
}

// checkSegmentImageContent 普通知识库的分段内容不允许包含图片 markdown，仅多模态知识库可以
func checkSegmentImageContent(ctx *gin.Context, userId, orgId, docId, content string) error {
	// 无图片内容直接放行，避免多余的知识库查询
	if !segmentImageMarkdownRegexp.MatchString(content) {
		return nil
	}
	// 内容含图片：仅多模态知识库允许
	knowledgeId, err := searchKnowledgeIdByDocId(ctx, userId, orgId, docId)
	if err != nil {
		return err
	}
	multimodal, err := isMultimodalKnowledge(ctx, userId, orgId, knowledgeId)
	if err != nil {
		return err
	}
	if !multimodal {
		return grpc_util.ErrorStatus(errs.Code_KnowledgeDocSegmentImageNotSupport)
	}
	return nil
}

// searchKnowledgeIdByDocId 根据文档 id 查询所属知识库 id
func searchKnowledgeIdByDocId(ctx *gin.Context, userId, orgId, docId string) (string, error) {
	docInfo, err := GetDocDetail(ctx, userId, orgId, docId)
	if err != nil {
		return "", err
	}
	return docInfo.KnowledgeId, nil
}

// isMultimodalKnowledge 判断知识库是否为多模态知识库（category==2）
func isMultimodalKnowledge(ctx *gin.Context, userId, orgId, knowledgeId string) (bool, error) {
	knowledgeInfo, err := knowledgeBase.SelectKnowledgeDetailById(ctx.Request.Context(), &knowledgebase_service.KnowledgeDetailSelectReq{
		KnowledgeId: knowledgeId,
		UserId:      userId,
		OrgId:       orgId,
	})
	if err != nil {
		return false, err
	}
	return int(knowledgeInfo.Category) == request.CategoryMultimodalKnowledge, nil
}
