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
		model.GlobalRole{},
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

type IDNameWithAvatar struct {
	ID         uint32
	Name       string
	AvatarPath string
	NameStatus *err_code.Status
}

type IDFullName struct {
	IDName
	FullName string
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
	IsGlobal bool
}

type OrgInfo struct {
	ID         uint32
	Name       string
	Remark     string
	Status     bool
	CreatedAt  int64
	Creator    IDName
	UserCount  int64
	Admins     []string
	AvatarPath string
}

type RoleInfo struct {
	ID         uint32
	IsAdmin    bool
	IsSystem   bool
	IsGlobal   bool
	Name       string
	Remark     string
	Status     bool
	CreatedAt  int64
	Creator    IDName
	OrgName    string
	UserCount  int32
	AvatarPath string
	Perms      []Perm
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
	IsAdmin              bool // 是否是当前组织的内置管理角色
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

type UsersInfo struct {
	UserName string
	Phone    string
	Email    string
	Company  string
	Remark   string
	Password string
	RoleName string
}

// CreateUserError 单条用户创建错误
type CreateUserError struct {
	Index  int // 数据索引（在请求users数组中的索引）
	Reason string
}

// CreateUsersResult 批量创建用户结果
type CreateUsersResult struct {
	Total   int
	Success int
	Failed  int
	Errors  []CreateUserError
}

// RoleUser 角色关联用户信息
type RoleUser struct {
	UserID     uint32
	UserName   string
	Phone      string
	Email      string
	AvatarPath string
	Orgs       []IDName
}

// AdminOrgTreeNode 管理员组织树节点
type AdminOrgTreeNode struct {
	ID       uint32
	Name     string
	HasPerm  bool
	Children []*AdminOrgTreeNode
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
