package orm

import (
	"context"
	"strconv"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
	"gorm.io/gorm"
)

// SelectDocMetaList 获取文档元数据列表
func SelectDocMetaList(ctx context.Context, userId, orgId, docId string) ([]*model.KnowledgeDocMeta, error) {
	var docMetaList []*model.KnowledgeDocMeta
	err := sqlopt.SQLOptions(sqlopt.WithDocID(docId), sqlopt.WithPermit(orgId, userId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeDocMeta{}).
		Order("create_at desc").
		Find(&docMetaList).
		Error
	if err != nil {
		return nil, err
	}
	return docMetaList, nil
}

// SelectMetaByDocIds 获取多个文档的元数据列表
func SelectMetaByDocIds(ctx context.Context, userId, orgId string, docIds []string) ([]*model.KnowledgeDocMeta, error) {
	var docMetaList []*model.KnowledgeDocMeta
	err := sqlopt.SQLOptions(sqlopt.WithDocIDs(docIds), sqlopt.WithPermit(orgId, userId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeDocMeta{}).
		Order("create_at desc").
		Find(&docMetaList).
		Error
	if err != nil {
		return nil, err
	}
	return docMetaList, nil
}

// SelectMetaByKnowledgeId 获取单个知识库的元数据列表
func SelectMetaByKnowledgeId(ctx context.Context, userId, orgId string, knowledgeId string) ([]*model.KnowledgeDocMeta, error) {
	var docMetaList []*model.KnowledgeDocMeta
	err := sqlopt.SQLOptions(sqlopt.WithKnowledgeID(knowledgeId), sqlopt.WithPermit(orgId, userId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeDocMeta{}).
		Order("create_at desc").
		Find(&docMetaList).
		Error
	if err != nil {
		return nil, err
	}
	return docMetaList, nil
}

// SelectDocIdListByMetaValue 根据元数据值去筛选单个知识库的文档列表
//   - metaType=time：按 [startTime, endTime] 时间区间过滤 time 类型元数据
//   - metaType=string/number：按该类型 + metaValue 模糊匹配过滤
//   - metaType 为空(旧逻辑)：metaValue 模糊匹配，排除 time 类型
func SelectDocIdListByMetaValue(ctx context.Context, userId, orgId, knowledgeId, metaType, metaValue, startTime, endTime string) ([]string, error) {
	var docMetaList []*model.KnowledgeDocMeta
	docIdList := make([]string, 0)
	opts := sqlopt.SqlOptions{sqlopt.WithKnowledgeID(knowledgeId), sqlopt.WithPermit(orgId, userId)}
	switch metaType {
	case model.MetaTypeTime:
		if startTime == "" && endTime == "" {
			return docIdList, nil
		}
		opts = append(opts, sqlopt.WithValueType(model.MetaTypeTime), sqlopt.BetweenMetaValueTime(startTime, endTime))
	case model.MetaTypeString, model.MetaTypeNumber:
		if metaValue == "" {
			return docIdList, nil
		}
		opts = append(opts, sqlopt.WithValueType(metaType), sqlopt.LikeMetaValue(metaValue))
	default:
		// metaValue 模糊匹配，排除 time 类型
		if metaValue == "" {
			return docIdList, nil
		}
		opts = append(opts, sqlopt.LikeMetaValue(metaValue), sqlopt.WithNonType(model.MetaTypeTime))
	}
	err := opts.
		Apply(db.GetHandle(ctx), &model.KnowledgeDocMeta{}).
		Order("create_at desc").
		Find(&docMetaList).
		Error
	if err != nil {
		return nil, err
	}
	for _, docMeta := range docMetaList {
		if docMeta.DocId != "" {
			docIdList = append(docIdList, docMeta.DocId)
		}
	}
	return docIdList, nil
}

func BatchUpdateQAMetaValue(ctx context.Context, addList, updateList []*model.KnowledgeDocMeta, deleteDataIdList []string, knowledge *model.KnowledgeBase, userId string, qaPairIds []string) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		// 1.更新数据库
		err := modifyMetaValue(tx, addList, updateList, deleteDataIdList, knowledge.KnowledgeId)
		if err != nil {
			return err
		}
		// 2.构造rag参数
		ragParams, err := buildBatchUpdateQAMetaRAGParams(tx, knowledge, userId, qaPairIds)
		if err != nil {
			return err
		}
		// 3.发送rag请求
		err = service.BatchRagQAMeta(ctx, ragParams)
		if err != nil {
			return err
		}
		return nil
	})
}

func modifyMetaValue(tx *gorm.DB, addList []*model.KnowledgeDocMeta, updateList []*model.KnowledgeDocMeta, deleteDataIdList []string, knowledgeId string) error {
	hasChanges := false
	if len(addList) > 0 {
		//插入数据
		err := tx.Model(&model.KnowledgeDocMeta{}).CreateInBatches(addList, len(addList)).Error
		if err != nil {
			return err
		}
		hasChanges = true
	}
	if len(updateList) > 0 {
		for _, meta := range updateList {
			//更新数据
			updateMap := map[string]interface{}{
				"value_main": meta.ValueMain,
			}
			err := tx.Model(&model.KnowledgeDocMeta{}).Where("meta_id = ?", meta.MetaId).Updates(updateMap).Error
			if err != nil {
				return err
			}
		}
		hasChanges = true
	}
	if len(deleteDataIdList) > 0 {
		err := tx.Unscoped().Model(&model.KnowledgeDocMeta{}).Where("meta_id IN ?", deleteDataIdList).Delete(&model.KnowledgeDocMeta{}).Error
		if err != nil {
			return err
		}
		hasChanges = true
	}
	if hasChanges {
		err := tx.Model(&model.KnowledgeBase{}).Where("knowledge_id = ?", knowledgeId).
			Update("update_at", time.Now().UnixMilli()).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func BatchUpdateDocMetaValue(ctx context.Context, addList, updateList []*model.KnowledgeDocMeta, deleteDataIdList []string, knowledge *model.KnowledgeBase, docList []*model.KnowledgeDoc, userId string, docIds []string) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		// 1.更新数据库
		err := modifyMetaValue(tx, addList, updateList, deleteDataIdList, knowledge.KnowledgeId)
		if err != nil {
			return err
		}
		// 2.构造rag参数
		ragParams, err := buildBatchUpdateDocMetaRAGParams(tx, knowledge, docList, userId, docIds)
		if err != nil {
			return err
		}
		// 3.发送rag请求
		err = service.BatchRagDocMeta(ctx, ragParams)
		if err != nil {
			return err
		}
		return nil
	})
}

func buildBatchUpdateDocMetaRAGParams(tx *gorm.DB, knowledge *model.KnowledgeBase, docList []*model.KnowledgeDoc, userId string, docIds []string) (*service.BatchRagDocMetaParams, error) {
	docNameMap := make(map[string]string)
	for _, doc := range docList {
		docNameMap[doc.DocId] = service.RebuildFileName(doc.DocId, doc.FileType, doc.Name)
	}
	docMetaList := make([]*model.KnowledgeDocMeta, 0)
	err := tx.Where("doc_id in ?", docIds).Find(&docMetaList).Error
	if err != nil {
		log.Errorf("docId %v select meta fail %v", docIds, err)
		return nil, err
	}
	metaList, err := buildBatchDocMetaParamsList(docMetaList, docNameMap, docIds)
	if err != nil {
		log.Errorf("docId %v build meta params fail %v", docIds, err)
		return nil, err
	}
	ragParams := &service.BatchRagDocMetaParams{
		UserId:        userId,
		KnowledgeBase: knowledge.RagName,
		KnowledgeId:   knowledge.KnowledgeId,
		MetaList:      metaList,
	}
	return ragParams, nil
}

func buildBatchUpdateQAMetaRAGParams(tx *gorm.DB, knowledge *model.KnowledgeBase, userId string, qaPairIds []string) (*service.BatchRagQAMetaParams, error) {
	qaMetaList := make([]*model.KnowledgeDocMeta, 0)
	err := tx.Where("doc_id in ?", qaPairIds).Find(&qaMetaList).Error
	if err != nil {
		log.Errorf("docId %v select meta fail %v", qaPairIds, err)
		return nil, err
	}
	metaList, err := buildBatchQAMetaParamsList(qaMetaList, qaPairIds)
	if err != nil {
		log.Errorf("docId %v build meta params fail %v", qaPairIds, err)
		return nil, err
	}
	ragParams := &service.BatchRagQAMetaParams{
		UserId: userId,
		QABase: knowledge.RagName,
		QAId:   knowledge.KnowledgeId,
		Metas:  metaList,
	}
	return ragParams, nil
}

// buildBatchDocMetaParamsList 构建rag元数据参数
func buildBatchDocMetaParamsList(docMetaList []*model.KnowledgeDocMeta, docNameMap map[string]string, docIds []string) ([]*service.DocMetaInfo, error) {
	// 分组元数据
	docMetaMap := make(map[string][]*model.KnowledgeDocMeta)
	for _, meta := range docMetaList {
		docMetaMap[meta.DocId] = append(docMetaMap[meta.DocId], meta)
	}
	dataList := make([]*service.DocMetaInfo, 0, len(docIds))
	for _, docId := range docIds {
		metaDataList := make([]*service.MetaData, 0)
		for _, meta := range docMetaMap[docId] {
			valueData, err := buildValueData(meta.ValueType, meta.ValueMain)
			if err != nil {
				log.Errorf("buildValueData error %s", err.Error())
				continue
			}
			metaDataList = append(metaDataList, &service.MetaData{
				Key:       meta.Key,
				Value:     valueData,
				ValueType: meta.ValueType,
			})
		}
		dataList = append(dataList, &service.DocMetaInfo{
			FileName:     docNameMap[docId],
			MetaDataList: metaDataList,
		})
	}
	return dataList, nil
}

// buildBatchQAMetaParamsList 构建问答库rag元数据参数
func buildBatchQAMetaParamsList(qaMetaList []*model.KnowledgeDocMeta, qaPairIds []string) ([]*service.QAMetaInfo, error) {
	// 1.按QaPairId分组元数据
	var qaMetaMap = make(map[string][]*model.KnowledgeDocMeta)
	for _, meta := range qaMetaList {
		qaMetaMap[meta.DocId] = append(qaMetaMap[meta.DocId], meta)
	}
	// 2.直接遍历所有QaPairId构建结果
	dataList := make([]*service.QAMetaInfo, 0, len(qaPairIds))
	for _, qaPairId := range qaPairIds {
		metaDataList := make([]*service.QAMetaData, 0)
		for _, meta := range qaMetaMap[qaPairId] {
			valueData, err := buildValueData(meta.ValueType, meta.ValueMain)
			if err != nil {
				log.Errorf("buildValueData error %s", err.Error())
				continue
			}
			metaDataList = append(metaDataList, &service.QAMetaData{
				Key:       meta.Key,
				Value:     valueData,
				ValueType: meta.ValueType,
			})
		}
		dataList = append(dataList, &service.QAMetaInfo{
			QAPairId:     qaPairId,
			MetaDataList: metaDataList,
		})
	}
	return dataList, nil
}

func buildValueData(valueType string, value string) (interface{}, error) {
	switch valueType {
	case model.MetaTypeNumber:
	case model.MetaTypeTime:
		return strconv.ParseInt(value, 10, 64)
	}
	return value, nil
}

// UpdateDocStatusMetaData 根据metaId更新元数据
func UpdateDocStatusMetaData(ctx context.Context, metaDataList []*model.KnowledgeDocMeta) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		// 遍历传入的元数据列表
		for _, meta := range metaDataList {
			err := tx.Model(&model.KnowledgeDocMeta{}).
				Where("meta_id = ?", meta.MetaId).         // 匹配metaId
				Update("value_main", meta.ValueMain).Error // 仅更新value
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteMetaDataByDocIdList 根据docIdList删除元数据
func DeleteMetaDataByDocIdList(tx *gorm.DB, knowledgeId string, docIdList []string) error {
	return tx.Unscoped().Model(&model.KnowledgeDocMeta{}).Where("doc_id IN ?", docIdList).Where("knowledge_id = ?", knowledgeId).Delete(&model.KnowledgeDocMeta{}).Error
}

// createBatchKnowledgeDocMeta 插入数据
func createBatchKnowledgeDocMeta(tx *gorm.DB, knowledgeDocMetaList []*model.KnowledgeDocMeta) error {
	err := tx.Model(&model.KnowledgeDocMeta{}).CreateInBatches(knowledgeDocMetaList, len(knowledgeDocMetaList)).Error
	if err != nil {
		return err
	}
	return nil
}

func BatchDeleteMeta(ctx context.Context, deleteList []string, knowledge *model.KnowledgeBase) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		// 批量删除元数据
		err := tx.Unscoped().Model(&model.KnowledgeDocMeta{}).Where("`key` IN ?", deleteList).Where("knowledge_id = ?", knowledge.KnowledgeId).Delete(&model.KnowledgeDocMeta{}).Error
		if err != nil {
			return err
		}
		// 调用rag
		switch knowledge.Category {
		case model.CategoryQA:
			return service.RagQABatchDeleteMeta(ctx, &service.RagQABatchDeleteMetaParams{
				UserId: knowledge.UserId,
				QABase: knowledge.RagName,
				QAId:   knowledge.KnowledgeId,
				Keys:   deleteList,
			})
		default:
			return service.RagBatchDeleteMeta(ctx, &service.RagBatchDeleteMetaParams{
				UserId:        knowledge.UserId,
				KnowledgeBase: knowledge.RagName,
				KnowledgeId:   knowledge.KnowledgeId,
				Keys:          deleteList,
			})
		}
	})
}

func BatchUpdateMetaKey(ctx context.Context, updateList []*service.RagMetaMapKeys, knowledge *model.KnowledgeBase) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		// 批量更新元数据
		for _, meta := range updateList {
			updateMap := map[string]interface{}{
				"key": meta.NewKey,
			}
			err := tx.Model(&model.KnowledgeDocMeta{}).Where("`key` = ?", meta.OldKey).Where("knowledge_id = ?", knowledge.KnowledgeId).Updates(updateMap).Error
			if err != nil {
				return err
			}
		}
		// 调用rag
		switch knowledge.Category {
		case model.CategoryQA:
			return service.RagQABatchUpdateMeta(ctx, &service.RagQABatchUpdateMetaKeyParams{
				UserId:   knowledge.UserId,
				QABase:   knowledge.RagName,
				QAId:     knowledge.KnowledgeId,
				Mappings: updateList,
			})
		default:
			return service.RagBatchUpdateMeta(ctx, &service.RagBatchUpdateMetaKeyParams{
				UserId:        knowledge.UserId,
				KnowledgeBase: knowledge.RagName,
				KnowledgeId:   knowledge.KnowledgeId,
				Mappings:      updateList,
			})
		}
	})
}

func BatchAddMeta(ctx context.Context, addList []*model.KnowledgeDocMeta) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		// 批量插入元数据
		err := tx.Model(&model.KnowledgeDocMeta{}).CreateInBatches(addList, len(addList)).Error
		if err != nil {
			return err
		}
		return nil
	})
}

func SyncUpdateKnowledgeBase(ctx context.Context, knowledgeId string) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.KnowledgeBase{}).Where("knowledge_id = ?", knowledgeId).
			Update("update_at", time.Now().UnixMilli()).Error
		return err
	})
}
