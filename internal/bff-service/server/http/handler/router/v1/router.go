package v1

import (
	"github.com/gin-gonic/gin"
)

func Register(apiV1 *gin.RouterGroup) {
	// guest
	registerGuest(apiV1)

	// common
	registerCommon(apiV1)

	// callback
	registerV1Callback(apiV1)

	// wga
	registerWGA(apiV1)

	// ontology
	registerOntology(apiV1)

	// model
	registerModel(apiV1)

	// knowledge
	registerKnowledge(apiV1)

	// mcp square
	registerMCPSquare(apiV1)

	// tool
	registerTool(apiV1)

	// skill
	registerAgentSkill(apiV1)

	// safety
	registerSafety(apiV1)

	// rag
	registerRag(apiV1)

	// workflow
	registerWorkflow(apiV1)

	// assistant
	registerAssistant(apiV1)

	// exploration
	registerExploration(apiV1)

	// statistic
	registerStatistic(apiV1)

	// statistic_client
	// registerStatisticClient(apiV1)

	// api_key
	registerAPIKey(apiV1)

	// channel
	registerChannel(apiV1)

	// admin_center
	registerAdminCenter(apiV1)
}
