package orm

import (
	"context"
	"errors"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/model"
	"github.com/gromitlee/access"
	"github.com/gromitlee/depend/v2"
	"gorm.io/gorm"
)

type Client struct {
	db *gorm.DB
}

func NewClient(db *gorm.DB) (*Client, error) {
	// rbac
	if err := access.InitAccessRBAC0Controller(db); err != nil {
		return nil, err
	}
	// depend
	if err := depend.Init(db); err != nil {
		return nil, err
	}
	// auto migrate
	if err := db.AutoMigrate(
		model.User{},
		model.UserRole{},
		model.Org{},
		model.OrgUser{},
		model.OrgRole{},
		model.Captcha{},
		model.OauthApp{},
	); err != nil {
		return nil, err
	}
	return &Client{
		db: db,
	}, nil
}

type IDName struct {
	ID         uint32
	Name       string
	NameStatus *err_code.Status
}

type OrgUserIDName struct {
	IDName
	Status bool
}

type RoleIDName struct {
	ID       uint32
	Name     string
	IsAdmin  bool
	IsSystem bool
}

type OrgInfo struct {
	ID        uint32
	Name      string
	Remark    string
	Status    bool
	CreatedAt int64
	Creator   IDName
}

type RoleInfo struct {
	ID        uint32
	IsAdmin   bool
	IsSystem  bool
	Name      string
	Remark    string
	Status    bool
	CreatedAt int64
	Creator   IDName
	Perms     []Perm
}

type UserInfo struct {
	ID         uint32
	Status     bool
	Name       string
	Nick       string
	Gender     string
	Phone      string
	Email      string
	Company    string
	Remark     string
	CreatedAt  int64
	Creator    IDName
	Orgs       []*UserOrg
	Language   string
	AvatarPath string
}

type UserOrg struct {
	Org   OrgUserIDName
	Roles []RoleIDName
}

type Permission struct {
	IsAdmin              bool // 是否是当前组织的内置管理角色 [EN] Is it a built-in management role for the current organization?
	IsSystem             bool
	Org                  IDName
	Roles                []RoleIDName
	Perms                []Perm
	LastUpdatePasswordAt int64
}

type Perm struct {
	Perm string
}

type EmailLoginInfo struct {
	ID                   uint32
	IsEmailCheck         bool
	LastUpdatePasswordAt int64
}

func toErrStatus(key string, args ...string) *err_code.Status {
	return &err_code.Status{
		TextKey: key,
		Args:    args,
	}
}

func (c *Client) transaction(ctx context.Context, fc func(tx *gorm.DB) *err_code.Status) *err_code.Status {
	var status *err_code.Status
	_ = c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if status = fc(tx); status != nil {
			return errors.New(status.String())
		}
		return nil
	})
	return status
}
