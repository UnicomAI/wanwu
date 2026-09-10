package knowledge_parse_template

import (
	"context"
	"encoding/json"
	"errors"
	"slices"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_template_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-template-service"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/log"
	wanwu_util "github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// GetParseTemplateList 查询解析模板列表
func (s *Service) GetParseTemplateList(ctx context.Context, req *knowledgebase_template_service.GetParseTemplateListReq) (*knowledgebase_template_service.GetParseTemplateListResp, error) {
	// 1.查询目前已有模板列表
	templateList, err := orm.GetParseTemplateList(ctx, req.Identity.UserId, req.Identity.OrgId, req.DocType)
	if err != nil {
		log.Errorf("GetParseTemplateList 失败(%v) 参数(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeParseTemplateListFailed)
	}
	// 2.若无内置模板则新建
	missing := missingBuiltInDocTypes(templateList, builtInDocTypes(req.DocType))
	if len(missing) > 0 {
		created, err := orm.CreateBuiltInParseTemplates(ctx, req.Identity.UserId, req.Identity.OrgId, missing)
		if err != nil {
			log.Errorf("CreateBuiltInParseTemplates 失败(%v) 参数(%v)", err, req)
			return nil, util.ErrCode(errs.Code_KnowledgeParseTemplateListFailed)
		}
		templateList = append(templateList, created...)
	}
	// 3.构造返回列表
	var templates []*knowledgebase_template_service.ParseTemplateInfo
	for _, template := range templateList {
		info, err := buildParseTemplateInfo(template)
		if err != nil {
			log.Errorf("GetParseTemplateList 模板(%v)解析失败(%v)", template.TemplateId, err)
			return nil, util.ErrCode(errs.Code_KnowledgeParseTemplateListFailed)
		}
		templates = append(templates, info)
	}
	return &knowledgebase_template_service.GetParseTemplateListResp{Templates: templates}, nil
}

// CreateParseTemplate 新建解析模板
func (s *Service) CreateParseTemplate(ctx context.Context, req *knowledgebase_template_service.CreateParseTemplateReq) (*emptypb.Empty, error) {
	// 1.检查有无重名模板
	if err := orm.CheckRepeatedParseTemplate(ctx, req.Identity.UserId, req.Identity.OrgId, req.DocType, req.Name, ""); err != nil {
		return nil, buildTemplateErr(err, errs.Code_KnowledgeParseTemplateCreateFailed)
	}
	// 1.1 按文档类型校验模板配置：音频必须配 ASR 模型，否则转写不出内容（图片/视频无模型要求）
	if err := checkTemplateMediaModel(req.DocType, req.Config); err != nil {
		return nil, err
	}
	// 2.构造新模板
	template, err := buildCreateTemplateModel(req)
	if err != nil {
		log.Errorf("CreateParseTemplate 参数(%v)构造失败(%v)", req, err)
		return nil, util.ErrCode(errs.Code_KnowledgeParseTemplateCreateFailed)
	}
	// 3.插入新模板
	if err := orm.CreateParseTemplate(ctx, template); err != nil {
		log.Errorf("CreateParseTemplate 失败(%v) 参数(%v)", err, req)
		return nil, util.ErrCode(errs.Code_KnowledgeParseTemplateCreateFailed)
	}
	return &emptypb.Empty{}, nil
}

// UpdateParseTemplate 编辑解析模板
func (s *Service) UpdateParseTemplate(ctx context.Context, req *knowledgebase_template_service.UpdateParseTemplateReq) (*emptypb.Empty, error) {
	// 1.获取原有模板信息
	old, err := orm.GetParseTemplate(ctx, req.Identity.UserId, req.Identity.OrgId, req.TemplateId)
	if err != nil {
		return nil, buildTemplateErr(err, errs.Code_KnowledgeParseTemplateUpdateFailed)
	}
	// 1.1 内置（默认）模板只读，禁止 API 修改，防止兜底的默认配置被悄悄改掉
	if old.BuiltIn == 1 {
		log.Errorf("内置模板(%v)不支持修改", req.TemplateId)
		return nil, util.ErrCode(errs.Code_KnowledgeParseTemplateUpdateFailed)
	}
	// 1.2 按文档类型校验模板配置：音频必须配 ASR 模型（文档类型不可改，用原模板的类型判断）
	if err := checkTemplateMediaModel(old.DocType, req.Config); err != nil {
		return nil, err
	}
	// 2.检查有无重名模板
	if err := orm.CheckRepeatedParseTemplate(ctx, req.Identity.UserId, req.Identity.OrgId, old.DocType, req.Name, req.TemplateId); err != nil {
		return nil, buildTemplateErr(err, errs.Code_KnowledgeParseTemplateUpdateFailed)
	}
	// 3.构造新模板
	template, err := buildUpdateTemplateModel(req)
	if err != nil {
		log.Errorf("UpdateParseTemplate 参数(%v)构造失败(%v)", req, err)
		return nil, util.ErrCode(errs.Code_KnowledgeParseTemplateUpdateFailed)
	}
	// 4.更新模板
	if err := orm.UpdateParseTemplate(ctx, req.Identity.UserId, req.Identity.OrgId, template); err != nil {
		log.Errorf("UpdateParseTemplate 失败(%v) 参数(%v)", err, req)
		return nil, buildTemplateErr(err, errs.Code_KnowledgeParseTemplateUpdateFailed)
	}
	return &emptypb.Empty{}, nil
}

// DeleteParseTemplate 删除解析模板
func (s *Service) DeleteParseTemplate(ctx context.Context, req *knowledgebase_template_service.DeleteParseTemplateReq) (*emptypb.Empty, error) {
	// 1.先取模板：内置（默认）模板只读，禁止 API 删除（删了下次查列表又会重建出默认配置，删了没有意义）
	old, err := orm.GetParseTemplate(ctx, req.Identity.UserId, req.Identity.OrgId, req.TemplateId)
	if err != nil {
		return nil, buildTemplateErr(err, errs.Code_KnowledgeParseTemplateDeleteFailed)
	}
	if old.BuiltIn == 1 {
		log.Errorf("内置模板(%v)不支持删除", req.TemplateId)
		return nil, util.ErrCode(errs.Code_KnowledgeParseTemplateDeleteFailed)
	}
	// 2.删除模板并解除知识库上的绑定
	if err := orm.DeleteParseTemplate(ctx, req.Identity.UserId, req.Identity.OrgId, req.TemplateId); err != nil {
		log.Errorf("DeleteParseTemplate 失败(%v) 参数(%v)", err, req)
		return nil, buildTemplateErr(err, errs.Code_KnowledgeParseTemplateDeleteFailed)
	}
	return &emptypb.Empty{}, nil
}

// missingBuiltInDocTypes 对比已查出的模板列表，返回还没有内置模板的文档类型
func missingBuiltInDocTypes(templateList []*model.KnowledgeParseTemplate, needBuiltInDocTypes []string) []string {
	builtInDocTypes := make([]string, 0, len(templateList))
	for _, template := range templateList {
		if template.BuiltIn == 1 {
			builtInDocTypes = append(builtInDocTypes, template.DocType)
		}
	}
	var missingDocTypes []string
	for _, docType := range needBuiltInDocTypes {
		if !slices.Contains(builtInDocTypes, docType) {
			missingDocTypes = append(missingDocTypes, docType)
		}
	}
	return missingDocTypes
}

// builtInDocTypes 指定文档类型时只补该类型，查全部时补齐所有类型；媒体类型不补
func builtInDocTypes(docType string) []string {
	if len(docType) == 0 {
		return model.BuiltInDocTypes
	}
	if !slices.Contains(model.BuiltInDocTypes, docType) {
		return nil
	}
	return []string{docType}
}

// buildTemplateErr 重名与不存在给专用错误码，其余归到调用方传入的默认码
func buildTemplateErr(err error, defaultCode errs.Code) error {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return util.ErrCode(errs.Code_KnowledgeParseTemplateDuplicateName)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return util.ErrCode(errs.Code_KnowledgeParseTemplateNotExist)
	}
	return util.ErrCode(defaultCode)
}

// buildCreateTemplateModel 新建的模板，模板 id 在这里生成
func buildCreateTemplateModel(req *knowledgebase_template_service.CreateParseTemplateReq) (*model.KnowledgeParseTemplate, error) {
	template := &model.KnowledgeParseTemplate{
		TemplateId: wanwu_util.GenUUID(),
		Name:       req.Name,
		DocType:    req.DocType,
		UserId:     req.Identity.UserId,
		OrgId:      req.Identity.OrgId,
	}
	if err := fillParseTemplateConfig(template, req.Config); err != nil {
		return nil, err
	}
	return template, nil
}

// buildUpdateTemplateModel 编辑的模板，只带 orm 白名单会更新的字段
func buildUpdateTemplateModel(req *knowledgebase_template_service.UpdateParseTemplateReq) (*model.KnowledgeParseTemplate, error) {
	template := &model.KnowledgeParseTemplate{
		TemplateId: req.TemplateId,
		Name:       req.Name,
	}
	if err := fillParseTemplateConfig(template, req.Config); err != nil {
		return nil, err
	}
	return template, nil
}

// buildSegmentConfig 分段配置：proto 转落库结构
func buildSegmentConfig(segment *knowledgebase_template_service.ParseTemplateSegment) *model.SegmentConfig {
	return &model.SegmentConfig{
		SegmentMethod:  segment.SegmentMethod,
		SegmentType:    segment.SegmentType,
		Splitter:       segment.Splitter,
		MaxSplitter:    int(segment.MaxSplitter),
		Overlap:        segment.Overlap,
		SubSplitter:    segment.SubSplitter,
		SubMaxSplitter: int(segment.SubMaxSplitter),
	}
}

// buildParseTemplateSegment 分段配置：落库结构转 proto
func buildParseTemplateSegment(segment *model.SegmentConfig) *knowledgebase_template_service.ParseTemplateSegment {
	return &knowledgebase_template_service.ParseTemplateSegment{
		SegmentMethod:  segment.SegmentMethod,
		SegmentType:    segment.SegmentType,
		Splitter:       segment.Splitter,
		MaxSplitter:    int32(segment.MaxSplitter),
		Overlap:        segment.Overlap,
		SubSplitter:    segment.SubSplitter,
		SubMaxSplitter: int32(segment.SubMaxSplitter),
	}
}

// checkTemplateMediaModel 媒体类型模板的模型要求：音频必须配 ASR 模型，否则转写不出内容
// 图片/视频无模型硬性要求（按需用 OCR/multimodal 模型即可），与导入时的 149026 校验口径一致
func checkTemplateMediaModel(docType string, config *knowledgebase_template_service.ParseTemplateConfig) error {
	if docType == "audio" && config != nil && len(config.AsrModelId) == 0 {
		log.Errorf("音频解析模板未配置ASR模型")
		return util.ErrCode(errs.Code_KnowledgeParseTemplateAsrMissing)
	}
	return nil
}

// fillParseTemplateConfig 把解析配置序列化后填进模板，创建模板与更新模板共用该helper函数
func fillParseTemplateConfig(template *model.KnowledgeParseTemplate, config *knowledgebase_template_service.ParseTemplateConfig) error {
	segment := config.GetDocSegment()
	if segment == nil {
		return errors.New("parse template segment is nil")
	}
	segmentConfig, err := json.Marshal(buildSegmentConfig(segment))
	if err != nil {
		return err
	}
	docAnalyzer, err := json.Marshal(&model.DocAnalyzer{
		AnalyzerList:      config.DocAnalyzer,
		AsrModelId:        config.AsrModelId,
		MultimodalModelId: config.MultimodalModelId,
	})
	if err != nil {
		return err
	}
	docPreProcess, err := json.Marshal(&model.DocPreProcess{PreProcessList: config.DocPreprocess})
	if err != nil {
		return err
	}
	template.SegmentConfig = string(segmentConfig)
	template.DocAnalyzer = string(docAnalyzer)
	template.DocPreProcess = string(docPreProcess)
	template.OcrModelId = config.OcrModelId
	return nil
}

func buildParseTemplateInfo(template *model.KnowledgeParseTemplate) (*knowledgebase_template_service.ParseTemplateInfo, error) {
	segment := &model.SegmentConfig{}
	if err := json.Unmarshal([]byte(template.SegmentConfig), segment); err != nil {
		return nil, err
	}
	analyzer := &model.DocAnalyzer{}
	if err := json.Unmarshal([]byte(template.DocAnalyzer), analyzer); err != nil {
		return nil, err
	}
	preProcess := &model.DocPreProcess{}
	if err := json.Unmarshal([]byte(template.DocPreProcess), preProcess); err != nil {
		return nil, err
	}
	return &knowledgebase_template_service.ParseTemplateInfo{
		TemplateId: template.TemplateId,
		Name:       template.Name,
		DocType:    template.DocType,
		Config: &knowledgebase_template_service.ParseTemplateConfig{
			DocSegment:        buildParseTemplateSegment(segment),
			DocAnalyzer:       analyzer.AnalyzerList,
			OcrModelId:        template.OcrModelId,
			DocPreprocess:     preProcess.PreProcessList,
			AsrModelId:        analyzer.AsrModelId,
			MultimodalModelId: analyzer.MultimodalModelId,
		},
		UpdatedAt: wanwu_util.Time2Str(template.UpdatedAt),
		CreatedAt: wanwu_util.Time2Str(template.CreatedAt),
		BuiltIn:   template.BuiltIn == 1,
	}, nil
}

// validDocTypes 文档类型白名单，避免脏 key 落库
var validDocTypes = func() map[string]struct{} {
	result := make(map[string]struct{}, len(model.ParseTemplateDocTypes))
	for _, docType := range model.ParseTemplateDocTypes {
		result[docType] = struct{}{}
	}
	return result
}()

// ValidateParseTemplateBinds 校验文档类型合法、模板属于当前用户且与所绑文档类型一致
// 入参用 map 表达绑定，同类型天然只保留一条
// 返回值只留自定义模板的绑定：绑内置和不绑都取内置配置，落库只会让详情接口多出一堆内置id
// 顺带把 docType→自定义模板对象 一并返回，导入展开阶段直接复用，避免同一批绑定被二次查库
func ValidateParseTemplateBinds(ctx context.Context, userId, orgId string, bindMap map[string]string) (map[string]string, map[string]*model.KnowledgeParseTemplate, error) {
	boundIds := make([]string, 0, len(bindMap))
	for docType, templateId := range bindMap {
		if _, ok := validDocTypes[docType]; !ok {
			log.Errorf("解析模板绑定的文档类型(%v)非法", docType)
			return nil, nil, util.ErrCode(errs.Code_KnowledgeParseTemplateNotExist)
		}
		if len(templateId) > 0 {
			boundIds = append(boundIds, templateId)
		}
	}
	templateMap, err := orm.GetParseTemplateMap(ctx, userId, orgId, boundIds)
	if err != nil {
		log.Errorf("校验解析模板归属失败(%v) 参数(%v)", err, boundIds)
		return nil, nil, util.ErrCode(errs.Code_KnowledgeParseTemplateListFailed)
	}
	customBinds := make(map[string]string, len(bindMap))
	customTemplates := make(map[string]*model.KnowledgeParseTemplate, len(bindMap))
	for docType, templateId := range bindMap {
		if len(templateId) == 0 {
			continue
		}
		template, ok := templateMap[templateId]
		if !ok {
			log.Errorf("解析模板(%v)不属于当前用户", templateId)
			return nil, nil, util.ErrCode(errs.Code_KnowledgeParseTemplateNotExist)
		}
		// 模板按文档类型划分，绑到别的类型上会用错解析参数
		if template.DocType != docType {
			log.Errorf("解析模板(%v)的文档类型(%v)与绑定的类型(%v)不符", templateId, template.DocType, docType)
			return nil, nil, util.ErrCode(errs.Code_KnowledgeParseTemplateNotExist)
		}
		if template.BuiltIn == 1 {
			continue
		}
		customBinds[docType] = templateId
		customTemplates[docType] = template
	}
	return customBinds, customTemplates, nil
}
