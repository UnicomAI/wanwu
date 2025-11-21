package orm

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"

	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm/sqlopt"
	async_task "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/async-task"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/generator"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
	"gorm.io/gorm"
)

const (
	MetaValueTypeNumber = "number"
	MetaValueTypeTime   = "time"
	PreprocessSymbol    = "replace_symbols"
	PreprocessLink      = "delete_links"
)

type KnowledgeGraph struct {
	KnowledgeGraphSwitch  bool   `json:"knowledgeGraphSwitch"`
	GraphModelId          string `json:"graphModelId"`
	GraphSchemaObjectName string `json:"graphSchemaObjectName"`
	GraphSchemaFileName   string `json:"graphSchemaFileName"`
}

// GetDocList queries the knowledge base file list
func GetDocList(ctx context.Context, userId, orgId, knowledgeId, name, tag string,
	statusList []int, pageSize int32, pageNum int32) ([]*model.KnowledgeDoc, int64, error) {
	tx := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId),
		sqlopt.WithKnowledgeID(knowledgeId),
		sqlopt.LikeName(name),
		sqlopt.LikeTag(tag),
		sqlopt.WithStatusList(statusList),
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

// GetDocDetail Query knowledge base file details
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

// GetDocListByKnowledgeIdNoDeleteCheck Query the knowledge base file list based on the knowledge base id
func GetDocListByKnowledgeIdNoDeleteCheck(ctx context.Context, userId, orgId string, knowledgeId string) ([]*model.KnowledgeDoc, error) {
	var docList []*model.KnowledgeDoc
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithKnowledgeID(knowledgeId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeDoc{}).Find(&docList).Error
	if err != nil {
		return nil, err
	}
	return docList, nil
}

// GetDocListByKnowledgeId Query the knowledge base file list based on the knowledge base id
func GetDocListByKnowledgeId(ctx context.Context, userId, orgId string, knowledgeId string) ([]*model.KnowledgeDoc, error) {
	var docList []*model.KnowledgeDoc
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithKnowledgeID(knowledgeId), sqlopt.WithDelete(0)).
		Apply(db.GetHandle(ctx), &model.KnowledgeDoc{}).Find(&docList).Error
	if err != nil {
		return nil, err
	}
	return docList, nil
}

// GetDocListByIdListNoDeleteCheck Query the knowledge base file list
func GetDocListByIdListNoDeleteCheck(ctx context.Context, userId, orgId string, idList []uint32) ([]*model.KnowledgeDoc, error) {
	var docList []*model.KnowledgeDoc
	err := sqlopt.SQLOptions(sqlopt.WithPermit("", userId), sqlopt.WithIDs(idList)).
		Apply(db.GetHandle(ctx), &model.KnowledgeDoc{}).Find(&docList).Error
	if err != nil {
		return nil, err
	}
	return docList, nil
}

// CheckKnowledgeDocSameName Knowledge base document same name verification
func CheckKnowledgeDocSameName(ctx context.Context, userId string, knowledgeId string, docName string, docUrl string) error {
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

// SelectDocByDocIdList queries knowledge base document information
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

func buildKnowledgeDocMeta(doc *model.KnowledgeDoc, importTask *model.KnowledgeImportTask, meta *model.KnowledgeDocMeta) (*model.KnowledgeDocMeta, error) {
	return &model.KnowledgeDocMeta{
		MetaId:    generator.GetGenerator().NewID(),
		DocId:     doc.DocId,
		Key:       meta.Key,
		Value:     meta.Value,
		ValueType: meta.ValueType,
		Rule:      meta.Rule,
		UserId:    importTask.UserId,
		OrgId:     importTask.OrgId,
	}, nil
}

// CreateKnowledgeDoc creates a knowledge base document
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
		//1. Insert data
		err = createKnowledgeDoc(tx, doc)
		if err != nil {
			return err
		}
		ragMetaList, err := buildAndCreateMetaData(tx, importTask, doc)
		if err != nil {
			log.Errorf("buildAndCreateMetaData error %s", err.Error())
		}
		//There is no need to import rag in non-initial state, because it may fail directly.
		if doc.Status != model.DocInit {
			return nil
		}
		//Constructing a knowledge base graph
		knowledgeGraph := BuildKnowledgeGraph(knowledge.KnowledgeGraph)
		//2.rag document import
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

// BuildKnowledgeGraph knowledge graph construction
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

// Configuration of sub-rag chunk
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

func buildAndCreateMetaData(tx *gorm.DB, importTask *model.KnowledgeImportTask, doc *model.KnowledgeDoc) ([]*service.RagMetaDataParams, error) {
	// Deserialize meta from importTask
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
		// Construct meta database structure
		meta, err := buildKnowledgeDocMeta(doc, importTask, importMeta)
		if err != nil {
			return nil, err
		}
		metaList = append(metaList, meta)
		// Construct rag parameters
		ragValue, err := convertMetaValue(meta)
		if err != nil {
			return nil, err
		}
		ragMetaList = append(ragMetaList, &service.RagMetaDataParams{
			MetaId:    meta.MetaId,
			Key:       meta.Key,
			Value:     ragValue,
			ValueType: meta.ValueType,
			Rule:      meta.Rule,
		})
	}
	// Batch insert into meta database
	err = createBatchKnowledgeDocMeta(tx, metaList)
	if err != nil {
		return nil, err
	}
	return ragMetaList, nil
}

func convertMetaValue(meta *model.KnowledgeDocMeta) (interface{}, error) {
	if len(meta.Value) == 0 {
		return nil, nil
	}
	// Convert value according to type
	if meta.ValueType == MetaValueTypeNumber {
		ragValue, err := strconv.Atoi(meta.Value)
		if err != nil {
			log.Errorf("convertMetaValue fail %v", err)
			return nil, err
		}
		return ragValue, nil
	}
	if meta.ValueType == MetaValueTypeTime {
		parseInt, err := strconv.ParseInt(meta.Value, 10, 64)
		if err != nil {
			log.Errorf("convertMetaValue fail %v", err)
			return nil, err
		}
		return parseInt, nil
	}
	return meta.Value, nil
}

// CreateKnowledgeUrlDoc creates a knowledge base url file
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
		//1. Logically delete data
		err = createKnowledgeDoc(tx, doc)
		if err != nil {
			return err
		}
		ragMetaList, err := buildAndCreateMetaData(tx, importTask, doc)
		if err != nil {
			log.Errorf("buildAndCreateMetaData error %s", err.Error())
		}
		//There is no need to import rag in non-initial state, because it may fail directly.
		if doc.Status != model.DocInit {
			return nil
		}
		//2.rag url document import
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
		//3.rag document starts importing operation
		var fileName = service.RebuildFileName(doc.DocId, doc.FileType, doc.Name)
		//Constructing a knowledge base graph
		knowledgeGraph := BuildKnowledgeGraph(knowledge.KnowledgeGraph)
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

// UpdateDocStatusDocId Update document status
func UpdateDocStatusDocId(ctx context.Context, docId string, status int, metaList []*model.KnowledgeDocMeta) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//Update document status
		var updateParams, metaUpdate = buildUpdateParams(status)
		err := tx.Model(&model.KnowledgeDoc{}).Where("doc_id = ?", docId).Updates(updateParams).Error
		if err != nil {
			return err
		}
		//Update document metadata
		if metaUpdate && len(metaList) > 0 {
			err := UpdateDocStatusMetaData(ctx, metaList)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// InitDocStatus initializes the document status
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

// DeleteDocByIdList deletes documents
func DeleteDocByIdList(ctx context.Context, idList []uint32, resultDocList []*model.KnowledgeDoc) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1. Logically delete data
		err := logicDeleteDocByIdList(tx, idList)
		if err != nil {
			return err
		}
		err = DeleteKnowledgeFileInfo(tx, resultDocList[0].KnowledgeId, buildDocInfoList(resultDocList))
		//2. Update the number of knowledge base items
		if err != nil {
			return err
		}
		//3. Execute deletion of data asynchronously
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

// ExecuteDeleteDocByIdList executes deletion
func ExecuteDeleteDocByIdList(tx *gorm.DB, idList []uint32) error {
	return tx.Unscoped().Where("id IN ?", idList).Delete(&model.KnowledgeDoc{}).Error
}

// logicDeleteDocByIdList logical deletion
func logicDeleteDocByIdList(tx *gorm.DB, idList []uint32) error {
	var updateParams = map[string]interface{}{
		"deleted": 1,
	}
	return tx.Model(&model.KnowledgeDoc{}).Where("id IN ?", idList).Updates(updateParams).Error
}

// createKnowledgeDoc inserts data
func createKnowledgeDoc(tx *gorm.DB, knowledgeDoc *model.KnowledgeDoc) error {
	return tx.Model(&model.KnowledgeDoc{}).Create(knowledgeDoc).Error
}

func buildUpdateParams(status int) (map[string]interface{}, bool) {
	if model.InGraphStatus(status) { //Map status
		return map[string]interface{}{
			"graph_status": model.GraphStatus(status),
		}, false
	}
	//Update document status
	return map[string]interface{}{
		"status":    status,
		"error_msg": util.BuildDocErrMessage(status),
	}, true
}

func stopDocProcess(tx *gorm.DB) error {
	// Get all documents with status under analysis and update the status
	updateDoc := map[string]interface{}{
		"status":    5,
		"error_msg": "know_doc_parsing_interrupted",
	}
	//The risk of table locking is extremely high
	return sqlopt.SQLOptions(sqlopt.WithStatusList(util.BuildAnalyzingStatus())).
		Apply(tx, &model.KnowledgeDoc{}).Updates(updateDoc).Error
}

func stopDocGraphProcess(tx *gorm.DB) error {
	// Get all documents with status under analysis and update the status
	updateDoc := map[string]interface{}{
		"graph_status": model.GraphInterruptFail,
	}
	//The risk of table locking is extremely high
	return tx.Model(&model.KnowledgeDoc{}).Where("graph_status = ?", model.GraphProcessing).Updates(updateDoc).Error
}

func stopKnowledgeReport(tx *gorm.DB) error {
	// Get all documents with status under analysis and update the status
	updateKnowledgeMap := map[string]interface{}{
		"report_status": model.ReportInterruptFail,
	}
	//The risk of table locking is extremely high
	return tx.Model(&model.KnowledgeBase{}).Where("report_status = ?", model.ReportProcessing).Updates(updateKnowledgeMap).Error
}

// SelectGraphStatus Query the status of the knowledge graph
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
