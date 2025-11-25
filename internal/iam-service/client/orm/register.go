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
	"gorm.io/gorm"
)

const (
	redisUserRegisterByEmailField  = "email"
	redisUserRegisterByEmailExpire = 5 * time.Minute
)

func (c *Client) RegisterSendEmailCode(ctx context.Context, username, email string) *errs.Status {
	if email == "" {
		return toErrStatus("iam_register_by_email_send_code", "email empty")
	}
	if username == "" {
		return toErrStatus("iam_register_by_email_send_code", "username empty")
	}
	// check user
	if err := sqlopt.WithName(username).Apply(c.db).WithContext(ctx).First(&model.User{}).Error; err != gorm.ErrRecordNotFound {
		if err == nil {
			return toErrStatus("iam_user_create_name") // Username already exists
		}
		return toErrStatus("iam_register_by_email_send_code", err.Error())
	}

	// check email
	if err := sqlopt.WithEmail(email).Apply(c.db).WithContext(ctx).First(&model.User{}).Error; err != gorm.ErrRecordNotFound {
		if err == nil {
			return toErrStatus("iam_user_create_email")
		}
		return toErrStatus("iam_register_by_email_send_code", err.Error())
	}

	now := time.Now()
	nowTs := now.UnixMilli()
	code := iam_util.RandText(config.Cfg().Register.Email.CodeLength)
	password := iam_util.RandText(config.Cfg().Register.Email.PasswordLength)

	item, err := redis.IAM().HGet(ctx, getRedisUserRegisterByEmailKey(email), redisUserRegisterByEmailField)
	if err != nil {
		return toErrStatus("iam_register_by_email_send_code", err.Error())
	} else if item != nil {
		// Verification code sent via email
		record, err := getRedisUserRegisterByEmailRecord(item.V)
		if err != nil {
			return toErrStatus("iam_register_by_email_send_code", err.Error())
		}
		// Less than 1 minute since last sent
		if now.Sub(time.UnixMilli(record.Timestamp)) < time.Minute {
			return toErrStatus("iam_register_by_email_send_code_frequent")
		}
	}
	// The verification code has not been sent by email or it has been more than 1 minute since it was last sent.
	// Send email
	if err := smtp_util.SendEmail([]string{email},
		config.Cfg().Register.Email.Template.Subject,
		config.Cfg().Register.Email.Template.ContentType,
		fmt.Sprintf(config.Cfg().Register.Email.Template.Body, code, password),
	); err != nil {
		return toErrStatus("iam_register_by_email_send_code", err.Error())
	}
	// Record redis
	if err := redis.IAM().HSet(ctx, getRedisUserRegisterByEmailKey(email), []redis.HashItem{
		{
			K: redisUserRegisterByEmailField,
			V: getRedisUserRegisterByEmailItemValue(redisUserRegisterByEmail{
				Code:      code,
				Password:  password,
				Timestamp: nowTs,
			}),
		},
	}); err != nil {
		return toErrStatus("iam_register_by_email_send_code", err.Error())
	}
	if err := redis.IAM().Expire(ctx, getRedisUserRegisterByEmailKey(email), redisUserRegisterByEmailExpire); err != nil {
		return toErrStatus("iam_register_by_email_send_code", err.Error())
	}
	return nil
}

func (c *Client) RegisterByEmail(ctx context.Context, username, email, code string) *errs.Status {
	// check email code
	item, err := redis.IAM().HGet(ctx, getRedisUserRegisterByEmailKey(email), redisUserRegisterByEmailField)
	if err != nil {
		return toErrStatus("iam_register_by_email", err.Error())
	} else if item == nil {
		return toErrStatus("iam_register_by_email_not_found")
	}
	record, err := getRedisUserRegisterByEmailRecord(item.V)
	if err != nil {
		return toErrStatus("iam_register_by_email", err.Error())
	}
	if !strings.EqualFold(record.Code, code) {
		return toErrStatus("iam_register_by_email_invalid_code")
	}

	var userID uint32
	var newOrgID uint32

	// register
	if err := c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// create user
		user := &model.User{
			Status:       true,
			CreatorID:    config.AdminUserID(),
			Name:         username,
			Nick:         username,
			Email:        email,
			Password:     record.Password,
			IsEmailCheck: true,
		}
		if err := createUserTx(tx, user, config.TopOrgID(), nil); err != nil {
			return err
		}
		userID = user.ID

		// create org
		code := iam_util.RandText(config.Cfg().Register.Email.CodeLength)
		newOrg := &model.Org{
			Status:    true,
			CreatorID: user.ID,
			ParentID:  config.TopOrgID(),
			Name:      fmt.Sprintf("%v-Space-%v", username, code),
		}
		if err := createOrgTx(tx, newOrg); err != nil {
			return err
		}
		newOrgID = newOrg.ID

		// redis del
		if err := redis.IAM().Del(tx.Statement.Context, getRedisUserRegisterByEmailKey(email)); err != nil {
			return toErrStatus("iam_register_by_email", err.Error())
		}
		return nil
	}); err != nil {
		return err
	}

	// Sync user to workflow space_user table for both TopOrgID and new org
	// User is owner (role_type=1) of their own space
	if err := c.SyncUserToWorkflowSpace(ctx, userID, config.TopOrgID(), 3); err != nil {
		// Log but don't fail registration
	}
	if err := c.SyncUserToWorkflowSpace(ctx, userID, newOrgID, 1); err != nil {
		// Log but don't fail registration
	}

	return nil
}

// --- internal ---

type redisUserRegisterByEmail struct {
	Code      string `json:"code"`
	Password  string `json:"password"`  // default password
	Timestamp int64  `json:"timestamp"` //  The creation time of the current verification code
}

func getRedisUserRegisterByEmailKey(email string) string {
	return fmt.Sprintf("user-register-by-email:%v", email)
}

func getRedisUserRegisterByEmailItemValue(record redisUserRegisterByEmail) string {
	b, _ := json.Marshal(record)
	return string(b)
}

func getRedisUserRegisterByEmailRecord(value string) (*redisUserRegisterByEmail, error) {
	var ret *redisUserRegisterByEmail
	if err := json.Unmarshal([]byte(value), &ret); err != nil {
		return nil, err
	}
	return ret, nil
}
