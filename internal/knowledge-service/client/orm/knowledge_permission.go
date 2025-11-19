package orm

import (
	"context"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/generator"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

// SelectKnowledgePermissionById Query user knowledge base permissions
func SelectKnowledgePermissionById(ctx context.Context, permissionId string) (*model.KnowledgePermission, error) {
	var permission = model.KnowledgePermission{}
	err := sqlopt.SQLOptions(sqlopt.WithPermissionId(permissionId)).
		Apply(db.GetHandle(ctx), &model.KnowledgePermission{}).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// SelectKnowledgeIdByPermission queries the knowledge base id that the user has permission to
func SelectKnowledgeIdByPermission(ctx context.Context, userId, orgId string, permission int) ([]*model.KnowledgePermission, error) {
	var knowledgePermissionList []*model.KnowledgePermission
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithOverKnowledgePermission(permission)).
		Apply(db.GetHandle(ctx), &model.KnowledgePermission{}).Find(&knowledgePermissionList).Error
	if err != nil {
		return nil, err
	}
	return knowledgePermissionList, nil
}

// SelectUserKnowledgePermissionList Query user knowledge base permissions
func SelectUserKnowledgePermissionList(ctx context.Context, knowledgeId string) ([]*model.KnowledgePermission, error) {
	var permissionList []*model.KnowledgePermission
	err := sqlopt.SQLOptions(sqlopt.WithKnowledgeID(knowledgeId)).
		Apply(db.GetHandle(ctx), &model.KnowledgePermission{}).Find(&permissionList).Error
	if err != nil {
		return nil, err
	}
	return permissionList, nil
}

// SelectUserKnowledgePermission Query user knowledge base permissions
func SelectUserKnowledgePermission(ctx context.Context, userId, orgId string, knowledgeId string) (*model.KnowledgePermission, error) {
	var permission = model.KnowledgePermission{}
	err := sqlopt.SQLOptions(sqlopt.WithPermit(orgId, userId), sqlopt.WithKnowledgeID(knowledgeId)).
		Apply(db.GetHandle(ctx), &model.KnowledgePermission{}).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// CreateKnowledgeIdPermission creates the knowledge base permission, which may be part of the transaction, so the first parameter of the method is db
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

// BatchCreateKnowledgeIdPermission batch creates knowledge base permissions, which may be part of the transaction, so the first parameter of the method is db
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

// DeleteKnowledgeIdPermission deletes the knowledge base permission, which may be part of the transaction, so the first parameter of the method is db
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

// DeleteKnowledgePermissionById deletes the knowledge base permission, which may be part of the transaction, so the first parameter of the method is db
func DeleteKnowledgePermissionById(ctx context.Context, permission *model.KnowledgePermission) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		// delete
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

// processKnowledgeShareCount processes the number of knowledge base shares
func processKnowledgeShareCount(tx *gorm.DB, knowledgeId string) error {
	var totalCount int64
	err := sqlopt.SQLOptions(sqlopt.WithKnowledgeID(knowledgeId)).
		Apply(tx, &model.KnowledgePermission{}).Count(&totalCount).Error
	if err != nil {
		return err
	}
	return UpdateKnowledgeShareCount(tx, knowledgeId, totalCount)
}

// AsyncDeletePermissionByKnowledgeId deletes the knowledge base permission, which may be part of the transaction, so the first parameter of the method is db
func AsyncDeletePermissionByKnowledgeId(knowledgeId string) error {
	go func() {
		err := pageDeletePermission(knowledgeId)
		if err != nil {
			log.Errorf("pageDeletePermission err: %s", err)
		}
	}()
	return nil
}

// pageDeletePermission Delete knowledge base permissions by page
func pageDeletePermission(knowledgeId string) (err error) {
	var limit = 30
	var lastPermissionId = "000"
	//It can process up to 300,000 data. In theory, a book knowledge base relationship should be less than 300,000.
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

// recordKnowledgePermission records knowledge base permission operations, and the outer layer guarantees transactions
func recordKnowledgePermission(db *gorm.DB, permission *model.KnowledgePermission, fromPermissionType, toPermissionType int) error {
	return db.Model(&model.KnowledgePermissionRecord{}).Create(buildPermissionRecord(permission, fromPermissionType, toPermissionType)).Error
}

// batchRecordKnowledgePermission records knowledge base permission operations, and the outer layer guarantees transactions
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
	// delete
	if err := db.Unscoped().Where("knowledge_id = ?", knowledgeId).
		Where("permission_id IN ?", permissionIdList).
		Delete(&model.KnowledgePermission{}).Error; err != nil {
		return err
	}
	return batchDeleteRecordKnowledgePermission(db, permissionList)
}

// batchDeleteRecordKnowledgePermission records knowledge base permission operations, and the outer layer guarantees transactions
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
		RecordId:           generator.GetGenerator().NewID(),
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
