package orm

import (
	"context"
	"sync"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/log"
	pkg_util "github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

type TagRelation struct {
	TagList      []*model.KnowledgeTag
	RelationList []*model.KnowledgeTagRelation
	TagErr       error
	RelationErr  error
}

type TagRelationDetail struct {
	TagId    string
	TagName  string
	Selected bool
}

// SelectKnowledgeTagListWithRelation 查询知识库标签列表 [EN] SelectKnowledgeTagListWithRelation queries the knowledge base tag list
func SelectKnowledgeTagListWithRelation(ctx context.Context, userId, orgId, name string, knowledgeIdList []string) *TagRelation {
	//因为tag表数据量并不会特别大，同时私有化，mysql 和 微服务同机部署，所以此方法并不会比sql 左联效率低 [EN] Because the amount of data in the tag table is not particularly large, it is also privatized, and mysql and microservices are deployed on the same machine, so this method is not less efficient than sql left join
	//不使用左联得原因，是因为需要进行name得模糊查询，sql 会比较复杂,如果此方法会影响性能再进行左联优化 [EN] The reason for not using a left join is because fuzzy query of name is required, and the sql will be more complicated. If this method will affect the performance, then optimize the left join.
	var tagRelation = TagRelation{}
	var ws sync.WaitGroup
	ws.Add(2)
	//查询tag返回数据 [EN] Query tag returns data
	go func() {
		defer pkg_util.PrintPanicStack()
		defer ws.Done()
		list, err := SelectKnowledgeTagList(ctx, userId, orgId, name)
		tagRelation.TagErr = err
		tagRelation.TagList = list
	}()
	//查询关系列表 [EN] Query relationship list
	go func() {
		defer pkg_util.PrintPanicStack()
		defer ws.Done()
		list, err := SelectKnowledgeTagRelationList(ctx, userId, orgId, knowledgeIdList)
		if err != nil {
			log.Errorf("SelectKnowledgeTagRelationList error %s", err)
		}
		tagRelation.RelationList = list
		tagRelation.RelationErr = err
	}()
	ws.Wait()
	return &tagRelation
}

// SelectKnowledgeTagList 查询知识库标签列表 [EN] SelectKnowledgeTagList Query the knowledge base tag list
func SelectKnowledgeTagList(ctx context.Context, userId, orgId, name string) ([]*model.KnowledgeTag, error) {
	var knowledgeTagList []*model.KnowledgeTag
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.LikeName(name)).
		Apply(db.GetHandle(ctx), &model.KnowledgeTag{}).
		Order("create_at desc").
		Find(&knowledgeTagList).
		Error
	if err != nil {
		return nil, err
	}
	return knowledgeTagList, nil
}

// SelectKnowledgeTagDetail 查询知识库标签详情 [EN] SelectKnowledgeTagDetail Query knowledge base tag details
func SelectKnowledgeTagDetail(ctx context.Context, userId, orgId, tagId string) (*model.KnowledgeTag, error) {
	var knowledgeTag model.KnowledgeTag
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithTagID(tagId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeTag{}).
		First(&knowledgeTag).Error
	if err != nil {
		return nil, err
	}
	return &knowledgeTag, nil
}

// CheckSameKnowledgeTagName 知识库标签是否存在同名 [EN] CheckSameKnowledgeTagName Whether the knowledge base tag has the same name
func CheckSameKnowledgeTagName(ctx context.Context, userId, orgId, name string) error {
	var count int64
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithName(name)).
		Apply(db.GetHandle(ctx), &model.KnowledgeTag{}).
		Count(&count).Error
	if err != nil {
		log.Errorf("KnowledgeTagNameExist userId %s name %s err: %v", userId, name, err)
		return util.ErrCode(errs.Code_KnowledgeTagDuplicateName)
	}
	if count > 0 {
		return util.ErrCode(errs.Code_KnowledgeTagDuplicateName)
	}
	return nil
}

// CreateKnowledgeTag 创建知识库标签 [EN] CreateKnowledgeTag creates a knowledge base tag
func CreateKnowledgeTag(ctx context.Context, knowledgeTag *model.KnowledgeTag) error {
	return db.GetHandle(ctx).Create(knowledgeTag).Error
}

// UpdateKnowledgeTag 更新知识库标签 [EN] UpdateKnowledgeTag updates the knowledge base tag
func UpdateKnowledgeTag(ctx context.Context, name string, id uint32) error {
	var updateParams = map[string]interface{}{
		"name": name,
	}
	return db.GetHandle(ctx).Model(&model.KnowledgeTag{}).Where("id = ?", id).Updates(updateParams).Error
}

// DeleteKnowledgeTag 删除知识库标签 [EN] DeleteKnowledgeTag Delete knowledge base tag
func DeleteKnowledgeTag(ctx context.Context, tagId string, id uint32) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Unscoped().Model(&model.KnowledgeTag{}).Where("id = ?", id).Delete(&model.KnowledgeTag{}).Error
		if err != nil {
			log.Errorf("DeleteKnowledgeTag err: %v", err)
			return err
		}
		err = tx.Unscoped().Model(&model.KnowledgeTagRelation{}).Where("tag_id = ?", tagId).Delete(&model.KnowledgeTagRelation{}).Error
		if err != nil {
			log.Errorf("DeleteKnowledgeTagRelation err: %v", err)
			return err
		}
		return nil
	})
}
