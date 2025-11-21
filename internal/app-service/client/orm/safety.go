package orm

import (
	"context"
	"errors"
	"fmt"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/internal/app-service/pkg"
	"github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

const (
	AppSafetySensitiveUploadSingle     = "single"
	AppSafetySensitiveUploadFile       = "file"
	MaxSensitiveUploadSize         int = 100
)

func (c *Client) CreateSensitiveWordTable(ctx context.Context, userId, orgId, tableName, remark string) (string, *errs.Status) {
	err := sqlopt.SQLOptions(
		sqlopt.WithOrgID(orgId),
		sqlopt.WithUserID(userId),
		sqlopt.WithName(tableName),
	).Apply(c.db.WithContext(ctx)).First(&model.SensitiveWordTable{}).Error
	if err == nil {
		return "", toErrStatus("app_safety_sensitive_table_exist")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", toErrStatus("app_safety_sensitive_table_get", tableName)
	}
	table := &model.SensitiveWordTable{
		Name:    tableName,
		Remark:  remark,
		Version: getSensitiveTableVersion(),
		UserID:  userId,
		OrgID:   orgId,
	}
	if err := c.db.WithContext(ctx).Create(table).Error; err != nil {
		return "", toErrStatus("app_safety_sensitive_table_create", tableName, err.Error())
	}
	return util.Int2Str(table.ID), nil
}

func (c *Client) UpdateSensitiveWordTable(ctx context.Context, tableId uint32, tableName, remark string) *errs.Status {
	var existingTable model.SensitiveWordTable
	err := sqlopt.SQLOptions(
		sqlopt.WithName(tableName),
	).Apply(c.db.WithContext(ctx)).First(&existingTable).Error
	if err == nil && existingTable.ID != tableId {
		return toErrStatus("app_safety_sensitive_table_exist")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return toErrStatus("app_safety_sensitive_table_get", tableName)
	}
	updates := map[string]interface{}{
		"name":   tableName,
		"remark": remark,
	}
	updateErr := sqlopt.WithID(tableId).
		Apply(c.db.WithContext(ctx)).
		Model(&model.SensitiveWordTable{}).
		Updates(updates).Error

	if updateErr != nil {
		return toErrStatus("app_safety_sensitive_table_update", util.Int2Str(tableId), updateErr.Error())
	}
	return nil
}

func (c *Client) UpdateSensitiveWordTableReply(ctx context.Context, tableId uint32, reply string) *errs.Status {
	var table model.SensitiveWordTable
	if err := sqlopt.WithID(tableId).Apply(c.db.WithContext(ctx)).Model(&table).
		Updates(map[string]interface{}{
			"reply":   reply,
			"version": getSensitiveTableVersion(),
		}).Error; err != nil {
		return toErrStatus("app_safety_sensitive_table_reply_update", util.Int2Str(tableId), err.Error())
	}
	return nil
}

func (c *Client) DeleteSensitiveWordTable(ctx context.Context, tableId uint32) *errs.Status {
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := sqlopt.SQLOptions(
			sqlopt.WithTableID(tableId),
		).Apply(tx).Delete(&model.SensitiveWordVocabulary{}).Error; err != nil {
			return fmt.Errorf("failed to delete sensitiveWordVocabulary: %v", err)
		}
		if err := sqlopt.SQLOptions(
			sqlopt.WithID(tableId),
		).Apply(tx).Delete(&model.SensitiveWordTable{}).Error; err != nil {
			return fmt.Errorf("failed to delete sensitiveWordTable: %v", err)
		}
		return nil
	})
	if err != nil {
		return toErrStatus("app_safety_sensitive_table_delete", util.Int2Str(tableId), err.Error())
	}
	return nil
}

func (c *Client) GetSensitiveWordTableList(ctx context.Context, userId, orgId string) ([]*model.SensitiveWordTable, *errs.Status) {
	var tables []*model.SensitiveWordTable
	if err := sqlopt.SQLOptions(
		sqlopt.WithOrgID(orgId),
		sqlopt.WithUserID(userId),
	).Apply(c.db.WithContext(ctx)).Find(&tables).Error; err != nil {
		return nil, toErrStatus("app_safety_sensitive_table_list_get", err.Error())
	}
	return tables, nil
}

func (c *Client) GetSensitiveVocabularyList(ctx context.Context, tableId uint32, offset, limit int32) ([]*model.SensitiveWordVocabulary, int64, *errs.Status) {
	var vocabularies []*model.SensitiveWordVocabulary
	var count int64
	// Query paging data
	if err := sqlopt.SQLOptions(
		sqlopt.WithTableID(tableId),
	).Apply(c.db.WithContext(ctx)).Offset(int(offset)).Limit(int(limit)).Order("id DESC").Find(&vocabularies).
		Offset(-1).Limit(-1).Count(&count).Error; err != nil {
		return nil, 0, toErrStatus("app_safety_sensitive_vocabulary_list_get", util.Int2Str(tableId), err.Error())
	}
	return vocabularies, count, nil
}

func (c *Client) UploadSensitiveVocabulary(ctx context.Context, userId, orgId, importType, word, sensitiveType, filePath string, tableId uint32) *errs.Status {
	var words []*model.SensitiveWordVocabulary
	if err := sqlopt.WithTableID(tableId).Apply(c.db.WithContext(ctx)).Find(&words).Error; err != nil {
		return toErrStatus("app_safety_sensitive_vocabulary_list_get", util.Int2Str(tableId), err.Error())
	}
	if len(words) >= MaxSensitiveUploadSize {
		return toErrStatus("app_safety_sensitive_table_full", util.Int2Str(MaxSensitiveUploadSize))
	}
	// single upload
	if importType == AppSafetySensitiveUploadSingle {
		var existingRecord model.SensitiveWordVocabulary
		err := sqlopt.SQLOptions(
			sqlopt.WithTableID(tableId),
			sqlopt.WithContent(word),
		).Apply(c.db.WithContext(ctx)).First(&existingRecord).Error
		if err == nil {
			return toErrStatus("app_safety_sensitive_vocabulary_exist", word)
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newRecord := &model.SensitiveWordVocabulary{
				OrgID:         orgId,
				UserID:        userId,
				Content:       word,
				SensitiveType: sensitiveType,
				TableID:       util.Int2Str(tableId),
			}
			err = c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
				if err := tx.Create(newRecord).Error; err != nil {
					return fmt.Errorf("create sensitive word failed: %w", err)
				}
				if err := sqlopt.WithID(tableId).Apply(tx).Model(&model.SensitiveWordTable{}).
					Update("version", getSensitiveTableVersion()).Error; err != nil {
					return fmt.Errorf("update table version failed: %w", err)
				}
				return nil
			})
			if err != nil {
				return toErrStatus("app_safety_sensitive_vocabulary_create", word, err.Error())
			}
			return nil
		}
		return toErrStatus("app_safety_sensitive_vocabulary_create", word, err.Error())
	}
	// 1. Download files from MinIO to memory
	fileData, err := minio.DownloadFileToMemory(ctx, filePath)
	if err != nil {
		return toErrStatus("app_safety_sensitive_download_fail", err.Error())
	}
	// 2. Parse Excel files
	sensitiveWords, parseErr := pkg.ParseSensitiveExcel(fileData)
	if parseErr != nil {
		return toErrStatus("app_safety_sensitive_download_fail", parseErr.Error())
	}
	// 2. Construct a quick search mapping of existing terms
	existingMap := make(map[string]bool, len(words))
	for _, word := range words {
		existingMap[word.Content] = true
	}
	// 3. Construct and directly filter sensitive word data
	filteredWords := make([]*model.SensitiveWordVocabulary, 0, len(sensitiveWords))
	for _, raw := range sensitiveWords {
		// Skip duplicate entries
		if existingMap[raw.Content] {
			continue
		}
		filteredWords = append(filteredWords, &model.SensitiveWordVocabulary{
			TableID:       util.Int2Str(tableId),
			SensitiveType: raw.SensitiveType,
			Content:       raw.Content,
			UserID:        userId,
			OrgID:         orgId,
		})
	}
	if len(filteredWords) == 0 {
		return nil
	}
	// 4. Calculate the effective data amount
	remaining := MaxSensitiveUploadSize - len(words)
	if remaining < len(filteredWords) {
		return toErrStatus("app_safety_sensitive_table_full", util.Int2Str(MaxSensitiveUploadSize))
	}
	// 5. Insert data in batches
	err = c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.SensitiveWordVocabulary{}).
			Create(filteredWords).Error; err != nil {
			return fmt.Errorf("batch create failed: %w", err)
		}
		if err := sqlopt.WithID(tableId).Apply(tx).Model(&model.SensitiveWordTable{}).
			Update("version", getSensitiveTableVersion()).Error; err != nil {
			return fmt.Errorf("update table version failed: %w", err)
		}
		return nil
	})
	if err != nil {
		return toErrStatus("app_safety_sensitive_word_file_create_err", util.Int2Str(tableId), err.Error())
	}
	return nil
}

func (c *Client) DeleteSensitiveVocabulary(ctx context.Context, tableId, wordId uint32) *errs.Status {
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := sqlopt.SQLOptions(
			sqlopt.WithID(tableId),
		).Apply(tx).Model(&model.SensitiveWordTable{}).
			Update("version", getSensitiveTableVersion()).Error; err != nil {
			return fmt.Errorf("update table version failed: %w", err)
		}
		if err := sqlopt.SQLOptions(
			sqlopt.WithTableID(tableId),
			sqlopt.WithID(wordId),
		).Apply(tx).Delete(&model.SensitiveWordVocabulary{}).Error; err != nil {
			return fmt.Errorf("failed to delete sensitiveWordVocabulary: %v", err)
		}
		return nil
	})
	if err != nil {
		return toErrStatus("app_safety_sensitive_vocabulary_delete", util.Int2Str(wordId), err.Error())
	}
	return nil
}

func (c *Client) GetSensitiveWordTableListWithWordsByIDs(ctx context.Context, tableIds []string) ([]*SensitiveWordTableWithWord, *errs.Status) {
	var vocabularies []*model.SensitiveWordVocabulary
	if err := sqlopt.WithTableIDs(tableIds).Apply(c.db.WithContext(ctx)).
		Find(&vocabularies).Error; err != nil {
		return nil, toErrStatus("app_safety_sensitive_vocabulary_list_get_by_ids", err.Error())
	}
	var tables []*model.SensitiveWordTable
	if err := sqlopt.WithIDs(tableIds).Apply(c.db.WithContext(ctx)).
		Find(&tables).Error; err != nil {
		return nil, toErrStatus("app_safety_sensitive_table_list_get", err.Error())
	}
	result := make([]*SensitiveWordTableWithWord, 0, len(tables))

	for _, t := range tables {
		tableID := util.Int2Str(t.ID)
		item := &SensitiveWordTableWithWord{
			SensitiveWordTable: *t,
			SensitiveWords:     make([]string, 0),
		}
		for _, v := range vocabularies {
			if tableID == v.TableID {
				item.SensitiveWords = append(item.SensitiveWords, v.Content)
			}
		}
		result = append(result, item)
	}
	return result, nil
}

func (c *Client) GetSensitiveWordTableListByIDs(ctx context.Context, tableIds []string) ([]*model.SensitiveWordTable, *errs.Status) {
	var tables []*model.SensitiveWordTable
	if err := sqlopt.SQLOptions(
		sqlopt.WithIDs(tableIds),
	).Apply(c.db.WithContext(ctx)).Find(&tables).Error; err != nil {
		return nil, toErrStatus("app_safety_sensitive_table_list_get", err.Error())
	}
	return tables, nil
}

func (c *Client) GetSensitiveWordTableByID(ctx context.Context, tableId uint32) (*model.SensitiveWordTable, *errs.Status) {
	var table model.SensitiveWordTable
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(tableId),
	).Apply(c.db.WithContext(ctx)).First(&table).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, toErrStatus("app_safety_sensitive_table_not_found", util.Int2Str(tableId))
		}
		return nil, toErrStatus("app_safety_sensitive_table_get", util.Int2Str(tableId), err.Error())
	}
	return &table, nil
}
func getSensitiveTableVersion() string {
	return util.Int2Str(time.Now().UnixMilli())
}
