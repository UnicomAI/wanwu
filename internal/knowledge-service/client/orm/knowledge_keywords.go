package orm

import (
	"context"
	"encoding/json"
	"errors"

	knowledgebase_keywords_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-keywords-service"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
	"gorm.io/gorm"
)

const (
	KeywordsAdd    = "add"    // Add keywords
	KeywordsDelete = "delete" // Delete keywords
	KeywordsUpdate = "update" // Update keywords
)

// GetKeywordsList queries the keyword list based on userId and orgId
func GetKeywordsList(ctx context.Context, req *knowledgebase_keywords_service.GetKnowledgeKeywordsListReq) ([]*model.KnowledgeKeywords, int64, error) {
	var keywordsList []*model.KnowledgeKeywords
	tx := sqlopt.SQLOptions(sqlopt.WithPermit(req.Identity.OrgId, req.Identity.UserId), sqlopt.WithNameOrAliasLike(req.Name)).
		Apply(db.GetHandle(ctx), &model.KnowledgeKeywords{})
	var total int64
	err := tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	limit := req.PageSize
	offset := req.PageSize * (req.PageNum - 1)
	err = tx.Order("updated_at desc").Limit(int(limit)).Offset(int(offset)).Find(&keywordsList).Error
	if err != nil {
		return nil, 0, err
	}
	return keywordsList, total, nil
}

// GetKeywordsById Query keywords based on keyword id
func GetKeywordsById(ctx context.Context, id uint32) (*model.KnowledgeKeywords, error) {
	var keywords model.KnowledgeKeywords
	err := sqlopt.SQLOptions(sqlopt.WithID(id)).
		Apply(db.GetHandle(ctx), &model.KnowledgeKeywords{}).First(&keywords).Error
	if err != nil {
		return nil, err
	}
	return &keywords, nil
}

// CheckRepeatedKeywords Checks whether the user has a keyword setting with the same name
func CheckRepeatedKeywords(ctx context.Context, req *knowledgebase_keywords_service.UpdateKnowledgeKeywordsReq) error {
	var keywordsList []*model.KnowledgeKeywords
	// Find a list of keywords with the same name
	err := sqlopt.SQLOptions(sqlopt.WithPermit(req.Detail.Identity.OrgId, req.Detail.Identity.UserId), sqlopt.WithName(req.Detail.Name), sqlopt.WithoutID(req.Id)).
		Apply(db.GetHandle(ctx), &model.KnowledgeKeywords{}).Find(&keywordsList).Error

	if err != nil {
		return err
	}
	// There are already keywords, check the knowledge base with the same name
	if len(keywordsList) > 0 {
		for _, kw := range keywordsList {
			dbKnowledgeIds, err := jsonToList(kw.KnowledgeBaseIds)
			if err != nil {
				return err
			}
			if util.HasIntersection(dbKnowledgeIds, req.Detail.KnowledgeBaseIds) {
				log.Errorf("有同名关键词")
				return gorm.ErrDuplicatedKey
			}
		}
	}
	return nil
}

// CreateKeywords Create keywords
func CreateKeywords(ctx context.Context, keywords *model.KnowledgeKeywords) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		// Create keywords
		err := db.GetHandle(ctx).Create(keywords).Error
		if err != nil {
			return err
		}
		keywordsParams, err := buildOperateKeywordsParams(ctx, keywords, KeywordsAdd)
		if err != nil {
			return err
		}
		// Sync RAG
		return service.RagOperateKeywords(ctx, keywordsParams)
	})
}

// DeleteKeywords deletes knowledge base keywords
func DeleteKeywords(ctx context.Context, id uint32) error {
	// Get keywords
	keywords, err := GetKeywordsById(ctx, id)
	if err != nil {
		return err
	}
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete keywords
		err = tx.Unscoped().Model(&model.KnowledgeKeywords{}).Where("id = ?", id).Delete(&model.KnowledgeKeywords{}).Error
		if err != nil {
			log.Errorf("DeleteKnowledgeKeywords err: %v", err)
			return err
		}
		// Sync RAG
		keywordsParams, err := buildOperateKeywordsParams(ctx, keywords, KeywordsDelete)
		if err != nil {
			return err
		}
		return service.RagOperateKeywords(ctx, keywordsParams)
	})
}

// UpdateKeywords Update knowledge base keywords
func UpdateKeywords(ctx context.Context, keywords *model.KnowledgeKeywords) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete keywords
		err := db.GetHandle(ctx).Model(&model.KnowledgeKeywords{}).Where("id = ?", keywords.Id).Updates(keywords).Debug().Error
		if err != nil {
			log.Errorf("UpdateKeywords err: %v", err)
			return err
		}
		// Sync RAG
		keywordsParams, err := buildOperateKeywordsParams(ctx, keywords, KeywordsUpdate)
		if err != nil {
			return err
		}
		return service.RagOperateKeywords(ctx, keywordsParams)
	})
}

func buildOperateKeywordsParams(ctx context.Context, keywords *model.KnowledgeKeywords, action string) (*service.RagOperateKeywordsParams, error) {
	// Deserialize id list
	knowledgeIds, err := jsonToList(keywords.KnowledgeBaseIds)
	if err != nil {
		return nil, err
	}
	if len(knowledgeIds) == 0 {
		return nil, errors.New("knowledgeIds length can not be zero")
	}

	list, _, err := SelectKnowledgeByIdList(ctx, knowledgeIds, keywords.UserId, keywords.OrgId)
	if err != nil || len(list) == 0 {
		return nil, errors.New("knowledgeId permission error")
	}
	knowledgeUserInfo := buildUserKnowledgeList(list)
	return &service.RagOperateKeywordsParams{
		Id:               keywords.Id,
		UserId:           keywords.UserId,
		Action:           action,
		Name:             keywords.Name,
		Alias:            []string{keywords.Alias},
		KnowledgeBaseIds: knowledgeIds,
		RagKnowledgeInfo: knowledgeUserInfo,
	}, nil
}

func jsonToList(knowledgeBaseIdStr string) ([]string, error) {
	var knowledgeIds []string
	err := json.Unmarshal([]byte(knowledgeBaseIdStr), &knowledgeIds)
	if err != nil {
		log.Errorf("反序列化错误: %v, 原始字符串: %s", err, knowledgeBaseIdStr)
		return nil, err
	}
	return knowledgeIds, nil
}

func buildUserKnowledgeList(knowledgeList []*model.KnowledgeBase) map[string][]*service.RagKnowledgeInfo {
	retMap := make(map[string][]*service.RagKnowledgeInfo)
	for _, knowledge := range knowledgeList {
		knowledgeInfos, exist := retMap[knowledge.UserId]
		if !exist {
			knowledgeInfos = make([]*service.RagKnowledgeInfo, 0)
		}
		knowledgeInfos = append(knowledgeInfos, &service.RagKnowledgeInfo{
			KnowledgeId:   knowledge.KnowledgeId,
			KnowledgeName: knowledge.RagName,
		})
		retMap[knowledge.UserId] = knowledgeInfos
	}
	return retMap
}
