package middleware

import (
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/UnicomAI/wanwu/pkg/gin-util/route"
	"github.com/gin-gonic/gin"
)

func Init(r *gin.Engine) {

	mid.InitWrapper([]gin.HandlerFunc{Record}, []gin.HandlerFunc{TraceUser})
	r.Use(CacheAvatar)

	// --- openapi ---
	mid.NewSub("openapi", "对外提供原子能力", route.PermNone, false, false)

	// --- callback ---
	mid.NewSub("callback", "系统内部调用", route.PermNone, false, false)

	// --- openurl ---
	mid.NewSub("openurl", "智能体Url", route.PermNone, false, false)

	// --- guest ---
	mid.NewSub("guest", "", route.PermNone, false, false)

	// --- common ---
	mid.NewSub("common", "", route.PermNeedEnable, false, false, JWTUser, CheckUserEnable)

	// --- wga ---
	mid.NewSub("wga", "通用智能体", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// wga.wanwu_bot
	mid.Sub("wga").NewSub("wanwu_bot", "快速对话", route.PermNeedCheck, true, true)

	// wga.openclaw
	mid.Sub("wga").NewSub("openclaw", "OpenClaw", route.PermNeedCheck, true, true)

	// --- ontology ---
	mid.NewSub("ontology", "本体智能体", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// ontology.digital_employee
	mid.Sub("ontology").NewSub("digital_employee", "数字员工", route.PermNeedCheck, true, true)

	// ontology.knowledge_network
	mid.Sub("ontology").NewSub("knowledge_network", "知识网络", route.PermNeedCheck, true, true)

	// ontology.data_source
	mid.Sub("ontology").NewSub("data_source", "数据连接", route.PermNeedCheck, true, true)

	// --- model ---
	mid.NewSub("model", "模型服务", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// model.model_management
	mid.Sub("model").NewSub("model_management", "模型管理", route.PermNeedCheck, true, true)

	// --- resource ---
	mid.NewSub("resource", "资源库", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// resource.knowledge
	mid.Sub("resource").NewSub("knowledge", "知识库", route.PermNeedCheck, true, true)

	// resource.mcp
	mid.Sub("resource").NewSub("mcp", "MCP服务", route.PermNeedCheck, true, true)

	// resource.tool
	mid.Sub("resource").NewSub("tool", "工具", route.PermNeedCheck, true, true)

	// resource.prompt
	mid.Sub("resource").NewSub("prompt", "提示词", route.PermNeedCheck, true, true)

	// resource.safety
	mid.Sub("resource").NewSub("safety", "安全护栏", route.PermNeedCheck, true, true)

	// resource.skill
	mid.Sub("resource").NewSub("skill", "Skills", route.PermNeedCheck, true, true)

	// --- app ---
	mid.NewSub("app", "应用开发", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// app.rag
	mid.Sub("app").NewSub("rag", "知识问答", route.PermNeedCheck, true, true)

	// app.workflow
	mid.Sub("app").NewSub("workflow", "工作流", route.PermNeedCheck, true, true)

	// app.agent
	mid.Sub("app").NewSub("agent", "智能体", route.PermNeedCheck, true, true)

	// --- exploration ---
	mid.NewSub("exploration", "探索广场", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// exploration.mcp
	mid.Sub("exploration").NewSub("mcp", "MCP广场", route.PermNeedCheck, true, true)

	// exploration.app
	mid.Sub("exploration").NewSub("app", "应用广场", route.PermNeedCheck, true, true)

	// exploration.template
	mid.Sub("exploration").NewSub("template", "模板广场", route.PermNeedCheck, true, true)

	// exploration.skill
	mid.Sub("exploration").NewSub("skill", "Skill广场", route.PermNeedCheck, true, true)

	// --- operation ---
	mid.NewSub("operation", "运营管理", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// operation.statistic_client
	//mid.Sub("operation").NewSub("statistic_client", "用户统计", route.PermNeedCheck, true, true)

	// --- api_key ---
	mid.NewSub("api_key", "API Key管理", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// api_key.api_key_management
	mid.Sub("api_key").NewSub("api_key_management", "API Key管理", route.PermNeedCheck, true, true)

	// --- app_observability ---
	mid.NewSub("app_observability", "应用观测", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// --- app_observability.statistic ---
	mid.Sub("app_observability").NewSub("statistic", "统计看板", route.PermNeedCheck, true, true)

	// --- admin_center ---
	mid.NewSub("admin_center", "管理员中心", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// admin_center.oauth
	mid.Sub("admin_center").NewSub("oauth", "OAuth密钥管理", route.PermNeedCheck, true, true)

	// admin_center.setting
	mid.Sub("admin_center").NewSub("setting", "平台配置", route.PermNeedCheck, true, true)

	// open_source
	mid.NewSub("open_source", "开源仓库", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

}
