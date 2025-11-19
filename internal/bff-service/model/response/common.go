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
	OrgPermission    UserOrgPermission `json:"orgPermission"`    // User organization permissions
	Language         Language          `json:"language"`         // language
	IsUpdatePassword bool              `json:"isUpdatePassword"` // Has the password been updated?
	Avatar           request.Avatar    `json:"avatar"`           // User avatar information
}

type UserOrgPermission struct {
	IsAdmin     bool         `json:"isAdmin"`     // Whether the system has a built-in administrator
	IsSystem    bool         `json:"isSystem"`    // Whether to use the system perspective (at this time, org.id is empty and org.name is "system")
	Org         IDName       `json:"org"`         // organize
	Roles       []IDName     `json:"roles"`       // role list
	Permissions []Permission `json:"permissions"` // Permission list
}

type Permission struct {
	Perm string `json:"perm"` // Permissions
	Name string `json:"name"` // Permission name (corresponding to menu name)
}

type Select struct {
	Select []IDName `json:"select"`
}

type DocMenu struct {
	Name     string     `json:"name"`     // directory name
	Index    string     `json:"index"`    // directory index
	Path     string     `json:"path"`     // Directory path (after transcoding)
	PathRaw  string     `json:"pathRaw"`  // directory path
	Children []*DocMenu `json:"children"` // Table of contents

	content string
}

func (dm *DocMenu) SetContent(content string) {
	dm.content = content
}

type DocSearchResp struct {
	Title       string             `json:"title"` // Document name
	ContentList []DocSearchContent `json:"list"`  // Contents list
}

type DocSearchContent struct {
	Title   string `json:"title"`   // Subtitles in the document
	Content string `json:"content"` // content
	Url     string `json:"url"`     // Documentation link
}
