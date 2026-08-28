package orm

import (
	"context"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_client "github.com/UnicomAI/wanwu/internal/model-service/client/model"
	"github.com/UnicomAI/wanwu/internal/model-service/client/orm/sqlopt"
	"gorm.io/gorm"
)

const maxTitleLen = 512

func (c *Client) SaveModelExperienceDialog(ctx context.Context, dialog *model_client.ModelExperienceDialog) (*model_client.ModelExperienceDialog, *errs.Status) {
	//  检查dialog.Title长度
	runes := []rune(dialog.Title)
	if len(runes) >= maxTitleLen {
		dialog.Title = string(runes[:maxTitleLen-1]) + "…"
	}
	// create
	if err := sqlopt.WithSessionID(dialog.SessionId).
		Apply(c.db).WithContext(ctx).First(&model_client.ModelExperienceDialog{}).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, toErrStatus("model_experience_dialog_create_err", err.Error())
		}
		if err := c.db.WithContext(ctx).Create(dialog).Error; err != nil {
			return nil, toErrStatus("model_experience_dialog_create_err", err.Error())
		}
		return dialog, nil
	}
	// update
	if err := sqlopt.WithSessionID(dialog.SessionId).
		Apply(c.db).WithContext(ctx).Model(&model_client.ModelExperienceDialog{}).
		Updates(map[string]interface{}{
			"model_setting": dialog.ModelSetting,
			"updated_at":    time.Now().UnixMilli(),
		}).Error; err != nil {
		return nil, toErrStatus("model_experience_dialog_update_err", err.Error())
	}
	// get
	var ret model_client.ModelExperienceDialog
	if err := sqlopt.WithSessionID(dialog.SessionId).
		Apply(c.db).WithContext(ctx).First(&ret).Error; err != nil {
		return nil, toErrStatus("model_experience_dialog_update_err", err.Error())
	}
	return &ret, nil
}

func (c *Client) GetModelExperienceDialog(ctx context.Context, userId, orgId string, modelExperienceId uint32) (*model_client.ModelExperienceDialog, *errs.Status) {
	dialog := &model_client.ModelExperienceDialog{}
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(modelExperienceId),
		sqlopt.WithOrgID(orgId),
		sqlopt.WithUserID(userId),
	).Apply(c.db).WithContext(ctx).First(dialog).Error; err != nil {
		return nil, toErrStatus("model_experience_dialog_get_err", err.Error())
	}
	return dialog, nil
}

func (c *Client) ListModelExperienceDialogs(ctx context.Context, userId, orgId string) ([]*model_client.ModelExperienceDialog, *errs.Status) {
	var dialogs []*model_client.ModelExperienceDialog
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userId),
	).Apply(c.db.WithContext(ctx)).Order("updated_at desc").Find(&dialogs).Error; err != nil {
		return nil, toErrStatus("model_experience_dialog_list_err", err.Error())
	}
	return dialogs, nil
}

func (c *Client) DeleteModelExperienceDialog(ctx context.Context, userId, orgId string, modelExperienceId uint32) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// delete dialog
		if err := sqlopt.SQLOptions(
			sqlopt.WithID(modelExperienceId),
			sqlopt.WithOrgID(orgId),
			sqlopt.WithUserID(userId),
		).Apply(tx).Delete(&model_client.ModelExperienceDialog{}).Error; err != nil {
			return toErrStatus("model_experience_dialog_delete_err", err.Error())
		}
		// delete dialog records
		if err := sqlopt.WithModelExperienceId(modelExperienceId).
			Apply(tx).Delete(&model_client.ModelExperienceDialogRecord{}).Error; err != nil {
			return toErrStatus("model_experience_dialog_delete_err", err.Error())
		}
		return nil
	})
}

func (c *Client) SaveModelExperienceDialogRecord(ctx context.Context, record *model_client.ModelExperienceDialogRecord) *errs.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		if err := tx.Create(record).Error; err != nil {
			return toErrStatus("model_experience_dialog_record_create_err", err.Error())
		}
		// 刷新下对应dialog的updated_at
		if err := sqlopt.WithID(record.ModelExperienceID).
			Apply(tx).Model(&model_client.ModelExperienceDialog{}).Updates(map[string]interface{}{
			"updated_at": time.Now().UnixMilli(),
		}).Error; err != nil {
			return toErrStatus("model_experience_dialog_record_create_err", err.Error())
		}
		return nil
	})
}

func (c *Client) ListModelExperienceDialogRecords(ctx context.Context, userId, orgId string, modelExperienceId uint32, sessionId string) ([]*model_client.ModelExperienceDialogRecord, *errs.Status) {
	var records []*model_client.ModelExperienceDialogRecord
	if err := sqlopt.SQLOptions(
		sqlopt.WithModelExperienceId(modelExperienceId),
		sqlopt.WithSessionID(sessionId),
		sqlopt.WithUserID(userId),
	).Apply(c.db).WithContext(ctx).Find(&records).Error; err != nil {
		return nil, toErrStatus("model_experience_dialog_record_list_err", err.Error())
	}
	return records, nil
}
