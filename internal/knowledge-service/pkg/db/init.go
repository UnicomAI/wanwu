package db

import (
	"context"
	"fmt"
	"time"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/generator"
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	"gorm.io/gorm"
)

const (
	knowledgeDBName     = "knowledge_base_service"
	docMetaTimestampOld = "1757692799000" //2025-09-12 23:59:59
)

var dbClient = DataBaseClient{}

type DataBaseClient struct {
	DB *gorm.DB
}

func init() {
	pkg.AddContainer(dbClient)
}

func GetClient() DataBaseClient {
	return dbClient
}

func GetHandle(ctx context.Context) *gorm.DB {
	return GetClient().DB.WithContext(ctx)
}

func (c DataBaseClient) LoadType() string {
	return "dbClient"
}

func (c DataBaseClient) Load() error {
	dbHandle, err := db.New(config.GetConfig().DB)
	if err != nil {
		log.Errorf("init knowledge_base_service db err: %v", err)
		return err
	}
	//创建数据库配置 [EN] Create database configuration
	err = createDB(dbHandle)
	if err != nil {
		return err
	}
	//注册表配置 [EN] Registry configuration
	err = registerTables(dbHandle)
	if err != nil {
		return err
	}
	//初始化数据 [EN] initialization data
	err = initData(dbHandle)
	if err != nil {
		return err
	}
	err = initKnowledgeName(dbHandle)
	if err != nil {
		return err
	}
	dbClient.DB = dbHandle
	return nil
}

func (c DataBaseClient) StopPriority() int {
	return pkg.DBPriority
}

func (c DataBaseClient) Stop() error {
	if dbClient.DB == nil {
		return nil
	}
	dbHandle, err := dbClient.DB.DB()
	if err != nil {
		return err
	}
	err = dbHandle.Close()
	return err
}

// 创建db [EN] createdb
func createDB(dbClient *gorm.DB) error {
	err := dbClient.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;", knowledgeDBName)).Error
	if err != nil {
		log.Errorf("MySQL创建数据库%s异常: %v", knowledgeDBName, err)
		return err
	}
	log.Infof("MySQL创建数据库成功: %s", knowledgeDBName)
	return nil
}

// 注册表信息 [EN] Registry information
func registerTables(dbClient *gorm.DB) error {
	err := dbClient.AutoMigrate(
		model.KnowledgeDoc{},
		model.KnowledgeBase{},
		model.KnowledgeImportTask{},
		model.KnowledgeTag{},
		model.KnowledgeTagRelation{},
		model.KnowledgeKeywords{},
		model.KnowledgeSplitter{},
		model.KnowledgeDocMeta{},
		model.DocSegmentImportTask{},
		model.KnowledgePermission{},
		model.KnowledgePermissionRecord{},
	)
	if err != nil {
		fmt.Printf("register knowledge tables failed: %v", err)
		return err
	}
	fmt.Printf("register knowledge tables table success")
	return nil
}

// initData初始化数据 [EN] initData initialization data
func initData(dbClient *gorm.DB) error {
	err := initDocMeta(dbClient)
	if err != nil {
		log.Errorf("init doc meta failed: %v", err)
		return err
	}
	err = initKnowledgePermission(dbClient)
	if err != nil {
		log.Errorf("init knowledge permission failed: %v", err)
		return err
	}
	return nil
}

func initDocMeta(dbClient *gorm.DB) error {
	var knowledgeDocMetaList []model.KnowledgeDocMeta
	//数据量不会太大直接getAll [EN] The amount of data will not be too large, just getAll
	err := dbClient.Model(&model.KnowledgeDocMeta{}).Where("create_at <= ?", docMetaTimestampOld).Find(&knowledgeDocMetaList).Error
	if err != nil {
		return err
	}
	if len(knowledgeDocMetaList) > 0 {
		for _, meta := range knowledgeDocMetaList {
			if len(meta.KnowledgeId) > 0 {
				continue
			}
			if len(meta.DocId) > 0 {
				var knowledgeDocList []model.KnowledgeDoc
				_ = dbClient.Model(&model.KnowledgeDoc{}).Where("doc_id = ?", meta.DocId).Find(&knowledgeDocList).Error
				if len(knowledgeDocList) > 0 {
					err = dbClient.Model(&model.KnowledgeDocMeta{}).Where("id = ?", meta.Id).
						Updates(map[string]interface{}{"knowledge_id": knowledgeDocList[0].KnowledgeId}).Error
					if err != nil {
						log.Errorf("update knowledge_doc_meta error: %v", err)
					}
				}
			}

		}
	}
	return nil
}

func initKnowledgeName(dbClient *gorm.DB) error {
	var knowledgeBaseList []model.KnowledgeBase
	//数据量不会太大直接getAll [EN] The amount of data will not be too large, just getAll
	err := dbClient.Model(&model.KnowledgeBase{}).Find(&knowledgeBaseList).Error
	if err != nil {
		return err
	}

	for _, knowledgeBase := range knowledgeBaseList {
		if len(knowledgeBase.RagName) > 0 {
			continue
		}
		updateMap := map[string]interface{}{
			"rag_name": knowledgeBase.Name,
		}
		err = dbClient.Model(&model.KnowledgeBase{}).Where("knowledge_id = ?", knowledgeBase.KnowledgeId).Updates(updateMap).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func initKnowledgePermission(dbClient *gorm.DB) error {
	var knowledgeBaseList []model.KnowledgeBase
	//数据量不会太大直接getAll [EN] The amount of data will not be too large, just getAll
	err := dbClient.Model(&model.KnowledgeBase{}).Find(&knowledgeBaseList).Error
	if err != nil {
		return err
	}
	var knowledgePermissionList []model.KnowledgePermission
	err = dbClient.Model(&model.KnowledgePermission{}).Find(&knowledgePermissionList).Error
	if err != nil {
		return err
	}
	var knowledgePermissionMap = make(map[string]bool)
	for _, knowledgePermission := range knowledgePermissionList {
		knowledgePermissionMap[knowledgePermission.KnowledgeId] = true
	}
	var permissionList []*model.KnowledgePermission
	timeNow := time.Now().UnixMilli()
	for _, knowledgeBase := range knowledgeBaseList {
		if knowledgePermissionMap[knowledgeBase.KnowledgeId] {
			continue
		}
		permissionList = append(permissionList, &model.KnowledgePermission{
			PermissionId:   generator.GetGenerator().NewID(),
			KnowledgeId:    knowledgeBase.KnowledgeId,
			UserId:         knowledgeBase.UserId,
			OrgId:          knowledgeBase.OrgId,
			GrantUserId:    knowledgeBase.UserId,
			GrantOrgId:     knowledgeBase.OrgId,
			PermissionType: model.PermissionTypeSystem,
			CreatedAt:      timeNow,
			UpdatedAt:      timeNow,
		})
	}
	err = dbClient.Model(&model.KnowledgePermission{}).CreateInBatches(permissionList, 30).Error
	if err != nil {
		return err
	}
	return nil
}
