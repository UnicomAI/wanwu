package orm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/model"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/internal/iam-service/config"
	iam_util "github.com/UnicomAI/wanwu/internal/iam-service/pkg/util"
	smtp_util "github.com/UnicomAI/wanwu/internal/iam-service/pkg/util/smtp-util"
	"github.com/UnicomAI/wanwu/pkg/redis"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

const (
	redisUserResetPwdByEmailField  = "email"
	redisUserResetPwdByEmailExpire = 5 * time.Minute
)

func (c *Client) ResetPasswordSendEmailCode(ctx context.Context, email string) *errs.Status {
	if email == "" {
		return toErrStatus("iam_user_email_code", "email empty")
	}

	now := time.Now()
	nowTs := now.UnixMilli()
	code := iam_util.RandText(config.Cfg().Password.Email.CodeLength)

	item, err := redis.IAM().HGet(ctx, getRedisUserResetPwdByEmailKey(email), redisUserResetPwdByEmailField)
	if err != nil {
		return toErrStatus("iam_user_email_code", err.Error())
	} else if item != nil {
		// email发送过验证码 [EN] Verification code sent via email
		record, err := getRedisUserResetPwdByEmailRecord(item.V)
		if err != nil {
			return toErrStatus("iam_user_email_code", err.Error())
		}
		// 距离上次发送不足1min [EN] Less than 1 minute since last sent
		if now.Sub(time.UnixMilli(record.Timestamp)) < time.Minute {
			return toErrStatus("iam_register_by_email_send_code_frequent")
		}
	}
	// email未发送过验证码 或者 距离上次发送超过1min [EN] The verification code has not been sent by email or it has been more than 1 minute since it was last sent.
	// 发送邮件 [EN] Send email
	if err := smtp_util.SendEmail([]string{email},
		config.Cfg().Password.Email.Template.Subject,
		config.Cfg().Password.Email.Template.ContentType,
		fmt.Sprintf(config.Cfg().Password.Email.Template.Body, code),
	); err != nil {
		return toErrStatus("iam_register_by_email_send_code", err.Error())
	}
	// 记录redis [EN] Record redis
	if err := redis.IAM().HSet(ctx, getRedisUserResetPwdByEmailKey(email), []redis.HashItem{
		{
			K: redisUserResetPwdByEmailField,
			V: getRedisUserResetPwdByEmailItemValue(redisUserResetPasswordByEmail{
				Code:      code,
				Timestamp: nowTs,
			}),
		},
	}); err != nil {
		return toErrStatus("iam_user_email_code", err.Error())
	}
	if err := redis.IAM().Expire(ctx, getRedisUserResetPwdByEmailKey(email), redisUserResetPwdByEmailExpire); err != nil {
		return toErrStatus("iam_user_email_code", err.Error())
	}
	return nil
}

func (c *Client) ResetPasswordByEmail(ctx context.Context, email, password, code string) *errs.Status {
	// check email code
	item, err := redis.IAM().HGet(ctx, getRedisUserResetPwdByEmailKey(email), redisUserResetPwdByEmailField)
	if err != nil {
		return toErrStatus("iam_user_password_reset", email, err.Error())
	} else if item == nil {
		return toErrStatus("iam_register_by_email_not_found")
	}
	record, err := getRedisUserResetPwdByEmailRecord(item.V)
	if err != nil {
		return toErrStatus("iam_user_password_reset", email, err.Error())
	}
	if !strings.EqualFold(record.Code, code) {
		return toErrStatus("iam_register_by_email_invalid_code")
	}
	// update password
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// check email
		user := &model.User{}
		if err := sqlopt.WithEmail(email).Apply(tx).First(user).Error; err != nil {
			return toErrStatus("iam_user_password_reset", email, err.Error())
		}
		// reset password
		if err := tx.Model(user).Updates(map[string]interface{}{
			"password": util.SHA256(password),
		}).Error; err != nil {
			return toErrStatus("iam_user_password_reset", email, err.Error())
		}
		// redis del
		if err := redis.IAM().Del(ctx, getRedisUserResetPwdByEmailKey(email)); err != nil {
			return toErrStatus("iam_user_password_reset", email, err.Error())
		}
		return nil
	})
}

// --- internal ---

type redisUserResetPasswordByEmail struct {
	Code      string `json:"code"`
	Timestamp int64  `json:"timestamp"` //  当前验证码的创建时间 [EN] The creation time of the current verification code
}

func getRedisUserResetPwdByEmailKey(email string) string {
	return fmt.Sprintf("user-reset-password-by-email:%v", email)
}

func getRedisUserResetPwdByEmailItemValue(record redisUserResetPasswordByEmail) string {
	b, _ := json.Marshal(record)
	return string(b)
}

func getRedisUserResetPwdByEmailRecord(value string) (*redisUserResetPasswordByEmail, error) {
	var ret *redisUserResetPasswordByEmail
	if err := json.Unmarshal([]byte(value), &ret); err != nil {
		return nil, err
	}
	return ret, nil
}
