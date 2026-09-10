package orm

import (
	"context"
	"errors"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetParseTemplateList 查询某个文档类型下的解析模板列表，模板仅创建者可见
func GetParseTemplateList(ctx context.Context, userId, orgId, docType string) ([]*model.KnowledgeParseTemplate, error) {
	var templateList []*model.KnowledgeParseTemplate
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithDocType(docType)).
		Apply(db.GetHandle(ctx), &model.KnowledgeParseTemplate{}).
		Order("create_at desc").Find(&templateList).Error
	if err != nil {
		log.Errorf("GetParseTemplateList err: %v", err)
		return nil, err
	}
	return templateList, nil
}

// GetParseTemplate 查询解析模板详情
func GetParseTemplate(ctx context.Context, userId, orgId, templateId string) (*model.KnowledgeParseTemplate, error) {
	var template *model.KnowledgeParseTemplate
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithTemplateId(templateId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeParseTemplate{}).First(&template).Error
	if err != nil {
		log.Errorf("GetParseTemplate err: %v", err)
		return nil, err
	}
	return template, nil
}

// CheckRepeatedParseTemplate 同一用户同一文档类型下模板名不可重复，templateId 非空时排除自身
func CheckRepeatedParseTemplate(ctx context.Context, userId, orgId, docType, name, templateId string) error {
	var count int64
	tx := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithDocType(docType), sqlopt.WithName(name),
		sqlopt.WithoutTemplateId(templateId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeParseTemplate{})
	if err := tx.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}
	return nil
}

// GetBuiltInParseTemplate 取该文档类型的内置（默认）模板，没有则返回 nil
func GetBuiltInParseTemplate(ctx context.Context, userId, orgId, docType string) (*model.KnowledgeParseTemplate, error) {
	var template model.KnowledgeParseTemplate
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId),
		sqlopt.WithTemplateId(model.BuiltInTemplateId(userId, orgId, docType))).
		Apply(db.GetHandle(ctx), &model.KnowledgeParseTemplate{}).First(&template).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// CreateBuiltInParseTemplates 按文档类型补齐内置（默认）模板，返回新建的，调用方负责算出缺哪些
func CreateBuiltInParseTemplates(ctx context.Context, userId, orgId string, docTypes []string) ([]*model.KnowledgeParseTemplate, error) {
	created := make([]*model.KnowledgeParseTemplate, 0, len(docTypes))
	for _, docType := range docTypes {
		template, err := model.NewBuiltInParseTemplate(userId, orgId, docType)
		if err != nil {
			return nil, err
		}
		// 并发下另一方可能已插入同一 template_id，撞唯一索引直接跳过
		if err := db.GetHandle(ctx).Clauses(clause.OnConflict{DoNothing: true}).
			Create(template).Error; err != nil {
			return nil, err
		}
		created = append(created, template)
	}
	return created, nil
}

// CreateParseTemplate 创建解析模板
func CreateParseTemplate(ctx context.Context, template *model.KnowledgeParseTemplate) error {
	return db.GetHandle(ctx).Create(template).Error
}

// UpdateParseTemplate 更新解析模板，按创建者过滤避免越权
func UpdateParseTemplate(ctx context.Context, userId, orgId string, template *model.KnowledgeParseTemplate) error {
	result := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithTemplateId(template.TemplateId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeParseTemplate{}).
		Updates(map[string]interface{}{
			"name":            template.Name,
			"segment_config":  template.SegmentConfig,
			"doc_analyzer":    template.DocAnalyzer,
			"ocr_model_id":    template.OcrModelId,
			"doc_pre_process": template.DocPreProcess,
		})
	if result.Error != nil {
		log.Errorf("UpdateParseTemplate err: %v", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteParseTemplate 删除解析模板，同时解除知识库上对该模板的绑定
func DeleteParseTemplate(ctx context.Context, userId, orgId, templateId string) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		result := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithTemplateId(templateId)).
			Apply(tx, &model.KnowledgeParseTemplate{}).Delete(&model.KnowledgeParseTemplate{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return unbindParseTemplate(tx, templateId)
	})
}

// DeleteTemplateBindByKnowledgeId 根据知识库id 删除解析模板绑定
func DeleteTemplateBindByKnowledgeId(tx *gorm.DB, knowledgeId string) error {
	return sqlopt.SQLOptions(sqlopt.WithKnowledgeID(knowledgeId)).
		Apply(tx, &model.KnowledgeParseTemplateBind{}).
		Delete(&model.KnowledgeParseTemplateBind{}).Error
}

// DeleteTemplateBindExceptUser 删除该知识库上除指定用户外所有人的解析模板绑定，转让时清场用
func DeleteTemplateBindExceptUser(tx *gorm.DB, knowledgeId, userId, orgId string) error {
	return sqlopt.SQLOptions(sqlopt.WithKnowledgeID(knowledgeId)).
		Apply(tx, &model.KnowledgeParseTemplateBind{}).
		Where("NOT (user_id = ? AND org_id = ?)", userId, orgId).
		Delete(&model.KnowledgeParseTemplateBind{}).Error
}

// unbindParseTemplate 解除所有知识库对该模板的绑定
func unbindParseTemplate(tx *gorm.DB, templateId string) error {
	return sqlopt.SQLOptions(sqlopt.WithTemplateId(templateId)).
		Apply(tx, &model.KnowledgeParseTemplateBind{}).
		Delete(&model.KnowledgeParseTemplateBind{}).Error
}

// ReplaceKnowledgeParseTemplate 整体替换某人在某知识库上的解析模板绑定，不动别人的
func ReplaceKnowledgeParseTemplate(ctx context.Context, knowledgeId, userId, orgId string, binds []*model.KnowledgeParseTemplateBind) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		if err := sqlopt.SQLOptions(sqlopt.WithKnowledgeID(knowledgeId), sqlopt.WithPermit(orgId, userId)).
			Apply(tx, &model.KnowledgeParseTemplateBind{}).
			Delete(&model.KnowledgeParseTemplateBind{}).Error; err != nil {
			return err
		}
		if len(binds) == 0 {
			return nil
		}
		return tx.Create(&binds).Error
	})
}

// GetKnowledgeParseTemplateMap 批量查询某人在这批知识库上的解析模板绑定，按 knowledgeId 分组
func GetKnowledgeParseTemplateMap(ctx context.Context, userId, orgId string, knowledgeIds []string) (map[string][]*model.KnowledgeParseTemplateBind, error) {
	result := make(map[string][]*model.KnowledgeParseTemplateBind)
	if len(knowledgeIds) == 0 {
		return result, nil
	}
	var binds []*model.KnowledgeParseTemplateBind
	if err := sqlopt.SQLOptions(sqlopt.WithKnowledgeIDList(knowledgeIds), sqlopt.WithPermit(orgId, userId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeParseTemplateBind{}).Find(&binds).Error; err != nil {
		return nil, err
	}
	for _, bind := range binds {
		result[bind.KnowledgeId] = append(result[bind.KnowledgeId], bind)
	}
	return result, nil
}

// GetParseTemplateMap 按模板id批量查询当前用户的解析模板，按 templateId 分组
func GetParseTemplateMap(ctx context.Context, userId, orgId string, templateIds []string) (map[string]*model.KnowledgeParseTemplate, error) {
	result := make(map[string]*model.KnowledgeParseTemplate)
	if len(templateIds) == 0 {
		return result, nil
	}
	var templateList []*model.KnowledgeParseTemplate
	if err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithTemplateIds(templateIds)).
		Apply(db.GetHandle(ctx), &model.KnowledgeParseTemplate{}).Find(&templateList).Error; err != nil {
		log.Errorf("GetParseTemplateMap err: %v", err)
		return nil, err
	}
	for _, template := range templateList {
		result[template.TemplateId] = template
	}
	return result, nil
}
