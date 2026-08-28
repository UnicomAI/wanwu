package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/middleware"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerCommon(apiV1 *gin.RouterGroup) {
	// 需鉴权的静态文件服务（JWTUser + CheckUserEnable）
	filesGroup := apiV1.Group("/files")
	filesGroup.Use(middleware.JWTUser, middleware.CheckUserEnable)
	filesGroup.Static("/", "./configs/microservice/bff-service/configs/files")

	mid.Sub("common").Reg(apiV1, "/user/permission", http.MethodGet, v1.GetUserPermission, "获取用户权限")
	mid.Sub("common").Reg(apiV1, "/user/info", http.MethodGet, v1.GetUserInfo, "获取用户信息")
	mid.Sub("common").Reg(apiV1, "/org/select", http.MethodGet, v1.GetOrgSelect, "获取用户组织列表")
	mid.Sub("common").Reg(apiV1, "/user/password", http.MethodPut, v1.ChangeUserPassword, "修改用户密码（by 个人）")
	mid.Sub("common").Reg(apiV1, "/avatar", http.MethodPost, v1.UploadAvatar, "上传自定义图标")
	mid.Sub("common").Reg(apiV1, "/user/avatar", http.MethodPut, v1.UpdateUserAvatar, "编辑用户头像")

	// 二阶段用户登录
	mid.Sub("common").Reg(apiV1, "/user/login", http.MethodPost, v1.LoginEmailCheck, "非首次登录邮箱绑定校验")
	mid.Sub("common").Reg(apiV1, "/user/login", http.MethodPut, v1.ChangeUserPasswordByEmail, "首次登录修改用户密码与邮箱绑定校验（by 个人）")
	mid.Sub("common").Reg(apiV1, "/user/login/email/code", http.MethodPost, v1.LoginSendEmailCode, "登录邮箱验证码发送")

	// 通用文件上传
	mid.Sub("common").Reg(apiV1, "/file/check", http.MethodGet, v1.CheckFile, "校验文件")
	mid.Sub("common").Reg(apiV1, "/file/check/list", http.MethodGet, v1.CheckFileList, "校验文件列表")
	mid.Sub("common").Reg(apiV1, "/file/upload", http.MethodPost, v1.UploadFile, "上传文件")
	mid.Sub("common").Reg(apiV1, "/file/upload/direct", http.MethodPost, v1.DirectUploadFiles, "直接上传文件(前端用)")
	mid.Sub("common").Reg(apiV1, "/file/merge", http.MethodPost, v1.MergeFile, "合并文件")
	mid.Sub("common").Reg(apiV1, "/file/clean", http.MethodPost, v1.CleanFile, "清除文件")
	mid.Sub("common").Reg(apiV1, "/file/delete", http.MethodDelete, v1.DeleteFile, "刪除文件")

	// 文档中心
	mid.Sub("common").Reg(apiV1, "/doc_center/search", http.MethodGet, v1.SearchDocCenter, "查找文档中心内容")
	mid.Sub("common").Reg(apiV1, "/doc_center/menu", http.MethodGet, v1.GetDocCenterMenu, "获取文档中心目录")
	mid.Sub("common").Reg(apiV1, "/doc_center/markdown", http.MethodGet, v1.GetDocCenterMarkdown, "获取文档中心Markdown文件内容")

	// 模型通用
	mid.Sub("common").Reg(apiV1, "/model/select/llm", http.MethodGet, v1.ListLlmModels, "llm模型列表展示")
	mid.Sub("common").Reg(apiV1, "/model/select/rerank", http.MethodGet, v1.ListRerankModels, "rerank模型列表展示")
	mid.Sub("common").Reg(apiV1, "/model/select/multi-rerank", http.MethodGet, v1.ListMultiRerankModels, "多模态rerank模型列表展示")
	mid.Sub("common").Reg(apiV1, "/model/select/embedding", http.MethodGet, v1.ListEmbeddingModels, "embedding模型列表展示")
	mid.Sub("common").Reg(apiV1, "/model/select/multi-embedding", http.MethodGet, v1.ListMultiEmbeddingModels, "多模态embedding模型列表展示")
	mid.Sub("common").Reg(apiV1, "/model/select/ocr", http.MethodGet, v1.ListOcrModels, "ocr模型列表展示")
	mid.Sub("common").Reg(apiV1, "/model/select/pdf-parser", http.MethodGet, v1.ListPdfParserModels, "pdf文档解析模型列表展示")
	mid.Sub("common").Reg(apiV1, "/model/select/gui", http.MethodGet, v1.ListGuiModels, "gui模型列表展示")
	mid.Sub("common").Reg(apiV1, "/model/select/sync-asr", http.MethodGet, v1.ListSyncAsrModels, "同步asr语音识别模型列表展示")
	//mid.Sub("common").Reg(apiV1, "/model/select/text2image", http.MethodGet, v1.ListText2ImageModels, "文生图模型列表展示")

	// 知识库通用
	mid.Sub("common").Reg(apiV1, "/knowledge/select", http.MethodPost, v1.GetKnowledgeSelect, "查询用户知识库列表")

	// rag/agent/workflow通用
	mid.Sub("common").Reg(apiV1, "/appspace/app", http.MethodDelete, v1.DeleteAppSapceApp, "刪除应用")
	mid.Sub("common").Reg(apiV1, "/appspace/app/publish", http.MethodPost, v1.PublishApp, "发布应用")
	mid.Sub("common").Reg(apiV1, "/appspace/app/publish", http.MethodDelete, v1.UnPublishApp, "取消发布应用")

	// 老版本AppKey接口暂时保留（mcp server在使用）
	mid.Sub("common").Reg(apiV1, "/appspace/app/url", http.MethodGet, v1.GetAppBaseUrl, "获取App根地址")
	mid.Sub("common").Reg(apiV1, "/appspace/app/key", http.MethodPost, v1.GenAppKey, "生成AppKey")
	mid.Sub("common").Reg(apiV1, "/appspace/app/key", http.MethodDelete, v1.DelAppKey, "删除AppKey")
	mid.Sub("common").Reg(apiV1, "/appspace/app/key/list", http.MethodGet, v1.GetAppKeyList, "获取AppKey列表")

	mid.Sub("common").Reg(apiV1, "/appspace/app/version/list", http.MethodGet, v1.GetAppVersionList, "获取应用版本列表")
	mid.Sub("common").Reg(apiV1, "/appspace/app/version", http.MethodPut, v1.UpdateAppVersion, "更新应用版本描述并设置公开范围")
	mid.Sub("common").Reg(apiV1, "/appspace/app/version/rollback", http.MethodPost, v1.RollbackAppVersion, "回滚应用到指定版本")
	mid.Sub("common").Reg(apiV1, "/appspace/app/version", http.MethodGet, v1.GetAppLatestVersion, "获取应用最新版本信息")

	// Safety通用
	mid.Sub("common").Reg(apiV1, "/safe/sensitive/table/select", http.MethodGet, v1.GetSensitiveWordTableSelect, "获取敏感词表下拉列表")

	// 应用Url
	mid.Sub("common").Reg(apiV1, "/appspace/app/openurl", http.MethodPost, v1.AppUrlCreate, "创建应用Url")
	mid.Sub("common").Reg(apiV1, "/appspace/app/openurl", http.MethodDelete, v1.AppUrlDelete, "删除应用Url")
	mid.Sub("common").Reg(apiV1, "/appspace/app/openurl", http.MethodPut, v1.AppUrlUpdate, "编辑应用Url")
	mid.Sub("common").Reg(apiV1, "/appspace/app/openurl/list", http.MethodGet, v1.GetAppUrlList, "获取应用Url列表")
	mid.Sub("common").Reg(apiV1, "/appspace/app/openurl/status", http.MethodPut, v1.AppUrlStatusSwitch, "启用/停用应用Url")

}
