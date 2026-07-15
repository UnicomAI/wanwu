package orm

import (
	"context"
	"errors"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	wanwu_util "github.com/UnicomAI/wanwu/pkg/util"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

// SelectKnowledgePermissionById 查询用户知识库权限
func SelectKnowledgePermissionById(ctx context.Context, permissionId string) (*model.KnowledgePermission, error) {
	var permission = model.KnowledgePermission{}
	err := sqlopt.SQLOptions(sqlopt.WithPermissionId(permissionId)).
		Apply(db.GetHandle(ctx), &model.KnowledgePermission{}).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// SelectKnowledgeIdByPermission 查询用户有权限的知识库id
func SelectKnowledgeIdByPermission(ctx context.Context, userId, orgId string, permission int, knowledgeIdList []string) ([]*model.KnowledgePermission, error) {
	var knowledgePermissionList []*model.KnowledgePermission
	err := sqlopt.SQLOptions(sqlopt.WithKnowledgeIDList(knowledgeIdList), sqlopt.WithPermit(orgId, userId), sqlopt.WithOverKnowledgePermission(permission)).
		Apply(db.GetHandle(ctx), &model.KnowledgePermission{}).Find(&knowledgePermissionList).Error
	if err != nil {
		return nil, err
	}
	return knowledgePermissionList, nil
}

// SelectKnowledgeIdByPermissionOrgList 查询用户有权限的知识库id
func SelectKnowledgeIdByPermissionOrgList(ctx context.Context, userId, orgId []string, permission int) ([]*model.KnowledgePermission, error) {
	var knowledgePermissionList []*model.KnowledgePermission
	err := sqlopt.SQLOptions(sqlopt.WithPermitList(orgId, userId), sqlopt.WithOverKnowledgePermission(permission)).
		Apply(db.GetHandle(ctx), &model.KnowledgePermission{}).Find(&knowledgePermissionList).Error
	if err != nil {
		return nil, err
	}
	return knowledgePermissionList, nil
}

// SelectUserKnowledgePermissionList 查询用户知识库权限
func SelectUserKnowledgePermissionList(ctx context.Context, knowledgeId string) ([]*model.KnowledgePermission, error) {
	var permissionList []*model.KnowledgePermission
	err := sqlopt.SQLOptions(sqlopt.WithKnowledgeID(knowledgeId)).
		Apply(db.GetHandle(ctx), &model.KnowledgePermission{}).Find(&permissionList).Error
	if err != nil {
		return nil, err
	}
	return permissionList, nil
}

// SelectUserKnowledgePermission 查询用户知识库权限
func SelectUserKnowledgePermission(ctx context.Context, userId, orgId string, knowledgeId string) (*model.KnowledgePermission, error) {
	var permission = model.KnowledgePermission{}
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithKnowledgeID(knowledgeId)).
		Apply(db.GetHandle(ctx), &model.KnowledgePermission{}).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// CreateKnowledgeIdPermission 创建知识库权限,有可能在事务中的一部分，所以方法第一个入参为db
func CreateKnowledgeIdPermission(db *gorm.DB, permission *model.KnowledgePermission) error {
	err := db.Model(&model.KnowledgePermission{}).Create(permission).Error
	if err != nil {
		return err
	}
	err = processKnowledgeShareCount(db, permission.KnowledgeId)
	if err != nil {
		return err
	}
	return recordKnowledgePermission(db, permission, model.PermissionTypeNone, permission.PermissionType)
}

// BatchCreateKnowledgeIdPermission 批量创建知识库权限,有可能在事务中的一部分，所以方法第一个入参为db
func BatchCreateKnowledgeIdPermission(db *gorm.DB, permission []*model.KnowledgePermission) error {
	if len(permission) == 0 {
		return nil
	}
	err := db.Model(&model.KnowledgePermission{}).CreateInBatches(permission, 50).Error
	if err != nil {
		return err
	}
	err = processKnowledgeShareCount(db, permission[0].KnowledgeId)
	if err != nil {
		return err
	}
	return batchRecordKnowledgePermission(db, permission)
}

func BatchEditKnowledgePermission(db *gorm.DB, permissionList []*model.KnowledgePermission) error {
	if len(permissionList) == 0 {
		return nil
	}
	permissionMap, err := buildPermissionMap(db, permissionList)
	if err != nil {
		return err
	}
	for _, permission := range permissionList {
		knowledgePermission := permissionMap[permission.PermissionId]
		if knowledgePermission == nil {
			continue
		}
		updateMap := map[string]interface{}{
			"permission_type": permission.PermissionType,
			"grant_user_id":   permission.GrantUserId,
			"grant_org_id":    permission.GrantOrgId,
		}
		err = db.Model(&model.KnowledgePermission{}).Where("permission_id = ?", permission.PermissionId).Updates(updateMap).Error
		if err != nil {
			return err
		}
		err = recordKnowledgePermission(db, knowledgePermission, knowledgePermission.PermissionType, permission.PermissionType)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteKnowledgeIdPermission 删除知识库权限,有可能在事务中的一部分，所以方法第一个入参为db
func DeleteKnowledgeIdPermission(db *gorm.DB, knowledgeId string, permissionIdList []string) error {
	var permissionList []*model.KnowledgePermission
	if err := db.Model(&model.KnowledgePermission{}).Where("knowledge_id = ?", knowledgeId).
		Where("permission_id IN ?", permissionIdList).
		Find(&permissionList).Error; err != nil {
		return err
	}

	err := processKnowledgeShareCount(db, knowledgeId)
	if err != nil {
		return err
	}
	return batchDeletePermissionByIdList(db, knowledgeId, permissionList)
}

// DeleteKnowledgePermissionById 删除知识库权限,有可能在事务中的一部分，所以方法第一个入参为db
func DeleteKnowledgePermissionById(ctx context.Context, permission *model.KnowledgePermission) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		// 删除
		if err := tx.Unscoped().Where("permission_id = ?", permission.PermissionId).
			Delete(&model.KnowledgePermission{}).Error; err != nil {
			return err
		}
		err := processKnowledgeShareCount(tx, permission.KnowledgeId)
		if err != nil {
			return err
		}
		return batchDeleteRecordKnowledgePermission(tx, []*model.KnowledgePermission{
			permission,
		})
	})
}

// DeleteKnowledgePermissionByUser 删除知识库权限,有可能在事务中的一部分，所以方法第一个入参为db
func DeleteKnowledgePermissionByUser(tx *gorm.DB, knowledgeId, userId, orgId string) error {
	var count int64
	var permission = model.KnowledgePermission{}
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithKnowledgeID(knowledgeId)).
		Apply(tx, &model.KnowledgePermission{}).First(&permission).Count(&count).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if count > 0 {
		if err = tx.Unscoped().Where("permission_id = ?", permission.PermissionId).
			Delete(&model.KnowledgePermission{}).Error; err != nil {
			return err
		}
		err = processKnowledgeShareCount(tx, permission.KnowledgeId)
		if err != nil {
			return err
		}
		return batchDeleteRecordKnowledgePermission(tx, []*model.KnowledgePermission{
			&permission,
		})
	}
	return nil
}

// processKnowledgeShareCount 处理知识库分享数量
func processKnowledgeShareCount(tx *gorm.DB, knowledgeId string) error {
	var totalCount int64
	err := sqlopt.SQLOptions(sqlopt.WithKnowledgeID(knowledgeId)).
		Apply(tx, &model.KnowledgePermission{}).Count(&totalCount).Error
	if err != nil {
		return err
	}
	return UpdateKnowledgeShareCount(tx, knowledgeId, totalCount)
}

// AsyncDeletePermissionByKnowledgeId 删除知识库权限,有可能在事务中的一部分，所以方法第一个入参为db
func AsyncDeletePermissionByKnowledgeId(knowledgeId string) error {
	go func() {
		err := pageDeletePermission(knowledgeId)
		if err != nil {
			log.Errorf("pageDeletePermission err: %s", err)
		}
	}()
	return nil
}

// pageDeletePermission 分页删除知识库权限
func pageDeletePermission(knowledgeId string) (err error) {
	var limit = 30
	var lastPermissionId = "000"
	//最多处理30w数据，理论上一个书知识库关系应该少于30w
	for i := 0; i < 10000; i++ {
		var permissionList []*model.KnowledgePermission
		if err = db.GetHandle(context.Background()).Model(&model.KnowledgePermission{}).
			Where("knowledge_id = ?", knowledgeId).
			Where("permission_id > ?", lastPermissionId).
			Order("permission_id").
			Limit(limit).
			Find(&permissionList).Error; err != nil {
			return err
		}
		if len(permissionList) < limit {
			err = batchDeletePermissionByIdList(db.GetHandle(context.Background()), knowledgeId, permissionList)
			break
		}
		err = batchDeletePermissionByIdList(db.GetHandle(context.Background()), knowledgeId, permissionList)
		lastPermissionId = permissionList[len(permissionList)-1].PermissionId
	}
	return err
}

func buildPermissionMap(db *gorm.DB, permissionList []*model.KnowledgePermission) (map[string]*model.KnowledgePermission, error) {
	var dataList []*model.KnowledgePermission
	err := db.Model(&model.KnowledgePermission{}).Where("permission_id IN ?", lo.Map(permissionList, func(item *model.KnowledgePermission, index int) string {
		return item.PermissionId
	})).Find(&dataList).Error
	if err != nil {
		return nil, err
	}
	var permissionMap = make(map[string]*model.KnowledgePermission)
	for _, item := range dataList {
		permissionMap[item.PermissionId] = item
	}
	return permissionMap, nil
}

// recordKnowledgePermission 记录知识库权限操作，外层保证事务
func recordKnowledgePermission(db *gorm.DB, permission *model.KnowledgePermission, fromPermissionType, toPermissionType int) error {
	return db.Model(&model.KnowledgePermissionRecord{}).Create(buildPermissionRecord(permission, fromPermissionType, toPermissionType)).Error
}

// batchRecordKnowledgePermission 记录知识库权限操作，外层保证事务
func batchRecordKnowledgePermission(db *gorm.DB, permission []*model.KnowledgePermission) error {
	var permissionRecord []*model.KnowledgePermissionRecord
	for _, knowledgePermission := range permission {
		permissionRecord = append(permissionRecord, buildPermissionRecord(knowledgePermission, model.PermissionTypeNone, knowledgePermission.PermissionType))
	}
	return db.Model(&model.KnowledgePermissionRecord{}).CreateInBatches(permissionRecord, 50).Error
}

func batchDeletePermissionByIdList(db *gorm.DB, knowledgeId string, permissionList []*model.KnowledgePermission) error {
	permissionIdList := lo.Map(permissionList, func(item *model.KnowledgePermission, index int) string {
		return item.PermissionId
	})
	// 删除
	if err := db.Unscoped().Where("knowledge_id = ?", knowledgeId).
		Where("permission_id IN ?", permissionIdList).
		Delete(&model.KnowledgePermission{}).Error; err != nil {
		return err
	}
	return batchDeleteRecordKnowledgePermission(db, permissionList)
}

// batchDeleteRecordKnowledgePermission 记录知识库权限操作，外层保证事务
func batchDeleteRecordKnowledgePermission(db *gorm.DB, permission []*model.KnowledgePermission) error {
	var permissionRecord []*model.KnowledgePermissionRecord
	for _, knowledgePermission := range permission {
		permissionRecord = append(permissionRecord, buildPermissionRecord(knowledgePermission, knowledgePermission.PermissionType, model.PermissionTypeNone))
	}
	return db.Model(&model.KnowledgePermissionRecord{}).CreateInBatches(permissionRecord, 50).Error
}

func buildPermissionRecord(permission *model.KnowledgePermission, fromPermissionType, toPermissionType int) *model.KnowledgePermissionRecord {
	var option int
	if fromPermissionType == model.PermissionTypeNone {
		option = model.RecordOptionAdd
	} else if toPermissionType == model.PermissionTypeNone {
		option = model.RecordOptionDelete
	} else {
		option = model.RecordOptionEdit
	}
	milli := time.Now().UnixMilli()
	return &model.KnowledgePermissionRecord{
		RecordId:           wanwu_util.NewID(),
		KnowledgeId:        permission.KnowledgeId,
		Option:             option,
		OperatorUserId:     permission.GrantUserId,
		OperatorOrgId:      permission.GrantOrgId,
		FromPermissionType: fromPermissionType,
		ToPermissionType:   toPermissionType,
		OwnerOrgId:         permission.OrgId,
		OwnerUserId:        permission.UserId,
		CreatedAt:          milli,
		UpdatedAt:          milli,
	}
}
