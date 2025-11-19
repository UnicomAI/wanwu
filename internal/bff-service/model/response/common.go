package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type Response struct {
	Code int64       `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

type PageResult struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	PageNo   int         `json:"pageNo"`
	PageSize int         `json:"pageSize"`
}

type ListResult struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
}

type IDName struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type UserPermission struct {
	OrgPermission    UserOrgPermission `json:"orgPermission"`    // 用户所在组织权限 [EN] User organization permissions
	Language         Language          `json:"language"`         // 语言 [EN] language
	IsUpdatePassword bool              `json:"isUpdatePassword"` // 是否已更新密码 [EN] Has the password been updated?
	Avatar           request.Avatar    `json:"avatar"`           // 用户头像信息 [EN] User avatar information
}

type UserOrgPermission struct {
	IsAdmin     bool         `json:"isAdmin"`     // 是否系统内置管理员 [EN] Whether the system has a built-in administrator
	IsSystem    bool         `json:"isSystem"`    // 是否系统视角（此时org.id为空，org.name为"系统"） [EN] Whether to use the system perspective (at this time, org.id is empty and org.name is "system")
	Org         IDName       `json:"org"`         // 组织 [EN] organize
	Roles       []IDName     `json:"roles"`       // 角色列表 [EN] role list
	Permissions []Permission `json:"permissions"` // 权限列表 [EN] Permission list
}

type Permission struct {
	Perm string `json:"perm"` // 权限 [EN] Permissions
	Name string `json:"name"` // 权限名（对应菜单名） [EN] Permission name (corresponding to menu name)
}

type Select struct {
	Select []IDName `json:"select"`
}

type DocMenu struct {
	Name     string     `json:"name"`     // 目录名称 [EN] directory name
	Index    string     `json:"index"`    // 目录索引 [EN] directory index
	Path     string     `json:"path"`     // 目录路径（转码后） [EN] Directory path (after transcoding)
	PathRaw  string     `json:"pathRaw"`  // 目录路径 [EN] directory path
	Children []*DocMenu `json:"children"` // 目录 [EN] Table of contents

	content string
}

func (dm *DocMenu) SetContent(content string) {
	dm.content = content
}

type DocSearchResp struct {
	Title       string             `json:"title"` // 文档名 [EN] Document name
	ContentList []DocSearchContent `json:"list"`  // 内容列表 [EN] Contents list
}

type DocSearchContent struct {
	Title   string `json:"title"`   // 文档中的子标题 [EN] Subtitles in the document
	Content string `json:"content"` // 内容 [EN] content
	Url     string `json:"url"`     // 文档链接 [EN] Documentation link
}
