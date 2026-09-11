package knowledge_doc

import (
	"context"
	"encoding/json"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_doc_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-doc-service"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/server/grpc/knowledge_parse_template"
	"github.com/UnicomAI/wanwu/pkg/log"
)

// importDocByTemplate 套用解析模板导入：同一文档类型共用一个模板，因此按文档类型拆成多条导入任务
func importDocByTemplate(ctx context.Context, req *knowledgebase_doc_service.ImportDocReq) error {
	// 1.绑定转 map 后校验文档类型合法、模板属于当前用户，并丢掉绑到内置模板的项
	//    校验时已批量查出模板对象，展开任务时直接复用，避免同批模板二次查库
	bindMap, customTemplates, err := knowledge_parse_template.ValidateParseTemplateBinds(ctx, req.UserId, req.OrgId, buildTemplateBindMap(req.ParseTemplate))
	if err != nil {
		return err
	}
	// 2.勾了覆盖要先确认是知识库拥有者，绑定只有拥有者能改
	if req.OverrideTemplate && !orm.IsKnowledgeOwner(ctx, req.KnowledgeId, req.UserId, req.OrgId) {
		log.Errorf("非知识库(%v)拥有者不能覆盖解析模板绑定", req.KnowledgeId)
		return util.ErrCode(errs.Code_KnowledgeParseTemplateOverrideDenied)
	}
	// 3.按文档类型分组，归不了类的压缩包落到空类型一组
	groups := groupDocByType(req.DocInfoList)
	// 4.媒体类型没有内置模板兜底，本次传了这些文件就必须选模板
	if err := checkMediaTemplate(groups, bindMap); err != nil {
		return err
	}
	// 5.构造导入任务
	tasks := make([]*model.KnowledgeImportTask, 0, len(groups))
	for docType, docList := range groups {
		task, err := buildTemplateImportTask(ctx, req, bindMap, customTemplates, docType, docList)
		if err != nil {
			return err
		}
		tasks = append(tasks, task)
	}
	// 6.逐条建导入任务
	for _, task := range tasks {
		if err := orm.CreateKnowledgeImportTask(ctx, task); err != nil {
			log.Errorf("import doc by template fail %v", err)
			return util.ErrCode(errs.Code_KnowledgeDocImportFail)
		}
	}
	// 7.勾了覆盖则把本次选择写回自己的绑定
	if req.OverrideTemplate {
		if err := overrideKnowledgeTemplate(ctx, req.KnowledgeId, req.UserId, req.OrgId, bindMap); err != nil {
			return err
		}
	}
	return nil
}

// overrideKnowledgeTemplate 把本次上传选定的模板写回自己在该知识库上的绑定，不动别人的
func overrideKnowledgeTemplate(ctx context.Context, knowledgeId, userId, orgId string, bindMap map[string]string) error {
	binds := make([]*model.KnowledgeParseTemplateBind, 0, len(bindMap))
	for docType, templateId := range bindMap {
		binds = append(binds, &model.KnowledgeParseTemplateBind{
			KnowledgeId: knowledgeId,
			DocType:     docType,
			UserId:      userId,
			OrgId:       orgId,
			TemplateId:  templateId,
		})
	}
	if err := orm.ReplaceKnowledgeParseTemplate(ctx, knowledgeId, userId, orgId, binds); err != nil {
		log.Errorf("override knowledge parse template fail %v", err)
		return util.ErrCode(errs.Code_KnowledgeDocImportFail)
	}
	return nil
}

// checkMediaTemplate 视频/音频/图片没有内置（默认）模板兜底，不选模板会按纯文字提取解析出空内容
// 压缩包归不到文档类型，落在空类型一组，不受此限
func checkMediaTemplate(groups map[string][]*knowledgebase_doc_service.DocFileInfo, bindMap map[string]string) error {
	for docType := range groups {
		if !model.IsMediaDocType(docType) {
			continue
		}
		if len(bindMap[docType]) == 0 {
			log.Errorf("文档类型(%v)未选择解析模板", docType)
			return util.ErrCode(errs.Code_KnowledgeParseTemplateMediaMissing)
		}
	}
	return nil
}

// buildTemplateBindMap 按文档类型建查找表，同一类型重复下发时后者覆盖前者
func buildTemplateBindMap(binds []*knowledgebase_doc_service.ParseTemplateBind) map[string]string {
	result := make(map[string]string, len(binds))
	for _, bind := range binds {
		result[bind.DocType] = bind.TemplateId
	}
	return result
}

func groupDocByType(docInfoList []*knowledgebase_doc_service.DocFileInfo) map[string][]*knowledgebase_doc_service.DocFileInfo {
	groups := make(map[string][]*knowledgebase_doc_service.DocFileInfo)
	for _, docInfo := range docInfoList {
		docType := model.DocTypeByExt(docInfo.DocType)
		groups[docType] = append(groups[docType], docInfo)
	}
	return groups
}

// buildTemplateImportTask 取该文档类型选定的模板展开成导入配置，没有模板则按内置默认配置
func buildTemplateImportTask(ctx context.Context, req *knowledgebase_doc_service.ImportDocReq, bindMap map[string]string, customTemplates map[string]*model.KnowledgeParseTemplate, docType string, docList []*knowledgebase_doc_service.DocFileInfo) (*model.KnowledgeImportTask, error) {
	groupReq := cloneImportReq(req, docList)
	// docType 为空是压缩包等归不了类的文件，先按默认参数建任务，解压后再逐个按类型套模板
	if len(docType) > 0 {
		template, err := resolveTemplate(ctx, req, bindMap, customTemplates, docType)
		if err != nil {
			return nil, err
		}
		// template 为 nil 表示既没选模板、也没存过内置模板，走内置默认配置
		if template != nil {
			if err := applyTemplateConfig(groupReq, template); err != nil {
				log.Errorf("apply parse template(%v) fail %v", template.TemplateId, err)
				return nil, util.ErrCode(errs.Code_KnowledgeDocImportFail)
			}
		}
		// 音频没有 ASR 模型转写不出内容，在这里拦截
		if docType == "audio" && len(groupReq.AsrModelId) == 0 {
			log.Errorf("音频解析模板未配置ASR模型 knowledgeId(%v)", req.KnowledgeId)
			return nil, util.ErrCode(errs.Code_KnowledgeParseTemplateAsrMissing)
		}
	}
	return buildImportTask(groupReq)
}

// resolveTemplate 优先用该类型校验通过的自定义模板，没绑定则回落到用户存过的内置（默认）模板
// 自定义模板在 ValidateParseTemplateBinds 里已按用户批量查出，这里直接复用，不再二次查库
func resolveTemplate(ctx context.Context, req *knowledgebase_doc_service.ImportDocReq, bindMap map[string]string, customTemplates map[string]*model.KnowledgeParseTemplate, docType string) (*model.KnowledgeParseTemplate, error) {
	templateId := bindMap[docType]
	if len(templateId) > 0 {
		template, ok := customTemplates[docType]
		if !ok {
			log.Errorf("解析模板(%v)不在校验结果中", templateId)
			return nil, util.ErrCode(errs.Code_KnowledgeParseTemplateNotExist)
		}
		return template, nil
	}
	template, err := orm.GetBuiltInParseTemplate(ctx, req.UserId, req.OrgId, docType)
	if err != nil {
		log.Errorf("get built-in parse template(%v) fail %v", docType, err)
		return nil, util.ErrCode(errs.Code_KnowledgeParseTemplateListFailed)
	}
	return template, nil
}

func cloneImportReq(req *knowledgebase_doc_service.ImportDocReq, docList []*knowledgebase_doc_service.DocFileInfo) *knowledgebase_doc_service.ImportDocReq {
	return &knowledgebase_doc_service.ImportDocReq{
		UserId:          req.UserId,
		OrgId:           req.OrgId,
		KnowledgeId:     req.KnowledgeId,
		DocImportType:   req.DocImportType,
		DocInfoList:     docList,
		DocMetaDataList: req.DocMetaDataList,
		ParseTemplate:   req.ParseTemplate,
		DocSegment:      buildProtoDocSegment(model.DefaultSegmentConfig()),
		DocAnalyzer:     model.DefaultDocAnalyzer().AnalyzerList,
		DocPreprocess:   model.DefaultDocPreProcess().PreProcessList,
	}
}

func buildProtoDocSegment(segment *model.SegmentConfig) *knowledgebase_doc_service.DocSegment {
	return &knowledgebase_doc_service.DocSegment{
		SegmentMethod:  segment.SegmentMethod,
		SegmentType:    segment.SegmentType,
		Splitter:       segment.Splitter,
		MaxSplitter:    int32(segment.MaxSplitter),
		Overlap:        segment.Overlap,
		SubSplitter:    segment.SubSplitter,
		SubMaxSplitter: int32(segment.SubMaxSplitter),
	}
}

func applyTemplateConfig(req *knowledgebase_doc_service.ImportDocReq, template *model.KnowledgeParseTemplate) error {
	segment := &model.SegmentConfig{}
	if err := json.Unmarshal([]byte(template.SegmentConfig), segment); err != nil {
		return err
	}
	analyzer := &model.DocAnalyzer{}
	if err := json.Unmarshal([]byte(template.DocAnalyzer), analyzer); err != nil {
		return err
	}
	preProcess := &model.DocPreProcess{}
	if len(template.DocPreProcess) > 0 {
		if err := json.Unmarshal([]byte(template.DocPreProcess), preProcess); err != nil {
			return err
		}
	}
	req.DocSegment = buildProtoDocSegment(segment)
	req.DocAnalyzer = analyzer.AnalyzerList
	req.AsrModelId = analyzer.AsrModelId
	req.MultimodalModelId = analyzer.MultimodalModelId
	req.OcrModelId = template.OcrModelId
	req.DocPreprocess = preProcess.PreProcessList
	return nil
}
