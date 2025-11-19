package orm

import (
	"context"
	"fmt"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm/sqlopt"
	async_task "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/async-task"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/generator"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

// SelectKnowledgeList 查询知识库列表 [EN] SelectKnowledgeList Query the knowledge base list
func SelectKnowledgeList(ctx context.Context, userId, orgId, name string, tagIdList []string) ([]*model.KnowledgeBase, map[string]int, error) {
	var knowledgeIdList []string
	var err error
	if len(tagIdList) > 0 {
		knowledgeIdList, err = SelectKnowledgeIdByTagId(ctx, tagIdList)
		if err != nil {
			return nil, nil, err
		}
	}
	//查询有权限的知识库列表，获取有权限的知识库id，目前是getALL，没有通过连表实现 [EN] Query the authorized knowledge base list and obtain the authorized knowledge base ID. Currently, it is getALL, which is not implemented through join tables.
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
	err = sqlopt.SQLOptions(sqlopt.WithKnowledgeIDList(knowledgeIdList), sqlopt.LikeName(name), sqlopt.WithDelete(0)).
		Apply(db.GetHandle(ctx), &model.KnowledgeBase{}).
		Order("create_at desc").
		Find(&knowledgeList).
		Error
	if err != nil {
		return nil, nil, err
	}
	return knowledgeList, buildPermissionKnowledgeIdMap(permissionKnowledgeList), nil
}

// SelectKnowledgeById 查询知识库信息,todo [EN] SelectKnowledgeById Query knowledge base information, todo
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

// SelectKnowledgeByIdList 查询知识库信息 [EN] SelectKnowledgeByIdList Query knowledge base information
func SelectKnowledgeByIdList(ctx context.Context, knowledgeIdList []string, userId, orgId string) ([]*model.KnowledgeBase, map[string]int, error) {
	//查询有权限的知识库列表，获取有权限的知识库id，目前是getALL，没有通过连表实现 [EN] Query the authorized knowledge base list and obtain the authorized knowledge base ID. Currently, it is getALL, which is not implemented through join tables.
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

// SelectKnowledgeByName 查询知识库信息 [EN] SelectKnowledgeByName Query knowledge base information
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

// SelectKnowledgeByIdNoDeleteCheck 查询知识库信息 [EN] SelectKnowledgeByIdNoDeleteCheck Query knowledge base information
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

// CheckSameKnowledgeName 知识库名称是否存在同名 [EN] CheckSameKnowledgeName Whether the knowledge base name has the same name
func CheckSameKnowledgeName(ctx context.Context, userId, orgId, name, knowledgeId string) error {
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

	list, _, err := SelectKnowledgeList(ctx, userId, orgId, name, nil)
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

// CreateKnowledge 创建知识库 [EN] CreateKnowledge Create a knowledge base
func CreateKnowledge(ctx context.Context, knowledge *model.KnowledgeBase, embeddingModelId string) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.插入数据 [EN] 1. Insert data
		err := createKnowledge(tx, knowledge)
		if err != nil {
			return err
		}
		//2.插入权限信息 [EN] 2. Insert permission information
		err = CreateKnowledgeIdPermission(tx, buildKnowledgePermission(knowledge))
		if err != nil {
			return err
		}
		//3.通知rag创建知识库 [EN] 3. Notify rag to create a knowledge base
		return service.RagKnowledgeCreate(ctx, &service.RagCreateParams{
			UserId:               knowledge.UserId,
			Name:                 knowledge.RagName,
			KnowledgeBaseId:      knowledge.KnowledgeId,
			EmbeddingModelId:     embeddingModelId,
			EnableKnowledgeGraph: knowledge.KnowledgeGraphSwitch > 0,
		})
	})
}

// UpdateKnowledge 更新知识库 [EN] UpdateKnowledge Update knowledge base
func UpdateKnowledge(ctx context.Context, name, description string, knowledgeBase *model.KnowledgeBase) error {
	//return updateKnowledge(db.GetHandle(ctx), knowledgeBase.Id, name, description)
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//已经区分为知识库展示名称和rag知识库名称，不需要再通知rag修改名称 [EN] It has been divided into knowledge base display name and rag knowledge base name. There is no need to notify rag to change the name.
		if knowledgeBase.Name != knowledgeBase.RagName {
			return updateKnowledge(tx, knowledgeBase.Id, name, description)
		}
		//2.更新数据 [EN] 2.Update data
		ragName := generator.GetGenerator().NewID()
		err := updateKnowledgeWithRagName(tx, knowledgeBase.Id, name, ragName, description)
		if err != nil {
			return err
		}

		//2.通知rag更新知识库,只有老的需要更新 [EN] 2. Notify rag to update the knowledge base, only the old ones need to be updated
		return service.RagKnowledgeUpdate(ctx, &service.RagUpdateParams{
			UserId:          knowledgeBase.UserId,
			KnowledgeBaseId: knowledgeBase.KnowledgeId,
			OldKbName:       knowledgeBase.RagName,
			NewKbName:       ragName,
		})
	})
}

// UpdateKnowledgeShareCount 更新知识库分享数量 [EN] UpdateKnowledgeShareCount updates the number of knowledge base shares
func UpdateKnowledgeShareCount(tx *gorm.DB, knowledgeId string, count int64) error {
	var updateParams = map[string]interface{}{
		"share_count": count,
	}
	return tx.Model(&model.KnowledgeBase{}).Where("knowledge_id=?", knowledgeId).Updates(updateParams).Error
}

// UpdateKnowledgeGraph 更新知识库图谱 [EN] UpdateKnowledgeGraph updates the knowledge base graph
func UpdateKnowledgeGraph(tx *gorm.DB, knowledgeId string, knowledgeGraph string) error {
	var updateParams = map[string]interface{}{
		"knowledge_graph": knowledgeGraph,
	}
	return tx.Model(&model.KnowledgeBase{}).Where("knowledge_id=?", knowledgeId).Updates(updateParams).Error
}

// UpdateKnowledgeReportStatus 更新社区报告状态 [EN] UpdateKnowledgeReportStatus updates community report status
func UpdateKnowledgeReportStatus(ctx context.Context, knowledgeId string, reportStatus int) error {
	var updateParams = map[string]interface{}{
		"report_status": model.ReportStatus(reportStatus),
	}
	return db.GetHandle(ctx).Model(&model.KnowledgeBase{}).Where("knowledge_id=?", knowledgeId).Updates(updateParams).Error
}

// DeleteKnowledge 删除知识库 [EN] DeleteKnowledge Delete knowledge base
func DeleteKnowledge(ctx context.Context, knowledgeBase *model.KnowledgeBase) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.逻辑删除数据 [EN] 1. Logically delete data
		err := logicDeleteKnowledge(tx, knowledgeBase)
		if err != nil {
			return err
		}
		//2.通知rag更新知识库 [EN] 2. Notify rag to update the knowledge base
		return async_task.SubmitTask(ctx, async_task.KnowledgeDeleteTaskType, &async_task.KnowledgeDeleteParams{
			KnowledgeId: knowledgeBase.KnowledgeId,
		})
	})
}

// ExecuteDeleteKnowledge 删除知识库 [EN] ExecuteDeleteKnowledge Delete knowledge base
func ExecuteDeleteKnowledge(tx *gorm.DB, id uint32) error {
	return tx.Unscoped().Model(&model.KnowledgeBase{}).Where("id = ?", id).Delete(&model.KnowledgeBase{}).Error
}

// UpdateKnowledgeFileInfo 更新知识库文档信息 [EN] UpdateKnowledgeFileInfo updates knowledge base document information
func UpdateKnowledgeFileInfo(tx *gorm.DB, knowledgeId string, resultList []*model.DocInfo) error {
	var docSize int64
	for _, result := range resultList {
		docSize += result.DocSize
	}
	return tx.Model(&model.KnowledgeBase{}).Where("knowledge_id = ?", knowledgeId).
		Update("doc_size", gorm.Expr("doc_size + ?", docSize)).
		Update("doc_count", gorm.Expr("doc_count + ?", len(resultList))).Error
}

// DeleteKnowledgeFileInfo 删除知识库文档信息 [EN] DeleteKnowledgeFileInfo deletes knowledge base document information
func DeleteKnowledgeFileInfo(tx *gorm.DB, knowledgeId string, resultList []*model.DocInfo) error {
	var docSize int64
	for _, result := range resultList {
		docSize += result.DocSize
	}
	return tx.Model(&model.KnowledgeBase{}).Where("knowledge_id = ?", knowledgeId).
		Update("doc_size", gorm.Expr("doc_size - ?", docSize)).
		Update("doc_count", gorm.Expr("doc_count - ?", len(resultList))).Error
}

// CreateKnowledgeReport 创建知识库社区报告 [EN] CreateKnowledgeReport creates a knowledge base community report
func CreateKnowledgeReport(ctx context.Context, knowledgeId string) error {
	knowledge, err := SelectKnowledgeById(ctx, knowledgeId, "", "")
	if err != nil {
		return err
	}
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.更新生成条数和状态 [EN] 1. Update the number and status of generated items
		err := tx.Model(&model.KnowledgeBase{}).Where("knowledge_id=?", knowledgeId).Update("report_create_count", gorm.Expr("report_create_count + ?", 1)).
			Update("report_status", model.ReportProcessing).Error
		if err != nil {
			return err
		}
		//构造知识库图谱 [EN] Constructing a knowledge base graph
		knowledgeGraph := BuildKnowledgeGraph(knowledge.KnowledgeGraph)
		//2.通知rag生成社区报告 [EN] 2. Notify rag to generate community report
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

// 逻辑删除 [EN] tombstone
func logicDeleteKnowledge(tx *gorm.DB, knowledge *model.KnowledgeBase) error {
	var updateParams = map[string]interface{}{
		"deleted": 1,
	}
	return tx.Model(&model.KnowledgeBase{}).Where("id=?", knowledge.Id).Updates(updateParams).Error
}

// buildKnowledgePermission 构建知识库权限信息 [EN] buildKnowledgePermission build knowledge base permission information
func buildKnowledgePermission(knowledge *model.KnowledgeBase) *model.KnowledgePermission {
	return &model.KnowledgePermission{
		PermissionId:   generator.GetGenerator().NewID(),
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

// intersectionKnowledgeIdList 计算两个知识库id 列表的交集 [EN] intersectionKnowledgeIdList calculates the intersection of two knowledge base id lists
func intersectionKnowledgeIdList(knowledgeIdList, permissionKnowledgeIdList []string) []string {
	//特殊逻辑，如果用户没有指定tag，则返回用户有权限的知识库id列表 [EN] Special logic, if the user does not specify a tag, a list of knowledge base IDs to which the user has permission is returned.
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
