package orm

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm/sqlopt"
	async_task "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/async-task"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
	wanwu_util "github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

const (
	MetaValueTypeNumber = "number"
	MetaValueTypeString = "string"
	MetaValueTypeTime   = "time"
	MetaValueTypeDir1   = "dir_1"
	MetaValueTypeDir2   = "dir_2"
	MetaValueTypeDir3   = "dir_3"
	MetaValueTypeDir4   = "dir_4"
	MetaValueTypeDir5   = "dir_5"
	MetaValueTypeDir6   = "dir_6"
	MetaValueTypeDir7   = "dir_7"
	MetaValueTypeDir8   = "dir_8"
	MetaValueTypeDir9   = "dir_9"
	MetaValueTypeDir10  = "dir_10"
	PreprocessSymbol    = "replace_symbols"
	PreprocessLink      = "delete_links"
)

var fileDirTypeMap = map[string]int{
	MetaValueTypeDir1:  1,
	MetaValueTypeDir2:  2,
	MetaValueTypeDir3:  3,
	MetaValueTypeDir4:  4,
	MetaValueTypeDir5:  5,
	MetaValueTypeDir6:  6,
	MetaValueTypeDir7:  7,
	MetaValueTypeDir8:  8,
	MetaValueTypeDir9:  9,
	MetaValueTypeDir10: 10,
}

type KnowledgeGraph struct {
	KnowledgeGraphSwitch  bool   `json:"knowledgeGraphSwitch"`
	GraphModelId          string `json:"graphModelId"`
	GraphSchemaObjectName string `json:"graphSchemaObjectName"`
	GraphSchemaFileName   string `json:"graphSchemaFileName"`
}

// GetDocList 查询知识库文件列表
func GetDocList(ctx context.Context, userId, orgId, knowledgeId, name, tag string,
	statusList []int, graphStatusList []int, docIdList []string, pageSize int32, pageNum int32) ([]*model.KnowledgeDoc, int64, error) {
	tx := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId),
		sqlopt.WithKnowledgeID(knowledgeId),
		sqlopt.LikeName(name),
		sqlopt.LikeTag(tag),
		sqlopt.WithStatusList(statusList),
		sqlopt.WithGraphStatusList(graphStatusList),
		sqlopt.WithDocIDsNonEmpty(docIdList),
		sqlopt.WithDelete(0)).
		Apply(db.GetHandle(ctx), &model.KnowledgeDoc{})
	var total int64
	err := tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	limit := pageSize
	offset := pageSize * (pageNum - 1)
	var docList []*model.KnowledgeDoc
	err = tx.Order("create_at desc").Limit(int(limit)).Offset(int(offset)).Find(&docList).Error
	if err != nil {
		return nil, 0, err
	}
	return docList, total, nil
}

// GetDocDetail 查询知识库文件详情
func GetDocDetail(ctx context.Context, userId, orgId, docId string) (*model.KnowledgeDoc, error) {
	var doc = model.KnowledgeDoc{}
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId),
		sqlopt.WithDocID(docId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeDoc{}).First(&doc).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// GetDocListByKnowledgeIdNoDeleteCheck 根据知识库id查询知识库文件列表
func GetDocListByKnowledgeIdNoDeleteCheck(ctx context.Context, userId, orgId string, knowledgeId string) ([]*model.KnowledgeDoc, error) {
	var docList []*model.KnowledgeDoc
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithKnowledgeID(knowledgeId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeDoc{}).Find(&docList).Error
	if err != nil {
		return nil, err
	}
	return docList, nil
}

// GetDocListByKnowledgeId 根据知识库id查询知识库文件列表
func GetDocListByKnowledgeId(ctx context.Context, userId, orgId string, knowledgeId string) ([]*model.KnowledgeDoc, error) {
	var docList []*model.KnowledgeDoc
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithKnowledgeID(knowledgeId), sqlopt.WithDelete(0)).
		Apply(db.GetHandle(ctx), &model.KnowledgeDoc{}).Find(&docList).Error
	if err != nil {
		return nil, err
	}
	return docList, nil
}

// GetDocListByKnowledgeIdAndFileTypeFilter 根据知识库id查询知识库文件列表
func GetDocListByKnowledgeIdAndFileTypeFilter(ctx context.Context, userId, orgId string, knowledgeId string, fileType string) ([]*model.KnowledgeDoc, error) {
	var docList []*model.KnowledgeDoc
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithKnowledgeID(knowledgeId), sqlopt.WithDelete(0), sqlopt.WithFileTypeFilter(fileType)).
		Apply(db.GetHandle(ctx), &model.KnowledgeDoc{}).Find(&docList).Error
	if err != nil {
		return nil, err
	}
	return docList, nil
}

// GetDocListByIdListNoDeleteCheck 查询知识库文件列表
func GetDocListByIdListNoDeleteCheck(ctx context.Context, userId, orgId string, idList []uint32) ([]*model.KnowledgeDoc, error) {
	var docList []*model.KnowledgeDoc
	err := sqlopt.SQLOptions(sqlopt.WithPermit("", userId), sqlopt.WithIDs(idList)).
		Apply(db.GetHandle(ctx), &model.KnowledgeDoc{}).Find(&docList).Error
	if err != nil {
		return nil, err
	}
	return docList, nil
}

// CheckKnowledgeDocSameName 知识库文档同名校验
func CheckKnowledgeDocSameName(ctx context.Context, userId string, knowledgeId string, docName string, docUrl string, docId string) error {
	var count int64
	var docUrlMd5 = ""
	if len(docUrl) > 0 {
		docUrlMd5 = util.MD5(docUrl)
	}
	err := sqlopt.SQLOptions(sqlopt.WithPermit("", userId),
		sqlopt.WithKnowledgeID(knowledgeId),
		sqlopt.WithName(docName),
		sqlopt.WithFilePathMd5(docUrlMd5),
		sqlopt.WithoutStatus(model.DocFail),
		sqlopt.WithoutDocId(docId),
		sqlopt.WithDelete(0)).
		Apply(db.GetHandle(ctx), &model.KnowledgeDoc{}).
		Count(&count).Error
	if err != nil {
		log.Errorf("CheckKnowledgeDocSameName knowledgeId %s err: %v", knowledgeId, err)
		return errors.New("CheckKnowledgeDocSameName error")
	}
	if count > 0 {
		return errors.New("CheckKnowledgeDocSameName exist error")
	}
	return nil
}

// SelectDocByDocIdList 查询知识库文档信息
func SelectDocByDocIdList(ctx context.Context, docIdList []string, userId, orgId string) ([]*model.KnowledgeDoc, error) {
	var docList []*model.KnowledgeDoc
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithDocIDs(docIdList)).
		Apply(db.GetHandle(ctx), &model.KnowledgeDoc{}).
		Find(&docList).Error
	if err != nil {
		log.Errorf("SelectDocByDocId userId %s err: %v", userId, err)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseAccessDenied)
	}
	if len(docList) == 0 {
		log.Errorf("SelectDocByDocId userId %s doc list empty", userId)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseAccessDenied)
	}
	return docList, nil
}

// SelectDocStatusByDocIdList 查询知识库文档信息
func SelectDocStatusByDocIdList(ctx context.Context, docIdList []string, status int) (int64, error) {
	var count int64
	err := sqlopt.SQLOptions(sqlopt.WithDocIDs(docIdList), sqlopt.WithStatus(status)).
		Apply(db.GetHandle(ctx), &model.KnowledgeDoc{}).
		Count(&count).Error
	if err != nil {
		log.Errorf("SelectDocStatusByDocIdList err: %v", err)
		return 0, util.ErrCode(errs.Code_KnowledgeBaseAccessDenied)
	}
	return count, nil
}

// SelectDocByDocIdListAndFileTypeFilter 查询知识库文档信息
func SelectDocByDocIdListAndFileTypeFilter(ctx context.Context, docIdList []string, userId, orgId string, fileTypeFilter string) ([]*model.KnowledgeDoc, error) {
	var docList []*model.KnowledgeDoc
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithDocIDs(docIdList), sqlopt.WithFileTypeFilter(fileTypeFilter)).
		Apply(db.GetHandle(ctx), &model.KnowledgeDoc{}).
		Find(&docList).Error
	if err != nil {
		log.Errorf("SelectDocByDocId userId %s err: %v", userId, err)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseAccessDenied)
	}
	if len(docList) == 0 {
		log.Errorf("SelectDocByDocId userId %s doc list empty", userId)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseAccessDenied)
	}
	return docList, nil
}

func buildKnowledgeDocMeta(doc *model.KnowledgeDoc, importTask *model.KnowledgeImportTask, meta *model.KnowledgeDocMeta) (*model.KnowledgeDocMeta, error) {
	var valueType = meta.ValueType
	if fileDirType(meta.ValueType) {
		valueType = MetaValueTypeString
		path := doc.DirFilePath
		if len(path) == 0 {
			return nil, nil
		}
		value := buildFileDirTypeValue(meta.ValueType)
		//这里效率稍低，可以挪到外面分隔，但是为了代码逻辑清晰，还是保留
		split := strings.Split(path, string(filepath.Separator))
		if len(split) > 0 {
			split = split[:len(split)-1]
		}
		log.Infof("path  %s", path)
		for index, part := range split {
			log.Infof("index  %d, part  %s, value %d", index, part, value)
			if index == value {
				meta.ValueMain = part //给value赋值
				break
			}
		}
		//如果valueMain为空，标识没有赋值，则返回nil
		if len(meta.ValueMain) == 0 {
			return nil, nil
		}
	}
	return &model.KnowledgeDocMeta{
		KnowledgeId: importTask.KnowledgeId,
		MetaId:      wanwu_util.NewID(),
		DocId:       doc.DocId,
		Key:         meta.Key,
		ValueMain:   meta.ValueMain,
		ValueType:   valueType,
		Rule:        meta.Rule,
		UserId:      importTask.UserId,
		OrgId:       importTask.OrgId,
	}, nil
}

func fileDirType(valueType string) bool {
	return fileDirTypeMap[valueType] > 0
}

func buildFileDirTypeValue(valueType string) int {
	return fileDirTypeMap[valueType]
}

// CreateKnowledgeDoc 创建知识库文件
func CreateKnowledgeDoc(ctx context.Context, doc *model.KnowledgeDoc, importTask *model.KnowledgeImportTask) error {
	knowledge, err := SelectKnowledgeById(ctx, doc.KnowledgeId, "", "")
	if err != nil {
		return err
	}
	var config = &model.SegmentConfig{}
	err = json.Unmarshal([]byte(importTask.SegmentConfig), config)
	if err != nil {
		log.Errorf("SegmentConfig process error %s", err.Error())
		return err
	}
	var analyzer = &model.DocAnalyzer{}
	err = json.Unmarshal([]byte(importTask.DocAnalyzer), analyzer)
	if err != nil {
		log.Errorf("DocAnalyzer process error %s", err.Error())

		return err
	}
	var preProcess = &model.DocPreProcess{}
	if len(importTask.DocPreProcess) > 0 {
		err = json.Unmarshal([]byte(importTask.DocPreProcess), preProcess)
		if err != nil {
			log.Errorf("DocPreprocess process error %s", err.Error())
			return err
		}
		preProcess.PreProcessList = normalizeList(preProcess.PreProcessList)
	}

	_, objectName, _ := service.SplitFilePath(doc.FilePath)
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.插入数据
		err = createKnowledgeDoc(tx, doc)
		if err != nil {
			return err
		}
		ragMetaList, err := buildAndCreateMetaData(tx, importTask, doc)
		if err != nil {
			log.Errorf("buildAndCreateMetaData error %s", err.Error())
		}
		//非初始话状态的不需要rag 导入，因为有可能直接失败了
		if doc.Status != model.DocInit {
			return nil
		}
		//构造知识库图谱
		knowledgeGraph := BuildKnowledgeGraph(string(knowledge.KnowledgeGraph))
		//2.rag文档导入
		return service.RagImportDoc(ctx, &service.RagImportDocParams{
			DocId:                 doc.DocId,
			KnowledgeName:         knowledge.RagName,
			CategoryId:            knowledge.KnowledgeId,
			UserId:                knowledge.UserId,
			Overlap:               config.Overlap,
			SegmentSize:           config.MaxSplitter,
			SegmentType:           service.RebuildSegmentType(config.SegmentType, config.SegmentMethod),
			SplitType:             service.RebuildSplitType(config.SegmentMethod),
			Separators:            config.Splitter,
			ParserChoices:         analyzer.AnalyzerList,
			ObjectName:            objectName,
			OriginalName:          doc.Name,
			IsEnhanced:            "false",
			OcrModelId:            importTask.OcrModelId,
			PreProcess:            preProcess.PreProcessList,
			RagMetaDataParams:     ragMetaList,
			RagChildChunkConfig:   buildSubRagChunkConfig(config),
			KnowledgeGraphSwitch:  knowledgeGraph.KnowledgeGraphSwitch,
			GraphModelId:          knowledgeGraph.GraphModelId,
			GraphSchemaObjectName: knowledgeGraph.GraphSchemaObjectName,
			GraphSchemaFileName:   knowledgeGraph.GraphSchemaFileName,
		})
	})
}

// ReImportKnowledgeDoc 创建知识库文件
func ReImportKnowledgeDoc(ctx context.Context, doc *model.KnowledgeDoc, importTask *model.KnowledgeImportTask) error {
	knowledge, err := SelectKnowledgeById(ctx, doc.KnowledgeId, "", "")
	if err != nil {
		return err
	}
	var config = &model.SegmentConfig{}
	err = json.Unmarshal([]byte(importTask.SegmentConfig), config)
	if err != nil {
		log.Errorf("SegmentConfig process error %s", err.Error())
		return err
	}
	var analyzer = &model.DocAnalyzer{}
	err = json.Unmarshal([]byte(importTask.DocAnalyzer), analyzer)
	if err != nil {
		log.Errorf("DocAnalyzer process error %s", err.Error())

		return err
	}
	var preProcess = &model.DocPreProcess{}
	if len(importTask.DocPreProcess) > 0 {
		err = json.Unmarshal([]byte(importTask.DocPreProcess), preProcess)
		if err != nil {
			log.Errorf("DocPreprocess process error %s", err.Error())
			return err
		}
		preProcess.PreProcessList = normalizeList(preProcess.PreProcessList)
	}

	_, objectName, _ := service.SplitFilePath(doc.FilePath)
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		ragMetaList, err := buildMetaData(doc.DocId)
		if err != nil {
			log.Errorf("buildAndCreateMetaData error %s", err.Error())
		}
		//非初始话状态的不需要rag 导入，因为有可能直接失败了
		if doc.Status != model.DocInit {
			return nil
		}
		//构造知识库图谱
		knowledgeGraph := BuildKnowledgeGraph(string(knowledge.KnowledgeGraph))
		//2.rag文档导入
		return service.RagImportDoc(ctx, &service.RagImportDocParams{
			DocId:                 doc.DocId,
			KnowledgeName:         knowledge.RagName,
			CategoryId:            knowledge.KnowledgeId,
			UserId:                knowledge.UserId,
			Overlap:               config.Overlap,
			SegmentSize:           config.MaxSplitter,
			SegmentType:           service.RebuildSegmentType(config.SegmentType, config.SegmentMethod),
			SplitType:             service.RebuildSplitType(config.SegmentMethod),
			Separators:            config.Splitter,
			ParserChoices:         analyzer.AnalyzerList,
			ObjectName:            objectName,
			OriginalName:          doc.Name,
			IsEnhanced:            "false",
			OcrModelId:            importTask.OcrModelId,
			PreProcess:            preProcess.PreProcessList,
			RagMetaDataParams:     ragMetaList,
			RagChildChunkConfig:   buildSubRagChunkConfig(config),
			KnowledgeGraphSwitch:  knowledgeGraph.KnowledgeGraphSwitch,
			GraphModelId:          knowledgeGraph.GraphModelId,
			GraphSchemaObjectName: knowledgeGraph.GraphSchemaObjectName,
			GraphSchemaFileName:   knowledgeGraph.GraphSchemaFileName,
		})
	})
}

// CopyDocAndRemoveRag 复制文档并删除rag
func CopyDocAndRemoveRag(ctx context.Context, knowledge *model.KnowledgeBase, docId string, status int) (*model.KnowledgeDoc, error) {
	knowledgeDoc, err := GetDocDetail(ctx, "", "", docId)
	if err != nil {
		log.Errorf("GetDocDetail docId %s error %s", docId, err)
		return nil, err
	}

	copyFile, _, _, err := service.CopyFile(ctx, knowledgeDoc.FilePath, "", true)
	if err != nil {
		log.Errorf("CopyFile error %s", err)
		return nil, err
	}
	return knowledgeDoc, db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.更新文档文件
		err = UpdateDocInfo(tx, docId, -1, copyFile, "")
		if err != nil {
			log.Errorf("UpdateDocInfo %s", err)
			return err
		}
		//2.rag删除
		var fileName = service.RebuildFileName(knowledgeDoc.DocId, knowledgeDoc.FileType, knowledgeDoc.Name)
		err = service.RagDeleteDoc(ctx, &service.RagDeleteDocParams{
			UserId:        knowledge.UserId,
			KnowledgeBase: knowledge.RagName,
			FileName:      fileName,
		})
		if err != nil {
			return processError(status, err)
		}
		return nil
	})
}

// BatchUpdateDocStatus 批量更新文档状态
func BatchUpdateDocStatus(ctx context.Context, docIdList []string, status int) error {
	var updateParams = map[string]interface{}{
		"status": status,
	}
	err := db.GetHandle(ctx).Model(&model.KnowledgeDoc{}).Where("doc_id IN ?", docIdList).Updates(updateParams).Error
	if err != nil {
		return err
	}
	return nil
}

// UpdateDocInfo 更新文档信息,由于status 0 是有含义的，所以不想更新status 请传-1
func UpdateDocInfo(tx *gorm.DB, docId string, status int, fileUrl string, importTaskId string) error {
	var updateParams = map[string]interface{}{}
	if status >= 0 {
		updateParams["status"] = status
	}
	if len(fileUrl) > 0 {
		updateParams["file_path"] = fileUrl
		updateParams["file_path_md5"] = util.MD5(fileUrl)
	}
	if len(importTaskId) > 0 {
		updateParams["batch_id"] = importTaskId
	}
	if len(updateParams) == 0 {
		return nil
	}
	err := tx.Model(&model.KnowledgeDoc{}).Where("doc_id = ?", docId).Updates(updateParams).Error
	if err != nil {
		return err
	}
	return nil
}

// BuildKnowledgeGraph 知识图谱构造
func BuildKnowledgeGraph(knowledgeGraph string) *KnowledgeGraph {
	if len(knowledgeGraph) > 0 {
		graph := knowledgebase_service.KnowledgeGraph{}
		err := json.Unmarshal([]byte(knowledgeGraph), &graph)
		if err != nil {
			log.Errorf("knowledgeGraph process error %s", err.Error())
		}
		var graphSchemaObjectName, graphSchemaFileName string
		if len(graph.SchemaUrl) > 0 {
			_, graphSchemaObjectName, graphSchemaFileName = service.SplitFilePath(graph.SchemaUrl)
		}
		return &KnowledgeGraph{
			KnowledgeGraphSwitch:  graph.Switch,
			GraphModelId:          graph.LlmModelId,
			GraphSchemaObjectName: graphSchemaObjectName,
			GraphSchemaFileName:   graphSchemaFileName,
		}
	}
	return &KnowledgeGraph{
		KnowledgeGraphSwitch: false,
	}
}

// 子rag chunk的配置
func buildSubRagChunkConfig(config *model.SegmentConfig) *service.RagChunkConfig {
	if config.SegmentMethod == model.ParentSegmentMethod {
		return &service.RagChunkConfig{
			SegmentSize: config.SubMaxSplitter,
			Separators:  config.SubSplitter,
		}
	}
	return nil
}

func normalizeList(list []string) []string {
	for i, item := range list {
		switch item {
		case "deleteLinks":
			list[i] = PreprocessLink
		case "replaceSymbols":
			list[i] = PreprocessSymbol
		}
	}
	return list
}

func buildMetaData(docId string) ([]*service.RagMetaDataParams, error) {
	metaList, err := SelectDocMetaList(context.Background(), "", "", docId)
	if err != nil {
		return nil, err
	}
	if len(metaList) == 0 {
		return nil, nil
	}
	var ragMetaList []*service.RagMetaDataParams
	for _, meta := range metaList {
		// 构造rag参数
		ragValue, err := convertMetaValue(meta)
		if err != nil {
			return nil, err
		}
		ragMetaList = append(ragMetaList, &service.RagMetaDataParams{
			MetaId:    meta.MetaId,
			Key:       meta.Key,
			Value:     ragValue,
			ValueType: buildMetaValueType(meta.ValueType),
			Rule:      meta.Rule,
		})
	}
	return ragMetaList, nil
}

func buildAndCreateMetaData(tx *gorm.DB, importTask *model.KnowledgeImportTask, doc *model.KnowledgeDoc) ([]*service.RagMetaDataParams, error) {
	// 从importTask反序列化meta
	if len(importTask.MetaData) == 0 {
		return nil, nil
	}
	var importMetaData = model.DocImportMetaData{}
	err := json.Unmarshal([]byte(importTask.MetaData), &importMetaData)
	if err != nil {
		log.Errorf("Unmarshal fail %v", err)
		return nil, err
	}
	var metaList []*model.KnowledgeDocMeta
	var ragMetaList []*service.RagMetaDataParams
	for _, importMeta := range importMetaData.DocMetaDataList {
		// 构造meta数据库结构
		meta, err := buildKnowledgeDocMeta(doc, importTask, importMeta)
		if err != nil {
			return nil, err
		}
		if meta == nil {
			continue
		}
		metaList = append(metaList, meta)
		// 构造rag参数
		ragValue, err := convertMetaValue(meta)
		if err != nil {
			return nil, err
		}
		ragMetaList = append(ragMetaList, &service.RagMetaDataParams{
			MetaId:    meta.MetaId,
			Key:       meta.Key,
			Value:     ragValue,
			ValueType: buildMetaValueType(meta.ValueType),
			Rule:      meta.Rule,
		})
	}
	// 批量插入meta数据库
	err = createBatchKnowledgeDocMeta(tx, metaList)
	if err != nil {
		return nil, err
	}
	return ragMetaList, nil
}

func buildMetaValueType(valueType string) string {
	if fileDirType(valueType) {
		return MetaValueTypeString
	}
	return valueType
}
func convertMetaValue(meta *model.KnowledgeDocMeta) (interface{}, error) {
	if len(meta.ValueMain) == 0 {
		return nil, nil
	}
	// 根据类型转换value
	if meta.ValueType == MetaValueTypeNumber {
		ragValue, err := strconv.Atoi(meta.ValueMain)
		if err != nil {
			log.Errorf("convertMetaValue fail %v", err)
			return nil, err
		}
		return ragValue, nil
	}
	if meta.ValueType == MetaValueTypeTime {
		parseInt, err := strconv.ParseInt(meta.ValueMain, 10, 64)
		if err != nil {
			log.Errorf("convertMetaValue fail %v", err)
			return nil, err
		}
		return parseInt, nil
	}
	return meta.ValueMain, nil
}

// CreateKnowledgeUrlDoc 创建知识库url文件
func CreateKnowledgeUrlDoc(ctx context.Context, doc *model.KnowledgeDoc, importTask *model.KnowledgeImportTask) error {
	knowledge, err := SelectKnowledgeById(ctx, doc.KnowledgeId, doc.UserId, doc.OrgId)
	if err != nil {
		return err
	}
	var config = &model.SegmentConfig{}
	err = json.Unmarshal([]byte(importTask.SegmentConfig), config)
	if err != nil {
		log.Errorf("SegmentConfig process error %s", err.Error())
		return err
	}
	var analyzer = &model.DocAnalyzer{}
	err = json.Unmarshal([]byte(importTask.DocAnalyzer), analyzer)
	if err != nil {
		log.Errorf("DocAnalyzer process error %s", err.Error())
		return err
	}
	var preProcess = &model.DocPreProcess{}
	err = json.Unmarshal([]byte(importTask.DocPreProcess), preProcess)
	if err != nil {
		log.Errorf("DocPreprocess process error %s", err.Error())
		return err
	}
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.逻辑删除数据
		err = createKnowledgeDoc(tx, doc)
		if err != nil {
			return err
		}
		ragMetaList, err := buildAndCreateMetaData(tx, importTask, doc)
		if err != nil {
			log.Errorf("buildAndCreateMetaData error %s", err.Error())
		}
		//非初始话状态的不需要rag 导入，因为有可能直接失败了
		if doc.Status != model.DocInit {
			return nil
		}
		//2.rag url文档导入
		err = service.RagImportUrlDoc(ctx, &service.RagImportUrlDocParams{
			TaskId:            doc.DocId,
			FileName:          doc.Name,
			Url:               url.QueryEscape(doc.FilePath),
			UserId:            doc.UserId,
			Overlap:           config.Overlap,
			SegmentSize:       config.MaxSplitter,
			SegmentType:       service.RebuildSegmentType(config.SegmentType, config.SegmentMethod),
			SplitType:         service.RebuildSplitType(config.SegmentMethod),
			Separators:        config.Splitter,
			KnowledgeBaseName: knowledge.RagName,
			OcrModelId:        importTask.OcrModelId,
			PreProcess:        preProcess.PreProcessList,
			RagMetaDataParams: ragMetaList,
		})
		if err != nil {
			return err
		}
		//3.rag 文档开始导入操作
		var fileName = service.RebuildFileName(doc.DocId, doc.FileType, doc.Name)
		//构造知识库图谱
		knowledgeGraph := BuildKnowledgeGraph(string(knowledge.KnowledgeGraph))
		return service.RagImportDoc(ctx, &service.RagImportDocParams{
			DocId:                 doc.DocId,
			KnowledgeName:         knowledge.RagName,
			CategoryId:            knowledge.KnowledgeId,
			UserId:                doc.UserId,
			Overlap:               config.Overlap,
			SegmentSize:           config.MaxSplitter,
			SegmentType:           service.RebuildSegmentType(config.SegmentType, config.SegmentMethod),
			SplitType:             service.RebuildSplitType(config.SegmentMethod),
			Separators:            config.Splitter,
			ParserChoices:         analyzer.AnalyzerList,
			ObjectName:            fileName,
			OriginalName:          fileName,
			IsEnhanced:            "false",
			OcrModelId:            importTask.OcrModelId,
			PreProcess:            preProcess.PreProcessList,
			RagMetaDataParams:     ragMetaList,
			RagChildChunkConfig:   buildSubRagChunkConfig(config),
			KnowledgeGraphSwitch:  knowledgeGraph.KnowledgeGraphSwitch,
			GraphModelId:          knowledgeGraph.GraphModelId,
			GraphSchemaObjectName: knowledgeGraph.GraphSchemaObjectName,
			GraphSchemaFileName:   knowledgeGraph.GraphSchemaFileName,
		})
	})
}

// UpdateDocStatusDocId 更新文档状态
func UpdateDocStatusDocId(ctx context.Context, docId string, status int, metaList []*model.KnowledgeDocMeta) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//更新文档状态
		var updateParams, metaUpdate = buildUpdateParams(status)
		err := tx.Model(&model.KnowledgeDoc{}).Where("doc_id = ?", docId).Updates(updateParams).Error
		if err != nil {
			return err
		}
		//更新文档元数据
		if metaUpdate && len(metaList) > 0 {
			err := UpdateDocStatusMetaData(ctx, metaList)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// InitDocStatus 初始化文档状态
func InitDocStatus(ctx context.Context, userId, orgId string) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		err := stopDocProcess(tx)
		if err != nil {
			return err
		}
		_ = stopDocGraphProcess(tx)
		if err != nil {
			return err
		}
		_ = stopKnowledgeReport(tx)
		if err != nil {
			return err
		}
		return nil
	})
}

// DeleteDocByIdList 删除文档
func DeleteDocByIdList(ctx context.Context, idList []uint32, resultDocList []*model.KnowledgeDoc) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.逻辑删除数据
		err := logicDeleteDocByIdList(tx, idList)
		if err != nil {
			return err
		}
		err = DeleteKnowledgeFileInfo(tx, resultDocList[0].KnowledgeId, buildDocInfoList(resultDocList))
		//2.更新知识库条数
		if err != nil {
			return err
		}
		//3.异步执行删除数据
		return async_task.SubmitTask(ctx, async_task.DocDeleteTaskType, &async_task.DocDeleteParams{
			DocIdList: idList,
		})
	})
}

func buildDocInfoList(docList []*model.KnowledgeDoc) []*model.DocInfo {
	var retList []*model.DocInfo
	for _, doc := range docList {
		retList = append(retList, &model.DocInfo{
			DocSize: doc.FileSize,
		})
	}
	return retList
}

// ExecuteDeleteDocByIdList 执行删除
func ExecuteDeleteDocByIdList(tx *gorm.DB, idList []uint32) error {
	return tx.Unscoped().Where("id IN ?", idList).Delete(&model.KnowledgeDoc{}).Error
}

// logicDeleteDocByIdList 逻辑删除
func logicDeleteDocByIdList(tx *gorm.DB, idList []uint32) error {
	var updateParams = map[string]interface{}{
		"deleted": 1,
	}
	return tx.Model(&model.KnowledgeDoc{}).Where("id IN ?", idList).Updates(updateParams).Error
}

// createKnowledgeDoc 插入数据
func createKnowledgeDoc(tx *gorm.DB, knowledgeDoc *model.KnowledgeDoc) error {
	return tx.Model(&model.KnowledgeDoc{}).Create(knowledgeDoc).Error
}

func buildUpdateParams(status int) (map[string]interface{}, bool) {
	if model.InGraphStatus(status) { //图谱状态
		return map[string]interface{}{
			"graph_status": model.GraphStatus(status),
		}, false
	}
	//更新文档状态
	return map[string]interface{}{
		"status":    status,
		"error_msg": util.BuildDocErrMessage(status),
	}, true
}

func stopDocProcess(tx *gorm.DB) error {
	// 获取所有分析中状态的文档并更新状态
	updateDoc := map[string]interface{}{
		"status":    5,
		"error_msg": "know_doc_parsing_interrupted",
	}
	//会锁表风险极高
	return sqlopt.SQLOptions(sqlopt.WithStatusList(util.BuildAnalyzingStatus())).
		Apply(tx, &model.KnowledgeDoc{}).Updates(updateDoc).Error
}

func stopDocGraphProcess(tx *gorm.DB) error {
	// 获取所有分析中状态的文档并更新状态
	updateDoc := map[string]interface{}{
		"graph_status": model.GraphInterruptFail,
	}
	//会锁表风险极高
	return tx.Model(&model.KnowledgeDoc{}).Where("graph_status = ?", model.GraphProcessing).Updates(updateDoc).Error
}

func stopKnowledgeReport(tx *gorm.DB) error {
	// 获取所有分析中状态的文档并更新状态
	updateKnowledgeMap := map[string]interface{}{
		"report_status": model.ReportInterruptFail,
	}
	//会锁表风险极高
	return tx.Model(&model.KnowledgeBase{}).Where("report_status = ?", model.ReportProcessing).Updates(updateKnowledgeMap).Error
}

// SelectGraphStatus 查询知识图谱状态
func SelectGraphStatus(ctx context.Context, knowledgeId string, userId, orgId string) ([]*model.KnowledgeDoc, error) {
	var docList []*model.KnowledgeDoc
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithKnowledgeID(knowledgeId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeDoc{}).Select("doc_id", "graph_status").
		Find(&docList).Error
	if err != nil {
		log.Errorf("SelectDocByDocId userId %s err: %v", userId, err)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseAccessDenied)
	}
	if len(docList) == 0 {
		log.Errorf("SelectDocByDocId userId %s doc list empty", userId)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseAccessDenied)
	}
	return docList, nil
}

func processError(status int, err error) error {
	//判断：如果是status属于(51,52,53,54,55,56)，说明是RAG本身导致的解析异常，此时放行该错误”
	errStatus := []int{51, 52, 53, 54, 55, 56, 61, 62}
	for _, ragStatus := range errStatus {
		if status == ragStatus {
			return nil
		}
	}
	return err
}
