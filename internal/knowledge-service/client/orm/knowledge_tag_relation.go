package orm

import (
	"context"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	"gorm.io/gorm"
)

const (
	BindTag   int32 = 0 //绑定 [EN] binding
	UnbindTag int32 = 1 //解绑定 [EN] Unbind
)

// SelectKnowledgeTagRelationList 查询知识库标签关系列表 [EN] SelectKnowledgeTagRelationList Query the knowledge base tag relationship list
func SelectKnowledgeTagRelationList(ctx context.Context, userId, orgId string, knowledgeIdList []string) ([]*model.KnowledgeTagRelation, error) {
	var knowledgeTagRelationList []*model.KnowledgeTagRelation
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithKnowledgeIDList(knowledgeIdList)).
		Apply(db.GetHandle(ctx), &model.KnowledgeTagRelation{}).
		Find(&knowledgeTagRelationList).
		Error
	if err != nil {
		return nil, err
	}
	return knowledgeTagRelationList, nil
}

// SelectKnowledgeIdByTagId 根据tagId 查询知识库id [EN] SelectKnowledgeIdByTagId Query the knowledge base id based on tagId
func SelectKnowledgeIdByTagId(ctx context.Context, tagIdList []string) ([]string, error) {
	var knowledgeIdList []string
	err := db.GetHandle(ctx).Model(&model.KnowledgeTagRelation{}).
		Where("tag_id IN (?)", tagIdList).
		Distinct("knowledge_id").
		Pluck("knowledge_id", &knowledgeIdList).Error
	if err != nil {
		return nil, err
	}
	return knowledgeIdList, nil
}

// SelectKnowledgeCountByTagId 根据tagId 查询知识库id数量 [EN] SelectKnowledgeCountByTagId Query the number of knowledge base ids based on tagId
func SelectKnowledgeCountByTagId(ctx context.Context, tagId string) (int64, error) {
	var count int64
	err := db.GetHandle(ctx).Model(&model.KnowledgeTagRelation{}).
		Where("tag_id = ?", tagId).
		Count(&count).Error
	if err != nil {
		log.Errorf("SelectKnowledgeCountByTagId error %v", err)
		return 0, err
	}
	return count, nil
}

// BindKnowledgeTag 绑定知识库标签 [EN] BindKnowledgeTag binds knowledge base tags
func BindKnowledgeTag(ctx context.Context, dataList []*model.KnowledgeTagRelation, knowledgeId string) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.先删除所有 [EN] 1. Delete all first
		err := tx.Unscoped().Model(&model.KnowledgeTagRelation{}).Where("knowledge_id = ?", knowledgeId).
			Delete(&model.KnowledgeTagRelation{}).Error
		if err != nil {
			return err
		}
		//2.再绑定 [EN] 2.Rebind
		if len(dataList) > 0 {
			err = tx.Model(&model.KnowledgeTagRelation{}).CreateInBatches(dataList, len(dataList)).Error
			return err
		}
		return nil
	})
}
