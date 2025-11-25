package middleware

import (
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/UnicomAI/wanwu/pkg/gin-util/route"
)

func Init() {

	mid.InitWrapper(Record)

	// --- openapi ---
	mid.NewSub("openapi", "Open API", route.PermNone, false, false)

	// --- callback ---
	mid.NewSub("callback", "Internal Callback", route.PermNone, false, false)

	// --- openurl ---
	mid.NewSub("openurl", "Agent URL", route.PermNone, false, false)

	// --- guest ---
	mid.NewSub("guest", "", route.PermNone, false, false)

	// --- common ---
	mid.NewSub("common", "", route.PermNeedEnable, false, false, JWTUser, CheckUserEnable)

	// --- model ---
	mid.NewSub("model", "Model Management", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// --- knowledge ---
	mid.NewSub("knowledge", "Knowledge Base", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// --- mcp ---
	mid.NewSub("mcp", "MCP Platform", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// --- tool ---
	mid.NewSub("tool", "Resource Library", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// --- safety ---
	mid.NewSub("safety", "Safety Guardrails", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// --- rag ---
	mid.NewSub("rag", "Text Q&A", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// --- workflow ---
	mid.NewSub("workflow", "Workflow", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// --- agent ---
	mid.NewSub("agent", "Agent", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// --- exploration ---
	mid.NewSub("exploration", "Application Plaza", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// --- permission ---
	mid.NewSub("permission", "Organization Management", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// permission.user
	mid.Sub("permission").NewSub("user", "User", route.PermNeedCheck, true, true)

	// permission.org
	mid.Sub("permission").NewSub("org", "Organization", route.PermNeedCheck, true, true)

	// permission.role
	mid.Sub("permission").NewSub("role", "Role", route.PermNeedCheck, true, true)

	// --- setting ---
	mid.NewSub("setting", "Platform Configuration", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// --- statistic_client ---
	mid.NewSub("statistic_client", "Statistics", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)

	// --- oauth ---
	mid.NewSub("oauth", "OAuth Key Management", route.PermNeedCheck, true, true, JWTUser, CheckUserPerm)
}
