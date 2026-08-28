package orm

import (
	"context"
	"errors"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"gorm.io/gorm"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
)

// GetQAPairListByKnowledgeIdNoDeleteCheck 根据问答库id查询问答库文件列表
func GetQAPairListByKnowledgeIdNoDeleteCheck(ctx context.Context, userId, orgId string, knowledgeId string) ([]*model.KnowledgeQAPair, error) {
	var qaPairList []*model.KnowledgeQAPair
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithKnowledgeID(knowledgeId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeQAPair{}).Find(&qaPairList).Error
	if err != nil {
		return nil, err
	}
	return qaPairList, nil
}

// SelectQAPairByQAPairIdList 查询问答库问答对信息
func SelectQAPairByQAPairIdList(ctx context.Context, qaPairIdList []string, userId, orgId string) ([]*model.KnowledgeQAPair, error) {
	var qaPairList []*model.KnowledgeQAPair
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithQAPairIDs(qaPairIdList)).
		Apply(db.GetHandle(ctx), &model.KnowledgeQAPair{}).
		Find(&qaPairList).Error
	if err != nil {
		log.Errorf("SelectQAPairByQAPairId userId %s err: %v", userId, err)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseAccessDenied)
	}
	if len(qaPairList) == 0 {
		log.Errorf("SelectQAPairByQAPairId userId %s pair list empty", userId)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseAccessDenied)
	}
	return qaPairList, nil
}

// CheckKnowledgeQAPairQuestion 问答库问题重复校验
func CheckKnowledgeQAPairQuestion(ctx context.Context, userId string, knowledgeId string, questionMd5 string) error {
	var count int64
	err := sqlopt.SQLOptions(sqlopt.WithPermit("", userId),
		sqlopt.WithKnowledgeID(knowledgeId),
		sqlopt.WithQuestionMd5(questionMd5),
		sqlopt.WithDelete(0)).
		Apply(db.GetHandle(ctx), &model.KnowledgeQAPair{}).
		Count(&count).Error
	if err != nil {
		log.Errorf("CheckKnowledgeQAPairQuestion knowledgeId %s err: %v", knowledgeId, err)
		return errors.New("CheckKnowledgeQAPairQuestion error")
	}
	if count > 0 {
		return errors.New("CheckKnowledgeQAPairQuestion exist error")
	}
	return nil
}

// CreateKnowledgeQAPair 创建问答对
func CreateKnowledgeQAPair(ctx context.Context, qaPairs []*model.KnowledgeQAPair, importParams *service.RagAddQAPairParams) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.创建问答库问答对
		err := tx.Model(&model.KnowledgeQAPair{}).CreateInBatches(qaPairs, len(qaPairs)).Error
		if err != nil {
			return err
		}
		//2.通知rag创建问答对
		return service.RagBatchAddQAPairs(ctx, importParams)
	})
}

// CreateKnowledgeQAPairAndCount 创建问答对
func CreateKnowledgeQAPairAndCount(ctx context.Context, knowledgeId string, qaPairs []*model.KnowledgeQAPair, importParams *service.RagAddQAPairParams) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.创建问答库问答对
		err := tx.Model(&model.KnowledgeQAPair{}).CreateInBatches(qaPairs, len(qaPairs)).Error
		if err != nil {
			return err
		}
		//2.更新问答库问答对数量
		err = UpdateKnowledgeDocCount(tx, knowledgeId)
		if err != nil {
			return err
		}
		//3.通知rag创建问答对
		return service.RagBatchAddQAPairs(ctx, importParams)
	})
}

// UpdateKnowledgeQAPair 更新问答对
func UpdateKnowledgeQAPair(ctx context.Context, qaPair *model.KnowledgeQAPair, updateParams *service.RagUpdateQAPairParams) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.创建问答库问答对
		err := tx.Model(&model.KnowledgeQAPair{}).Where("qa_pair_id = ?", qaPair.QAPairId).
			Updates(map[string]interface{}{
				"question":     qaPair.Question,
				"answer":       qaPair.Answer,
				"question_md5": qaPair.QuestionMd5,
			}).Error
		if err != nil {
			return err
		}
		//2. 更新问答库记录
		err = tx.Model(&model.KnowledgeBase{}).Where("knowledge_id = ?", qaPair.KnowledgeId).
			Update("update_at", time.Now().UnixMilli()).Error
		if err != nil {
			return err
		}
		//3.通知rag更新问答对
		return service.RagUpdateQAPair(ctx, updateParams)
	})
}

// UpdateKnowledgeQAPairSwitch 启停问答对
func UpdateKnowledgeQAPairSwitch(ctx context.Context, qaPair *model.KnowledgeQAPair, updateParams *service.RagUpdateQAPairStatusParams) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.启停问答对
		err := tx.Model(&model.KnowledgeQAPair{}).Where("qa_pair_id = ?", qaPair.QAPairId).
			Updates(map[string]interface{}{
				"switch": qaPair.Switch,
			}).Error
		if err != nil {
			return err
		}
		//2. 更新问答库记录
		err = tx.Model(&model.KnowledgeBase{}).Where("knowledge_id = ?", qaPair.KnowledgeId).
			Update("update_at", time.Now().UnixMilli()).Error
		if err != nil {
			return err
		}
		//3.通知rag更新问答对
		return service.RagUpdateQAPairStatus(ctx, updateParams)
	})
}

// DeleteKnowledgeQAPair 删除问答对
func DeleteKnowledgeQAPair(ctx context.Context, knowledgeId string, qaPairIds []string, deleteParams *service.RagDeleteQAPairParams) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.删除问答库问答对
		err := tx.Model(&model.KnowledgeQAPair{}).Where("qa_pair_id in ?", qaPairIds).Delete(&model.KnowledgeQAPair{}).Error
		if err != nil {
			return err
		}
		//2.更新问答库记录
		err = UpdateKnowledgeDocCount(tx, knowledgeId)
		if err != nil {
			return err
		}
		//3.更新元数据记录
		err = DeleteMetaDataByDocIdList(tx, knowledgeId, qaPairIds)
		if err != nil {
			return err
		}
		//4.通知rag更新问答对
		return service.RagDeleteQAPair(ctx, deleteParams)
	})
}

// GetQAPairList 查询问答库问答对列表
func GetQAPairList(ctx context.Context, userId, orgId, knowledgeId, name string, status []int32, qaPairIds []string, pageSize int32, pageNum int32) ([]*model.KnowledgeQAPair, int64, error) {
	tx := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId),
		sqlopt.WithKnowledgeID(knowledgeId),
		sqlopt.WithQAStatusList(status),
		sqlopt.LikeQuestion(name),
		sqlopt.WithQAPairIDsNonEmpty(qaPairIds),
		sqlopt.WithDelete(0)).
		Apply(db.GetHandle(ctx), &model.KnowledgeQAPair{})
	var total int64
	err := tx.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	limit := pageSize
	offset := pageSize * (pageNum - 1)
	var qaPairList []*model.KnowledgeQAPair
	err = tx.Order("create_at asc, id asc").Limit(int(limit)).Offset(int(offset)).Find(&qaPairList).Error
	if err != nil {
		return nil, 0, err
	}
	return qaPairList, total, nil
}

// GetQAPairInfoById 查询问答库问答对详情
func GetQAPairInfoById(ctx context.Context, qaPairId, userId, orgId string) (*model.KnowledgeQAPair, error) {
	qaPair := &model.KnowledgeQAPair{}
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId),
		sqlopt.WithQAPairID(qaPairId),
		sqlopt.WithDelete(0)).
		Apply(db.GetHandle(ctx), &model.KnowledgeQAPair{}).Find(&qaPair).Error
	if err != nil {
		return nil, err
	}
	return qaPair, nil
}

// GetQAPairByQAPairIdNoDeleteCheck 查询问答对详情
func GetQAPairByQAPairIdNoDeleteCheck(ctx context.Context, userId, orgId string, qaPairId string) (*model.KnowledgeQAPair, error) {
	var qaPair *model.KnowledgeQAPair
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithQAPairID(qaPairId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeQAPair{}).First(&qaPair).Error
	if err != nil {
		return nil, err
	}
	return qaPair, nil
}

// ExecuteDeleteQAPairByQAPairIdList 执行删除
func ExecuteDeleteQAPairByQAPairIdList(tx *gorm.DB, qaPairIdList []string) error {
	return tx.Unscoped().Where("qa_pair_id IN ?", qaPairIdList).Delete(&model.KnowledgeQAPair{}).Error
}
