package orm

import (
	"context"
	"strconv"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
	"gorm.io/gorm"
)

// SelectDocMetaList 获取文档元数据列表 [EN] SelectDocMetaList gets the document metadata list
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

// SelectMetaByDocIds 获取多个文档的元数据列表 [EN] SelectMetaByDocIds gets the metadata list of multiple documents
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

// SelectDocMetaListByKey 根据key,docId获取文档元数据值列表 [EN] SelectDocMetaListByKey Gets the list of document metadata values ​​based on key and docId
func SelectDocMetaListByKey(ctx context.Context, userId, orgId, docId, metaKey string) ([]*model.KnowledgeDocMeta, error) {
	var docMetaList []*model.KnowledgeDocMeta
	err := sqlopt.SQLOptions(sqlopt.WithDocID(docId), sqlopt.WithPermit(orgId, userId), sqlopt.WithKey(metaKey), sqlopt.WithNonEmptyValue()).
		Apply(db.GetHandle(ctx), &model.KnowledgeDocMeta{}).
		Order("create_at desc").
		Find(&docMetaList).
		Error
	if err != nil {
		return nil, err
	}
	return docMetaList, nil
}

// SelectMetaByKnowledgeId 获取单个知识库的元数据列表 [EN] SelectMetaByKnowledgeId Gets the metadata list of a single knowledge base
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

// UpdateDocStatusDocMeta 更新文档元数据 [EN] UpdateDocStatusDocMeta updates document metadata
func UpdateDocStatusDocMeta(ctx context.Context, docId string, addList []*model.KnowledgeDocMeta,
	updateList []*model.KnowledgeDocMeta, deleteDataIdList []string, ragDocMetaParams *service.RagDocMetaParams) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//todo 文档元数据应该不会特别多，所以先这么做，如果比较多，后续优化 [EN] There shouldn’t be too much metadata in todo documents, so do this first. If there are more metadata, then optimize later.
		if len(deleteDataIdList) > 0 {
			err := tx.Unscoped().Model(&model.KnowledgeDocMeta{}).Where("meta_id IN ?", deleteDataIdList).Delete(&model.KnowledgeDocMeta{}).Error
			if err != nil {
				return err
			}
		}
		if len(addList) > 0 {
			//插入数据 [EN] Insert data
			err := tx.Model(&model.KnowledgeDocMeta{}).CreateInBatches(addList, len(addList)).Error
			if err != nil {
				return err
			}
		}
		if len(updateList) > 0 {
			for _, meta := range updateList {
				//更新数据 [EN] Update data
				updateMap := map[string]interface{}{
					"value": meta.Value,
				}
				err := tx.Model(&model.KnowledgeDocMeta{}).Where("meta_id = ?", meta.MetaId).Updates(updateMap).Error
				if err != nil {
					return err
				}
			}
		}
		if ragDocMetaParams != nil {
			//调用rag [EN] callrag
			return service.RagDocMeta(ctx, ragDocMetaParams)
		}
		return nil
	})
}

func BatchUpdateDocMetaValue(ctx context.Context, addList, updateList []*model.KnowledgeDocMeta, deleteDataIdList []string, knowledge *model.KnowledgeBase, docList []*model.KnowledgeDoc, userId string, docIds []string) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		if len(addList) > 0 {
			//插入数据 [EN] Insert data
			err := tx.Model(&model.KnowledgeDocMeta{}).CreateInBatches(addList, len(addList)).Error
			if err != nil {
				return err
			}
		}
		if len(updateList) > 0 {
			for _, meta := range updateList {
				//更新数据 [EN] Update data
				updateMap := map[string]interface{}{
					"value": meta.Value,
				}
				err := tx.Model(&model.KnowledgeDocMeta{}).Where("meta_id = ?", meta.MetaId).Updates(updateMap).Error
				if err != nil {
					return err
				}
			}
		}
		if len(deleteDataIdList) > 0 {
			err := tx.Unscoped().Model(&model.KnowledgeDocMeta{}).Where("meta_id IN ?", deleteDataIdList).Delete(&model.KnowledgeDocMeta{}).Error
			if err != nil {
				return err
			}
		}
		ragParams, err := buildBatchUpdateMetaRAGParams(tx, knowledge, docList, userId, docIds)
		if err != nil {
			return err
		}
		err = service.BatchRagDocMeta(ctx, ragParams)
		if err != nil {
			return err
		}
		return nil
	})
}

func buildBatchUpdateMetaRAGParams(tx *gorm.DB, knowledge *model.KnowledgeBase, docList []*model.KnowledgeDoc, userId string, docIds []string) (*service.BatchRagDocMetaParams, error) {
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
	metaList, err := buildBatchMetaParamsList(docMetaList, docNameMap, docIds)
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

// UpdateBatchStatusDocMeta 批量更新文档tag [EN] UpdateBatchStatusDocMeta Batch update document tags
func UpdateBatchStatusDocMeta(ctx context.Context, knowledgeId string, docNameMap map[string]string, addList []*model.KnowledgeDocMeta,
	updateList []*model.KnowledgeDocMeta, ragDocMetaParams *service.BatchRagDocMetaParams) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		if len(addList) > 0 {
			//插入数据 [EN] Insert data
			err := tx.Model(&model.KnowledgeDocMeta{}).CreateInBatches(addList, len(addList)).Error
			if err != nil {
				return err
			}
		}
		if len(updateList) > 0 {
			for _, meta := range updateList {
				//更新数据 [EN] Update data
				updateMap := map[string]interface{}{
					"value": meta.Value,
				}
				err := tx.Model(&model.KnowledgeDocMeta{}).Where("meta_id = ?", meta.MetaId).Updates(updateMap).Error
				if err != nil {
					return err
				}
			}
		}
		//查询目前全量数据 [EN] Query the current full data
		var docMetaList []*model.KnowledgeDocMeta
		err := sqlopt.SQLOptions(sqlopt.WithKnowledgeID(knowledgeId)).
			Apply(tx, &model.KnowledgeDocMeta{}).
			Order("create_at desc").
			Find(&docMetaList).
			Error
		if err != nil {
			return err
		}
		list, err := buildMetaParamsList(docMetaList, docNameMap)
		if err != nil {
			return err
		}
		ragDocMetaParams.MetaList = list
		//调用rag [EN] callrag
		return service.BatchRagDocMeta(ctx, ragDocMetaParams)
	})
}

// buildBatchMetaParamsList 构建rag元数据参数 [EN] buildBatchMetaParamsList build rag metadata parameters
func buildBatchMetaParamsList(docMetaList []*model.KnowledgeDocMeta, docNameMap map[string]string, docIds []string) ([]*service.DocMetaInfo, error) {
	var docMetaMap = make(map[string][]*model.KnowledgeDocMeta)
	for _, meta := range docMetaList {
		metas := docMetaMap[meta.DocId]
		if len(metas) == 0 {
			metas = make([]*model.KnowledgeDocMeta, 0)
		}
		metas = append(metas, meta)
		docMetaMap[meta.DocId] = metas
	}
	var docTrueMap = make(map[string]bool)
	for _, docId := range docIds {
		docTrueMap[docId] = false
	}
	dataList := make([]*service.DocMetaInfo, 0)
	for docId, metas := range docMetaMap {
		var retList = make([]*service.MetaData, 0)
		for _, meta := range metas {
			valueData, err := buildValueData(meta.ValueType, meta.Value)
			if err != nil {
				log.Errorf("buildValueData error %s", err.Error())
				return nil, err
			}
			retList = append(retList, &service.MetaData{
				Key:       meta.Key,
				Value:     valueData,
				ValueType: meta.ValueType,
			})
		}
		dataList = append(dataList, &service.DocMetaInfo{
			FileName:     docNameMap[docId],
			MetaDataList: retList,
		})
		docTrueMap[docId] = true
	}
	for docId, isTrue := range docTrueMap {
		if !isTrue {
			dataList = append(dataList, &service.DocMetaInfo{
				FileName:     docNameMap[docId],
				MetaDataList: []*service.MetaData{},
			})
		}
	}
	return dataList, nil
}

// buildMetaParamsList 构建rag元数据参数 [EN] buildMetaParamsList build rag metadata parameters
func buildMetaParamsList(docMetaList []*model.KnowledgeDocMeta, docNameMap map[string]string) ([]*service.DocMetaInfo, error) {
	var docMetaMap = make(map[string][]*model.KnowledgeDocMeta)
	for _, meta := range docMetaList {
		metas := docMetaMap[meta.DocId]
		if len(metas) == 0 {
			metas = make([]*model.KnowledgeDocMeta, 0)
		}
		metas = append(metas, meta)
		docMetaMap[meta.DocId] = metas
	}
	dataList := make([]*service.DocMetaInfo, 0)
	for docId, metas := range docMetaMap {
		var retList = make([]*service.MetaData, 0)
		for _, meta := range metas {
			valueData, err := buildValueData(meta.ValueType, meta.Value)
			if err != nil {
				log.Errorf("buildValueData error %s", err.Error())
				return nil, err
			}
			retList = append(retList, &service.MetaData{
				Key:       meta.Key,
				Value:     valueData,
				ValueType: meta.ValueType,
			})
		}
		dataList = append(dataList, &service.DocMetaInfo{
			FileName:     docNameMap[docId],
			MetaDataList: retList,
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

// UpdateDocStatusMetaData 根据metaId更新元数据 [EN] UpdateDocStatusMetaData updates metadata based on metaId
func UpdateDocStatusMetaData(ctx context.Context, metaDataList []*model.KnowledgeDocMeta) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		// 遍历传入的元数据列表 [EN] Iterate over the passed in metadata list
		for _, meta := range metaDataList {
			err := tx.Model(&model.KnowledgeDocMeta{}).
				Where("meta_id = ?", meta.MetaId). // 匹配metaId [EN] Match metaId
				Update("value", meta.Value).Error  // 仅更新value [EN] Only update value
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteMetaDataByDocIdList 根据docIdList删除元数据 [EN] DeleteMetaDataByDocIdList Delete metadata based on docIdList
func DeleteMetaDataByDocIdList(tx *gorm.DB, knowledgeId string, docIdList []string) error {
	return tx.Unscoped().Model(&model.KnowledgeDocMeta{}).Where("doc_id IN ?", docIdList).Where("knowledge_id = ?", knowledgeId).Delete(&model.KnowledgeDocMeta{}).Error
}

// createBatchKnowledgeDocMeta 插入数据 [EN] createBatchKnowledgeDocMeta inserts data
func createBatchKnowledgeDocMeta(tx *gorm.DB, knowledgeDocMetaList []*model.KnowledgeDocMeta) error {
	err := tx.Model(&model.KnowledgeDocMeta{}).CreateInBatches(knowledgeDocMetaList, len(knowledgeDocMetaList)).Error
	if err != nil {
		return err
	}
	return nil
}

func BatchDeleteMeta(ctx context.Context, deleteList []string, knowledgeId string, ragDeleteParams *service.RagBatchDeleteMetaParams) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		// 批量删除元数据 [EN] Delete metadata in batches
		err := tx.Unscoped().Model(&model.KnowledgeDocMeta{}).Where("`key` IN ?", deleteList).Where("knowledge_id = ?", knowledgeId).Delete(&model.KnowledgeDocMeta{}).Error
		if err != nil {
			return err
		}
		// 调用rag [EN] callrag
		if ragDeleteParams != nil {
			return service.RagBatchDeleteMeta(ctx, ragDeleteParams)
		}
		return nil
	})
}

func BatchUpdateMetaKey(ctx context.Context, updateList []*service.RagMetaMapKeys, knowledgeId string, ragUpdateParams *service.RagBatchUpdateMetaKeyParams) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		// 批量更新元数据 [EN] Update metadata in batches
		for _, meta := range updateList {
			updateMap := map[string]interface{}{
				"key": meta.NewKey,
			}
			err := tx.Model(&model.KnowledgeDocMeta{}).Where("`key` = ?", meta.OldKey).Where("knowledge_id = ?", knowledgeId).Updates(updateMap).Error
			if err != nil {
				return err
			}
		}

		// 调用rag [EN] callrag
		if ragUpdateParams != nil {
			return service.RagBatchUpdateMeta(ctx, ragUpdateParams)
		}
		return nil
	})
}

func BatchAddMeta(ctx context.Context, addList []*model.KnowledgeDocMeta) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		// 批量插入元数据 [EN] Insert metadata in batches
		err := tx.Model(&model.KnowledgeDocMeta{}).CreateInBatches(addList, len(addList)).Error
		if err != nil {
			return err
		}
		return nil
	})
}
