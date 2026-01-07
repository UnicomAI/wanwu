package orm

import (
	"context"
	"fmt"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm/sqlopt"
	async_task "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/async-task"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
	wanwu_util "github.com/UnicomAI/wanwu/pkg/util"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

// SelectKnowledgeList 查询知识库列表
func SelectKnowledgeList(ctx context.Context, userId, orgId, name string, category int, tagIdList []string) ([]*model.KnowledgeBase, map[string]int, error) {
	var knowledgeIdList []string
	var err error
	if len(tagIdList) > 0 {
		knowledgeIdList, err = SelectKnowledgeIdByTagId(ctx, tagIdList)
		if err != nil {
			return nil, nil, err
		}
		if len(knowledgeIdList) == 0 {
			return make([]*model.KnowledgeBase, 0), nil, nil
		}
	}
	//查询有权限的知识库列表，获取有权限的知识库id，目前是getALL，没有通过连表实现
	permissionKnowledgeList, err := SelectKnowledgeIdByPermission(ctx, userId, orgId, model.PermissionTypeView)
	if err != nil {
		return nil, nil, err
	}
	if len(permissionKnowledgeList) == 0 {
		return make([]*model.KnowledgeBase, 0), nil, nil
	}
	knowledgeIdList = intersectionKnowledgeIdList(knowledgeIdList, buildPermissionKnowledgeIdList(permissionKnowledgeList))
	if len(knowledgeIdList) == 0 {
		return make([]*model.KnowledgeBase, 0), nil, nil
	}
	var knowledgeList []*model.KnowledgeBase
	err = sqlopt.SQLOptions(sqlopt.WithKnowledgeIDList(knowledgeIdList), sqlopt.LikeName(name), sqlopt.WithDelete(0), sqlopt.WithCategory(category)).
		Apply(db.GetHandle(ctx), &model.KnowledgeBase{}).
		Order("update_at desc").
		Find(&knowledgeList).
		Error
	if err != nil {
		return nil, nil, err
	}
	return knowledgeList, buildPermissionKnowledgeIdMap(permissionKnowledgeList), nil
}

// SelectKnowledgeById 查询知识库信息,todo
func SelectKnowledgeById(ctx context.Context, knowledgeId, userId, orgId string) (*model.KnowledgeBase, error) {
	var knowledge model.KnowledgeBase
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithKnowledgeID(knowledgeId), sqlopt.WithDelete(0)).
		Apply(db.GetHandle(ctx), &model.KnowledgeBase{}).
		First(&knowledge).Error
	if err != nil {
		log.Errorf("SelectKnowledgeById userId %s err: %v", userId, err)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseAccessDenied)
	}
	return &knowledge, nil
}

// SelectKnowledgeByIdList 查询知识库信息
func SelectKnowledgeByIdList(ctx context.Context, knowledgeIdList []string, userId, orgId string) ([]*model.KnowledgeBase, map[string]int, error) {
	//查询有权限的知识库列表，获取有权限的知识库id，目前是getALL，没有通过连表实现
	permissionKnowledgeList, err := SelectKnowledgeIdByPermission(ctx, userId, orgId, model.PermissionTypeView)
	if err != nil {
		return nil, nil, err
	}
	if len(permissionKnowledgeList) == 0 {
		return make([]*model.KnowledgeBase, 0), nil, nil
	}
	knowledgeIdList = intersectionKnowledgeIdList(knowledgeIdList, buildPermissionKnowledgeIdList(permissionKnowledgeList))
	if len(knowledgeIdList) == 0 {
		return make([]*model.KnowledgeBase, 0), nil, nil
	}
	var knowledgeList []*model.KnowledgeBase
	err = sqlopt.SQLOptions(sqlopt.WithKnowledgeIDList(knowledgeIdList), sqlopt.WithDelete(0)).
		Apply(db.GetHandle(ctx), &model.KnowledgeBase{}).
		Find(&knowledgeList).Error
	if err != nil {
		log.Errorf("SelectKnowledgeByIdList userId %s err: %v", userId, err)
		return nil, nil, util.ErrCode(errs.Code_KnowledgeBaseAccessDenied)
	}
	return knowledgeList, buildPermissionKnowledgeIdMap(permissionKnowledgeList), nil
}

// SelectKnowledgeByName 查询知识库信息
func SelectKnowledgeByName(ctx context.Context, knowledgeName, userId, orgId string) (*model.KnowledgeBase, error) {
	var knowledge model.KnowledgeBase
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithName(knowledgeName), sqlopt.WithDelete(0)).
		Apply(db.GetHandle(ctx), &model.KnowledgeBase{}).
		First(&knowledge).Error
	if err != nil {
		log.Errorf("SelectKnowledgeByName userId %s err: %v", userId, err)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseAccessDenied)
	}
	return &knowledge, nil
}

// SelectKnowledgeByIdNoDeleteCheck 查询知识库信息
func SelectKnowledgeByIdNoDeleteCheck(ctx context.Context, knowledgeId, userId, orgId string) (*model.KnowledgeBase, error) {
	var knowledge model.KnowledgeBase
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithKnowledgeID(knowledgeId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeBase{}).
		First(&knowledge).Error
	if err != nil {
		log.Errorf("SelectKnowledgeById userId %s err: %v", userId, err)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseAccessDenied)
	}
	return &knowledge, nil
}

// CheckSameKnowledgeName 知识库名称是否存在同名
func CheckSameKnowledgeName(ctx context.Context, userId, orgId, name, knowledgeId string, category int) error {
	//var count int64
	//err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithName(name), sqlopt.WithoutKnowledgeID(knowledgeId), sqlopt.WithDelete(0)).
	//	Apply(db.GetHandle(ctx), &model.KnowledgeBase{}).
	//	Count(&count).Error
	//if err != nil {
	//	log.Errorf("KnowledgeNameExist userId %s name %s err: %v", userId, name, err)
	//	return util.ErrCode(errs.Code_KnowledgeBaseDuplicateName)
	//}
	//if count > 0 {
	//	return util.ErrCode(errs.Code_KnowledgeBaseDuplicateName)
	//}
	//return nil

	list, _, err := SelectKnowledgeList(ctx, userId, orgId, name, category, nil)
	if err != nil {
		log.Errorf(fmt.Sprintf("获取知识库列表失败(%v)  参数(%v)", err, name))
		return util.ErrCode(errs.Code_KnowledgeBaseDuplicateName)
	}
	var resultList []*model.KnowledgeBase
	for _, base := range list {
		if base.Name == name {
			resultList = append(resultList, base)
		}
	}
	if len(resultList) > 1 {
		return util.ErrCode(errs.Code_KnowledgeBaseDuplicateName)
	}

	if len(resultList) == 1 && resultList[0].KnowledgeId != knowledgeId {
		return util.ErrCode(errs.Code_KnowledgeBaseDuplicateName)
	}

	return nil
}

// CreateKnowledge 创建知识库
func CreateKnowledge(ctx context.Context, knowledge *model.KnowledgeBase, embeddingModelId string, category int) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.插入数据
		err := createKnowledge(tx, knowledge)
		if err != nil {
			return err
		}
		//2.插入权限信息
		err = CreateKnowledgeIdPermission(tx, buildKnowledgePermission(knowledge))
		if err != nil {
			return err
		}
		//3.通知rag创建知识库
		switch category {
		case model.CategoryQA:
			return service.RagQACreate(ctx, &service.RagQACreateParams{
				UserId:           knowledge.UserId,
				QABase:           knowledge.RagName,
				QAId:             knowledge.KnowledgeId,
				EmbeddingModelId: embeddingModelId,
			})
		default:
			return service.RagKnowledgeCreate(ctx, &service.RagCreateParams{
				UserId:               knowledge.UserId,
				Name:                 knowledge.RagName,
				KnowledgeBaseId:      knowledge.KnowledgeId,
				EmbeddingModelId:     embeddingModelId,
				EnableKnowledgeGraph: knowledge.KnowledgeGraphSwitch > 0,
			})
		}
	})
}

// UpdateKnowledge 更新知识库
func UpdateKnowledge(ctx context.Context, name, description string, knowledgeBase *model.KnowledgeBase) error {
	//return updateKnowledge(db.GetHandle(ctx), knowledgeBase.Id, name, description)
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//已经区分为知识库展示名称和rag知识库名称，不需要再通知rag修改名称
		if knowledgeBase.Name != knowledgeBase.RagName {
			return updateKnowledge(tx, knowledgeBase.Id, name, description)
		}
		//2.更新数据
		ragName := wanwu_util.NewID()
		err := updateKnowledgeWithRagName(tx, knowledgeBase.Id, name, ragName, description)
		if err != nil {
			return err
		}

		//2.通知rag更新知识库,只有老的需要更新
		return service.RagKnowledgeUpdate(ctx, &service.RagUpdateParams{
			UserId:          knowledgeBase.UserId,
			KnowledgeBaseId: knowledgeBase.KnowledgeId,
			OldKbName:       knowledgeBase.RagName,
			NewKbName:       ragName,
		})
	})
}

// UpdateKnowledgeShareCount 更新知识库分享数量
func UpdateKnowledgeShareCount(tx *gorm.DB, knowledgeId string, count int64) error {
	var updateParams = map[string]interface{}{
		"share_count": count,
	}
	return tx.Model(&model.KnowledgeBase{}).Where("knowledge_id=?", knowledgeId).Updates(updateParams).Error
}

// UpdateKnowledgeGraph 更新知识库图谱
func UpdateKnowledgeGraph(tx *gorm.DB, knowledgeId string, knowledgeGraph string) error {
	var updateParams = map[string]interface{}{
		"knowledge_graph": knowledgeGraph,
	}
	return tx.Model(&model.KnowledgeBase{}).Where("knowledge_id=?", knowledgeId).Updates(updateParams).Error
}

// UpdateKnowledgeReportStatus 更新社区报告状态
func UpdateKnowledgeReportStatus(ctx context.Context, knowledgeId string, reportStatus int) error {
	var updateParams = map[string]interface{}{
		"report_status": model.ReportStatus(reportStatus),
	}
	return db.GetHandle(ctx).Model(&model.KnowledgeBase{}).Where("knowledge_id=?", knowledgeId).Updates(updateParams).Error
}

// DeleteKnowledge 删除知识库
func DeleteKnowledge(ctx context.Context, knowledgeBase *model.KnowledgeBase) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.逻辑删除数据
		err := logicDeleteKnowledge(tx, knowledgeBase)
		if err != nil {
			return err
		}
		//2.通知rag更新知识库
		switch knowledgeBase.Category {
		case model.CategoryQA:
			return async_task.SubmitTask(ctx, async_task.QADeleteTaskType, &async_task.KnowledgeDeleteParams{
				KnowledgeId: knowledgeBase.KnowledgeId,
			})
		default:
			return async_task.SubmitTask(ctx, async_task.KnowledgeDeleteTaskType, &async_task.KnowledgeDeleteParams{
				KnowledgeId: knowledgeBase.KnowledgeId,
			})
		}

	})
}

// ExecuteDeleteKnowledge 删除知识库
func ExecuteDeleteKnowledge(tx *gorm.DB, id uint32) error {
	return tx.Unscoped().Model(&model.KnowledgeBase{}).Where("id = ?", id).Delete(&model.KnowledgeBase{}).Error
}

// ExecuteDeleteKnowledgeMeta 删除知识库元数据
func ExecuteDeleteKnowledgeMeta(tx *gorm.DB, knowledgeId string) error {
	return tx.Unscoped().Model(&model.KnowledgeDocMeta{}).Where("knowledge_id = ?", knowledgeId).Delete(&model.KnowledgeDocMeta{}).Error
}

// UpdateKnowledgeFileInfo 更新知识库文档信息
func UpdateKnowledgeFileInfo(tx *gorm.DB, knowledgeId string, resultList []*model.DocInfo) error {
	var docSize int64
	for _, result := range resultList {
		docSize += result.DocSize
	}
	return tx.Model(&model.KnowledgeBase{}).Where("knowledge_id = ?", knowledgeId).
		Update("doc_size", gorm.Expr("doc_size + ?", docSize)).
		Update("doc_count", gorm.Expr("doc_count + ?", len(resultList))).Error
}

// UpdateKnowledgeDocCount 更新知识库文档数量
func UpdateKnowledgeDocCount(tx *gorm.DB, knowledgeId string) error {
	var total int64
	err := tx.Model(&model.KnowledgeQAPair{}).Where("knowledge_id = ?", knowledgeId).
		Count(&total).Error
	if err != nil {
		return err
	}
	return tx.Model(&model.KnowledgeBase{}).Where("knowledge_id = ?", knowledgeId).
		Update("doc_count", total).Error
}

// DeleteKnowledgeFileInfo 删除知识库文档信息
func DeleteKnowledgeFileInfo(tx *gorm.DB, knowledgeId string, resultList []*model.DocInfo) error {
	var docSize int64
	for _, result := range resultList {
		docSize += result.DocSize
	}
	return tx.Model(&model.KnowledgeBase{}).Where("knowledge_id = ?", knowledgeId).
		Update("doc_size", gorm.Expr("doc_size - ?", docSize)).
		Update("doc_count", gorm.Expr("doc_count - ?", len(resultList))).Error
}

// CreateKnowledgeReport 创建知识库社区报告
func CreateKnowledgeReport(ctx context.Context, knowledgeId string) error {
	knowledge, err := SelectKnowledgeById(ctx, knowledgeId, "", "")
	if err != nil {
		return err
	}
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.更新生成条数和状态
		err := tx.Model(&model.KnowledgeBase{}).Where("knowledge_id=?", knowledgeId).Update("report_create_count", gorm.Expr("report_create_count + ?", 1)).
			Update("report_status", model.ReportProcessing).Error
		if err != nil {
			return err
		}
		//构造知识库图谱
		knowledgeGraph := BuildKnowledgeGraph(string(knowledge.KnowledgeGraph))
		//2.通知rag生成社区报告
		return service.RagCreateKnowledgeReport(ctx, &service.RagImportDocParams{
			KnowledgeName:        knowledge.RagName,
			CategoryId:           knowledge.KnowledgeId,
			UserId:               knowledge.UserId,
			KnowledgeGraphSwitch: knowledgeGraph.KnowledgeGraphSwitch,
			GraphModelId:         knowledgeGraph.GraphModelId,
		})
	})
}

func createKnowledge(tx *gorm.DB, knowledge *model.KnowledgeBase) error {
	return tx.Create(knowledge).Error
}

func updateKnowledge(tx *gorm.DB, id uint32, name, description string) error {
	var updateParams = map[string]interface{}{
		"name":        name,
		"description": description,
	}
	return tx.Model(&model.KnowledgeBase{}).Where("id=?", id).Updates(updateParams).Error
}

func updateKnowledgeWithRagName(tx *gorm.DB, id uint32, name, ragName, description string) error {
	var updateParams = map[string]interface{}{
		"name":        name,
		"rag_name":    ragName,
		"description": description,
	}
	return tx.Model(&model.KnowledgeBase{}).Where("id=?", id).Updates(updateParams).Error
}

// 逻辑删除
func logicDeleteKnowledge(tx *gorm.DB, knowledge *model.KnowledgeBase) error {
	var updateParams = map[string]interface{}{
		"deleted": 1,
	}
	return tx.Model(&model.KnowledgeBase{}).Where("id=?", knowledge.Id).Updates(updateParams).Error
}

// buildKnowledgePermission 构建知识库权限信息
func buildKnowledgePermission(knowledge *model.KnowledgeBase) *model.KnowledgePermission {
	return &model.KnowledgePermission{
		PermissionId:   wanwu_util.NewID(),
		KnowledgeId:    knowledge.KnowledgeId,
		GrantUserId:    knowledge.UserId,
		GrantOrgId:     knowledge.OrgId,
		PermissionType: model.PermissionTypeSystem,
		CreatedAt:      knowledge.CreatedAt,
		UpdatedAt:      knowledge.UpdatedAt,
		UserId:         knowledge.UserId,
		OrgId:          knowledge.OrgId,
	}
}

func buildPermissionKnowledgeIdList(permissionList []*model.KnowledgePermission) []string {
	return lo.Map(permissionList, func(item *model.KnowledgePermission, index int) string {
		return item.KnowledgeId
	})
}

func buildPermissionKnowledgeIdMap(permissionList []*model.KnowledgePermission) map[string]int {
	var permissionMap = make(map[string]int)
	for _, permission := range permissionList {
		permissionMap[permission.KnowledgeId] = permission.PermissionType
	}
	return permissionMap
}

// intersectionKnowledgeIdList 计算两个知识库id 列表的交集
func intersectionKnowledgeIdList(knowledgeIdList, permissionKnowledgeIdList []string) []string {
	//特殊逻辑，如果用户没有指定tag，则返回用户有权限的知识库id列表
	if len(knowledgeIdList) == 0 {
		return permissionKnowledgeIdList
	}
	var knowledgeIdMap = make(map[string]bool)
	for _, permissionKnowledgeId := range permissionKnowledgeIdList {
		knowledgeIdMap[permissionKnowledgeId] = true
	}
	var retKnowledgeIdList []string
	for _, knowledgeId := range knowledgeIdList {
		if knowledgeIdMap[knowledgeId] {
			retKnowledgeIdList = append(retKnowledgeIdList, knowledgeId)
		}
	}
	return retKnowledgeIdList
}
