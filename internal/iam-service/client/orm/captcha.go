package orm

import (
	"context"
	"strings"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/model"
	"gorm.io/gorm"
)

const (
	expire    = 60 // 验证码有效期 秒 [EN] Verification code validity period seconds
	frequency = 10 // 验证码每分钟可刷新次数（一分钟过后次数重置） [EN] The number of times the verification code can be refreshed per minute (the number is reset after one minute)
)

func (c *Client) RefreshCaptcha(ctx context.Context, key, code string) *errs.Status {
	if key == "" || code == "" {
		return toErrStatus("iam_captcha_refresh")
	}
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		now := time.Now()
		captcha := &model.Captcha{}
		if err := tx.Where("id = ?", key).First(captcha).Error; err != nil {
			// 1. 未创建过 [EN] 1. Never created
			if err == gorm.ErrRecordNotFound {
				if err := tx.Create(&model.Captcha{
					ID:         key,
					Code:       code,
					StartAt:    now.UnixMilli(),
					RefreshAt:  now.UnixMilli(),
					RefreshCnt: 1,
				}).Error; err != nil {
					return toErrStatus("iam_captcha_create", err.Error())
				}
				return nil
			}
			return toErrStatus("iam_captcha_create", err.Error())
		}
		// 2. 创建过 [EN] 2. Created
		// 2.A 距离上次刷新不足1min [EN] 2.A Less than 1 minute has passed since the last refresh
		if now.Sub(time.UnixMilli(captcha.StartAt)) < time.Minute {
			// 2.A.a 未达刷新次数上限 [EN] 2.A.a The upper limit of refresh times has not been reached.
			if captcha.RefreshCnt < int32(frequency) {
				if err := tx.Model(captcha).Updates(map[string]interface{}{
					"code":        code,
					"refresh_at":  now.UnixMilli(),
					"refresh_cnt": captcha.RefreshCnt + 1,
				}).Error; err != nil {
					return toErrStatus("iam_captcha_create", err.Error())
				}
				return nil
			}
			// 2.A.b 超过刷新次数上限 [EN] 2.A.b Exceeds the upper limit of refresh times
			return toErrStatus("iam_captcha_create", "check captcha but key or code empty")
		}
		// 2.B 距离上次刷新超过1min [EN] 2.B More than 1 minute has passed since the last refresh
		if err := tx.Model(captcha).Updates(map[string]interface{}{
			"code":        code,
			"start_at":    now.UnixMilli(),
			"refresh_at":  now.UnixMilli(),
			"refresh_cnt": 1,
		}).Error; err != nil {
			return toErrStatus("iam_captcha_create", err.Error())
		}
		return nil
	})
}

func (c *Client) CheckCaptcha(ctx context.Context, key, code string) *errs.Status {
	if key == "" || code == "" {
		return toErrStatus("iam_captcha_check")
	}
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check exist
		captcha := &model.Captcha{}
		if err := tx.Where("id = ?", key).First(captcha).Error; err != nil {
			return toErrStatus("iam_captcha_check")
		}
		// check expire
		if time.Since(time.UnixMilli(captcha.RefreshAt)) > time.Duration(expire)*time.Second {
			return toErrStatus("iam_captcha_expired")
		}
		// check match
		if !strings.EqualFold(captcha.Code, code) {
			return toErrStatus("iam_captcha_err")
		}
		if err := tx.Unscoped().Delete(captcha).Error; err != nil {
			return toErrStatus("iam_captcha_delete", err.Error())
		}
		return nil
	})
}
