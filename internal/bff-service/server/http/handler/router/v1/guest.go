package v1

import (
	"mime"
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/middleware"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerGuest(apiV1 *gin.RouterGroup) {

	// 修复静态文件服务下载.csv文件response header Content-Type为"text/plain; charset=utf-8"的问题
	_ = mime.AddExtensionType(".csv", "text/csv; charset=utf-8")

	apiV1.Static("/static", "./configs/microservice/bff-service/static")

	mid.Sub("guest").Reg(apiV1, "/base/register/email", http.MethodPost, v1.RegisterByEmail, "邮箱注册用户")
	mid.Sub("guest").Reg(apiV1, "/base/register/email/code", http.MethodPost, v1.ResgisterSendEmailCode, "邮箱注册验证码发送")

	mid.Sub("guest").Reg(apiV1, "/base/password/email/code", http.MethodPost, v1.ResetPasswordSendEmailCode, "重置密码邮箱验证码发送")
	mid.Sub("guest").Reg(apiV1, "/base/password/email", http.MethodPost, v1.ResetPasswordByEmail, "邮箱重置密码")

	mid.Sub("guest").Reg(apiV1, "/base/login", http.MethodPost, v1.Login, "用户登录", middleware.LoginRateLimit)
	mid.Sub("guest").Reg(apiV1, "/base/login/email", http.MethodPost, v1.LoginByEmail, "用户邮箱双因子登录", middleware.LoginRateLimit)

	mid.Sub("guest").Reg(apiV1, "/base/captcha", http.MethodGet, v1.GetCaptcha, "获取验证码")
	mid.Sub("guest").Reg(apiV1, "/base/custom", http.MethodGet, v1.GetLogoCustomInfo, "自定义logo和title")
	mid.Sub("guest").Reg(apiV1, "/base/language/select", http.MethodGet, v1.GetLanguageSelect, "获取语言列表（用于下拉选择）")
	mid.Sub("guest").Reg(apiV1, "/custom/about", http.MethodGet, v1.GetReleaseNotes, "获取版本和更新日志")
	mid.Sub("guest").Reg(apiV1, "/base/rsa/public-key", http.MethodGet, v1.GetRSAPublicKey, "获取RSA公钥")

	mid.Sub("guest").Reg(apiV1, "/workflow/template/list", http.MethodGet, v1.GetWorkflowTemplateList, "获取工作流模板列表")
	mid.Sub("guest").Reg(apiV1, "/workflow/template/detail", http.MethodGet, v1.GetWorkflowTemplateDetail, "获取工作流模板详情")
	mid.Sub("guest").Reg(apiV1, "/workflow/template/recommend", http.MethodGet, v1.GetWorkflowTemplateRecommend, "获取工作流模板推荐")
	mid.Sub("guest").Reg(apiV1, "/workflow/template/download", http.MethodGet, v1.DownloadWorkflowTemplate, "下载工作流模板")

}
