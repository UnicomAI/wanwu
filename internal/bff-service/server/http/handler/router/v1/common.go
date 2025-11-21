package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerCommon(apiV1 *gin.RouterGroup) {
	mid.Sub("common").Reg(apiV1, "/user/permission", http.MethodGet, v1.GetUserPermission, "获取用户权限")
	mid.Sub("common").Reg(apiV1, "/user/info", http.MethodGet, v1.GetUserInfo, "获取用户信息")
	mid.Sub("common").Reg(apiV1, "/org/select", http.MethodGet, v1.GetOrgSelect, "获取用户组织列表")
	mid.Sub("common").Reg(apiV1, "/user/password", http.MethodPut, v1.ChangeUserPassword, "修改用户密码（by 个人）")
	mid.Sub("common").Reg(apiV1, "/avatar", http.MethodPost, v1.UploadAvatar, "上传自定义图标")
	mid.Sub("common").Reg(apiV1, "/user/avatar", http.MethodPut, v1.UpdateUserAvatar, "编辑用户头像")

	// Second stage user login
	mid.Sub("common").Reg(apiV1, "/user/login", http.MethodPost, v1.LoginEmailCheck, "非首次登录邮箱绑定校验")
	mid.Sub("common").Reg(apiV1, "/user/login", http.MethodPut, v1.ChangeUserPasswordByEmail, "首次登录修改用户密码与邮箱绑定校验（by 个人）")
	mid.Sub("common").Reg(apiV1, "/user/login/email/code", http.MethodPost, v1.LoginSendEmailCode, "登录邮箱验证码发送")

	// Universal file upload
	mid.Sub("common").Reg(apiV1, "/file/check", http.MethodGet, v1.CheckFile, "校验文件")
	mid.Sub("common").Reg(apiV1, "/file/check/list", http.MethodGet, v1.CheckFileList, "校验文件列表")
	mid.Sub("common").Reg(apiV1, "/file/upload", http.MethodPost, v1.UploadFile, "上传文件")
	mid.Sub("common").Reg(apiV1, "/file/merge", http.MethodPost, v1.MergeFile, "合并文件")
	mid.Sub("common").Reg(apiV1, "/file/clean", http.MethodPost, v1.CleanFile, "清除文件")
	mid.Sub("common").Reg(apiV1, "/file/delete", http.MethodDelete, v1.DeleteFile, "刪除文件")
	mid.Sub("common").Reg(apiV1, "/proxy/file/upload", http.MethodPost, v1.ProxyUploadFile, "代理上传文件")

	// Document Center
	mid.Sub("common").Reg(apiV1, "/doc_center/search", http.MethodGet, v1.SearchDocCenter, "查找文档中心内容")
	mid.Sub("common").Reg(apiV1, "/doc_center/menu", http.MethodGet, v1.GetDocCenterMenu, "获取文档中心目录")
	mid.Sub("common").Reg(apiV1, "/doc_center/markdown", http.MethodGet, v1.GetDocCenterMarkdown, "获取文档中心Markdown文件内容")

	// Model universal
	mid.Sub("common").Reg(apiV1, "/model/select/llm", http.MethodGet, v1.ListLlmModels, "llm模型列表展示")
	mid.Sub("common").Reg(apiV1, "/model/select/rerank", http.MethodGet, v1.ListRerankModels, "rerank模型列表展示")
	mid.Sub("common").Reg(apiV1, "/model/select/embedding", http.MethodGet, v1.ListEmbeddingModels, "embedding模型列表展示")
	mid.Sub("common").Reg(apiV1, "/model/select/ocr", http.MethodGet, v1.ListOcrModels, "ocr模型列表展示")
	mid.Sub("common").Reg(apiV1, "/model/select/pdf-parser", http.MethodGet, v1.ListPdfParserModels, "pdf文档解析模型列表展示")
	mid.Sub("common").Reg(apiV1, "/model/select/gui", http.MethodGet, v1.ListGuiModels, "gui模型列表展示")

	// Knowledge base common
	mid.Sub("common").Reg(apiV1, "/knowledge/select", http.MethodPost, v1.GetKnowledgeSelect, "查询用户知识库列表")

	// rag/agent/workflow general
	mid.Sub("common").Reg(apiV1, "/appspace/app", http.MethodDelete, v1.DeleteAppSapceApp, "刪除应用")
	mid.Sub("common").Reg(apiV1, "/appspace/app/list", http.MethodGet, v1.GetAppSpaceAppList, "获取应用列表")
	mid.Sub("common").Reg(apiV1, "/appspace/app/publish", http.MethodPost, v1.PublishApp, "发布应用")
	mid.Sub("common").Reg(apiV1, "/appspace/app/publish", http.MethodDelete, v1.UnPublishApp, "取消发布应用")
	mid.Sub("common").Reg(apiV1, "/appspace/app/url", http.MethodGet, v1.GetApiBaseUrl, "获取Api根地址")
	mid.Sub("common").Reg(apiV1, "/appspace/app/key", http.MethodPost, v1.GenApiKey, "生成ApiKey")
	mid.Sub("common").Reg(apiV1, "/appspace/app/key", http.MethodDelete, v1.DelApiKey, "删除ApiKey")
	mid.Sub("common").Reg(apiV1, "/appspace/app/key/list", http.MethodGet, v1.GetApiKeyList, "获取ApiKey列表")

	// SafetyUniversal
	mid.Sub("common").Reg(apiV1, "/safe/sensitive/table/select", http.MethodGet, v1.GetSensitiveWordTableSelect, "获取敏感词表下拉列表")

	// ApplicationUrl
	mid.Sub("common").Reg(apiV1, "/appspace/app/openurl", http.MethodPost, v1.AppUrlCreate, "创建应用Url")
	mid.Sub("common").Reg(apiV1, "/appspace/app/openurl", http.MethodDelete, v1.AppUrlDelete, "删除应用Url")
	mid.Sub("common").Reg(apiV1, "/appspace/app/openurl", http.MethodPut, v1.AppUrlUpdate, "编辑应用Url")
	mid.Sub("common").Reg(apiV1, "/appspace/app/openurl/list", http.MethodGet, v1.GetAppUrlList, "获取应用Url列表")
	mid.Sub("common").Reg(apiV1, "/appspace/app/openurl/status", http.MethodPut, v1.AppUrlStatusSwitch, "启用/停用应用Url")

	// Oauth
	mid.Sub("common").Reg(apiV1, "/oauth/code/authorize", http.MethodGet, v1.OAuthAuthorize, "获取授权码")
}
