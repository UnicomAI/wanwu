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
	redisLoginEmailField  = "email"
	redisLoginEmailExpire = 5 * time.Minute
)

func (c *Client) Login(ctx context.Context, username, password, language string) (*UserInfo, *Permission, *errs.Status) {
	var userInfo *UserInfo
	var permission *Permission

	return userInfo, permission, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// user
		user := &model.User{}
		if err := sqlopt.WithName(username).Apply(tx).First(user).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return toErrStatus("iam_login_err", err.Error())
			}

		}
		if user.Password != util.SHA256(password) {
			return toErrStatus("iam_login_invalid_name_pwd")
		}
		nowTS := time.Now().UnixMilli()
		var (
			get_err *errs.Status
			update  map[string]interface{}
		)
		userInfo, permission, update, get_err = getUserInfoAndPermission(tx, user, language, nowTS)
		if get_err != nil {
			return get_err
		}
		if err := tx.Model(user).Updates(update).Error; err != nil {
			return toErrStatus("iam_login_err", err.Error())
		}
		return nil
	})
}

func (c *Client) LoginByEmail(ctx context.Context, username, password string) (*EmailLoginInfo, *errs.Status) {
	var emailLoginInfo *EmailLoginInfo
	// user
	user := &model.User{}
	if err := sqlopt.WithName(username).Apply(c.db).First(user).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, toErrStatus("iam_login_err", err.Error())
		}

	}
	if user.Password != util.SHA256(password) {
		return nil, toErrStatus("iam_login_invalid_name_pwd")
	}
	// user info
	emailLoginInfo = &EmailLoginInfo{
		ID:                   user.ID,
		IsEmailCheck:         user.IsEmailCheck,
		LastUpdatePasswordAt: user.LastUpdatePasswordAt,
	}

	return emailLoginInfo, nil
}

func (c *Client) LoginEmailCheck(ctx context.Context, userID uint32, email, code, language string) (*UserInfo, *Permission, *errs.Status) {
	var userInfo *UserInfo
	var permission *Permission

	// check email code

	if err := checkEmailCode(ctx, email, code); err != nil {
		return nil, nil, err
	}

	return userInfo, permission, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// user
		user := &model.User{}

		if err := sqlopt.WithID(userID).Apply(tx).First(user).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return toErrStatus("iam_login_err", err.Error())
			}
		}
		// user not email check but input email binding
		if !user.IsEmailCheck {
			checkBindEmail := &model.User{}
			err := sqlopt.WithEmail(email).Apply(tx).First(&checkBindEmail).Error // find same email
			if err == nil && checkBindEmail.ID != userID {
				return toErrStatus("iam_login_email_bind")
			}
		}

		// have bind email check email
		if user.IsEmailCheck && user.Email != email {
			return toErrStatus("iam_login_email")
		}
		nowTS := time.Now().UnixMilli()
		var (
			get_err *errs.Status
			update  map[string]interface{}
		)
		userInfo, permission, update, get_err = getUserInfoAndPermission(tx, user, language, nowTS)
		if get_err != nil {
			return get_err
		}
		//email binding
		if !user.IsEmailCheck {
			update["email"] = email
			update["is_email_check"] = true
		}
		if err := tx.Model(user).Updates(update).Error; err != nil {
			return toErrStatus("iam_login_err", err.Error())
		}
		return nil
	})
}

func (c *Client) ChangeUserPasswordByEmail(ctx context.Context, userID uint32, oldPassword, newPassword, email, code, language string) (*UserInfo, *Permission, *errs.Status) {
	var userInfo *UserInfo
	var permission *Permission
	if err := checkEmailCode(ctx, email, code); err != nil {
		return nil, nil, err
	}
	return userInfo, permission, c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		// user
		user := &model.User{}
		if err := sqlopt.WithID(userID).Apply(tx).First(user).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return toErrStatus("iam_login_err", err.Error())
			}
		}
		// user not email check but input email binding
		if !user.IsEmailCheck {
			checkBindEmail := &model.User{}
			err := sqlopt.WithEmail(email).Apply(tx).First(&checkBindEmail).Error // find same email
			if err == nil && checkBindEmail.ID != userID {
				return toErrStatus("iam_login_email_bind")
			}
		}
		// have bind email check email
		if user.IsEmailCheck && user.Email != email {
			return toErrStatus("iam_login_email")
		}
		// check password
		if user.Password != util.SHA256(oldPassword) {
			return toErrStatus("iam_user_password_update", util.Int2Str(userID), "password error")
		}
		nowTS := time.Now().UnixMilli()
		var (
			get_err *errs.Status
			update  map[string]interface{}
		)
		userInfo, permission, update, get_err = getUserInfoAndPermission(tx, user, language, nowTS)
		if get_err != nil {
			return get_err
		}

		update["password"] = util.SHA256(newPassword)
		update["last_update_password_at"] = nowTS
		permission.LastUpdatePasswordAt = nowTS

		//email check
		if !user.IsEmailCheck {
			update["email"] = email
			update["is_email_check"] = true
		}
		if err := tx.Model(user).Updates(update).Error; err != nil {
			return toErrStatus("iam_login_err", err.Error())
		}
		return nil
	})
}

func (c *Client) LoginSendEmailCode(ctx context.Context, email string) *errs.Status {
	if email == "" {
		return toErrStatus("iam_user_email_code", "email empty")
	}

	now := time.Now()
	nowTs := now.UnixMilli()
	code := iam_util.RandText(config.Cfg().Password.Email.CodeLength)

	item, err := redis.IAM().HGet(ctx, getRedisLoginEmailKey(email), redisLoginEmailField)
	if err != nil {
		return toErrStatus("iam_user_email_code", err.Error())
	} else if item != nil {
		// Verification code sent via email
		record, err := getRedisLoginEmailRecord(item.V)
		if err != nil {
			return toErrStatus("iam_user_email_code", err.Error())
		}
		// Less than 1 minute since last sent
		if now.Sub(time.UnixMilli(record.Timestamp)) < time.Minute {
			return toErrStatus("iam_register_by_email_send_code_frequent")
		}
	}
	// The verification code has not been sent by email or it has been more than 1 minute since it was last sent.
	// Send email
	if err := smtp_util.SendEmail([]string{email},
		config.Cfg().Password.Email.Template.Subject,
		config.Cfg().Password.Email.Template.ContentType,
		fmt.Sprintf(config.Cfg().Password.Email.Template.Body, code),
	); err != nil {
		return toErrStatus("iam_register_by_email_send_code", err.Error())
	}
	// Record redis
	if err := redis.IAM().HSet(ctx, getRedisLoginEmailKey(email), []redis.HashItem{
		{
			K: redisLoginEmailField,
			V: getRedisLoginEmailItemValue(redisLoginEmail{
				Code:      code,
				Timestamp: nowTs,
			}),
		},
	}); err != nil {
		return toErrStatus("iam_user_email_code", err.Error())
	}
	if err := redis.IAM().Expire(ctx, getRedisLoginEmailKey(email), redisLoginEmailExpire); err != nil {
		return toErrStatus("iam_user_email_code", err.Error())
	}
	return nil
}

// --- internal ---

type redisLoginEmail struct {
	Code      string `json:"code"`
	Timestamp int64  `json:"timestamp"` //  The creation time of the current verification code
}

func getRedisLoginEmailKey(email string) string {
	return fmt.Sprintf("user-login-email-code:%v", email)
}

func getRedisLoginEmailItemValue(record redisLoginEmail) string {
	b, _ := json.Marshal(record)
	return string(b)
}

func getRedisLoginEmailRecord(value string) (*redisLoginEmail, error) {
	var ret *redisLoginEmail
	if err := json.Unmarshal([]byte(value), &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func checkEmailCode(ctx context.Context, email, code string) *errs.Status {
	// check email code
	item, err := redis.IAM().HGet(ctx, getRedisLoginEmailKey(email), redisLoginEmailField)
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
	return nil
}

func getUserInfoAndPermission(tx *gorm.DB, user *model.User, language string, nowTS int64) (*UserInfo, *Permission, map[string]interface{}, *errs.Status) {
	var permission *Permission
	update := make(map[string]interface{})
	// check status
	if !user.Status {
		return nil, nil, nil, toErrStatus("iam_login_user_disable")
	}
	// org tree
	orgTree, err := getOrgTree(tx)
	if err != nil {
		return nil, nil, nil, toErrStatus("iam_login_err", err.Error())
	}
	// user info
	userInfo, err := toUserInfoTx(tx, user, orgTree, true)
	if err != nil {
		return nil, nil, nil, toErrStatus("iam_login_err", err.Error())
	}
	var orgID uint32
	if len(userInfo.Orgs) > 0 {
		orgID = userInfo.Orgs[0].Org.ID
	}
	if len(userInfo.Orgs) > 1 {
		for _, org := range userInfo.Orgs {
			if !org.Org.Status {
				continue
			}
			if len(org.Roles) > 0 {
				orgID = org.Org.ID
				break
			}
		}
	}
	if orgID != 0 {
		permission, err = getUserPermission(tx, user.ID, orgID)
	}
	if err != nil {
		return nil, nil, nil, toErrStatus("iam_login_err", err.Error())
	}
	// last_login_at & last_exec_at
	update["last_login_at"] = nowTS
	update["last_exec_at"] = nowTS
	if language != "" {
		update["language"] = language
		userInfo.Language = language
	}
	return userInfo, permission, update, nil
}
