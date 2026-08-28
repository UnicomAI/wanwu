package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerOntology(apiV1 *gin.RouterGroup) {

	mid.Sub("ontology.digital_employee").Reg(apiV1, "/ontology/skill/select", http.MethodGet, v1.GetOntologySkillSelect, "获取skill选择列表(Ontology专用)")
}
