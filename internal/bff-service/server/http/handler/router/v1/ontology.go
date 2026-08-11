package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/middleware"
	"github.com/UnicomAI/wanwu/pkg/constant"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerOntology(apiV1 *gin.RouterGroup) {

	mid.Sub("ontology.digital_employee").Reg(apiV1, "/ontology/skill/select", http.MethodGet, v1.GetOntologySkillSelect, "获取skill选择列表(Ontology专用)")

	// 数字员工发布会话相关接口（复用 ontology.digital_employee 子权限树）
	mid.Sub("ontology.digital_employee").Reg(apiV1, "/digital-employee/conversation", http.MethodPost, v1.CreateDigitalEmployeeConversation, "创建数字员工发布会话")
	mid.Sub("ontology.digital_employee").Reg(apiV1, "/digital-employee/conversation", http.MethodDelete, v1.DeleteDigitalEmployeeConversation, "删除数字员工发布会话")
	mid.Sub("ontology.digital_employee").Reg(apiV1, "/digital-employee/conversation/list", http.MethodGet, v1.GetDigitalEmployeeConversationList, "数字员工发布会话列表")
	mid.Sub("ontology.digital_employee").Reg(apiV1, "/digital-employee/conversation/detail", http.MethodGet, v1.GetDigitalEmployeeConversationDetail, "数字员工发布会话详情")
	mid.Sub("ontology.digital_employee").Reg(apiV1, "/digital-employee/conversation/config", http.MethodGet, v1.GetDigitalEmployeeConversationConfig, "数字员工发布会话配置")
	mid.Sub("ontology.digital_employee").Reg(apiV1, "/digital-employee/conversation/config", http.MethodPut, v1.UpdateDigitalEmployeeConversationConfig, "修改数字员工发布会话配置")
	mid.Sub("ontology.digital_employee").Reg(apiV1, "/digital-employee/chat", http.MethodPost, v1.DigitalEmployeeChat, "数字员工发布对话", append([]gin.HandlerFunc{
		middleware.TraceWeb(constant.BizModuleAppDigitalEmployee, middleware.WithAppResource(constant.AppTypeDigitalEmployee, "employeeId")),
	}, middleware.AppHistoryRecord("employeeId", constant.AppTypeDigitalEmployee))...)
}
