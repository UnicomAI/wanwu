package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// CreateMCP
//
//	@Tags			mcp
//	@Summary		创建自定义MCP [EN] @Summary Create a custom MCP
//	@Description	创建自定义MCP [EN] @Description Create a custom MCP
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.MCPCreate	true	"自定义MCP信息" [EN] @Param data body request.MCPCreate true "Customized MCP information"
//	@Success		200		{object}	response.Response{}
//	@Router			/mcp [post]
func CreateMCP(ctx *gin.Context) {
	var req request.MCPCreate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.CreateMCP(ctx, getUserID(ctx), getOrgID(ctx), req))
}

// UpdateMCP
//
//	@Tags			mcp
//	@Summary		修改自定义MCP [EN] @Summary Modify custom MCP
//	@Description	修改自定义MCP [EN] @Description Modify custom MCP
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.MCPUpdate	true	"自定义MCP信息" [EN] @Param data body request.MCCPupdate true "Customized MCP information"
//	@Success		200		{object}	response.Response{}
//	@Router			/mcp [put]
func UpdateMCP(ctx *gin.Context) {
	var req request.MCPUpdate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.UpdateMCP(ctx, getUserID(ctx), getOrgID(ctx), req))
}

// GetMCP
//
//	@Tags			mcp
//	@Summary		获取自定义MCP详情 [EN] @Summary Get custom MCP details
//	@Description	获取自定义MCP详情 [EN] @Description Get custom MCP details
//	@Accept			json
//	@Produce		json
//	@Param			mcpId	query		string	true	"mcpId"
//	@Success		200		{object}	response.Response{data=response.MCPDetail}
//	@Router			/mcp [get]
func GetMCP(ctx *gin.Context) {
	resp, err := service.GetMCP(ctx, ctx.Query("mcpId"))
	gin_util.Response(ctx, resp, err)
}

// DeleteMCP
//
//	@Tags			mcp
//	@Summary		删除自定义MCP [EN] @Summary Delete custom MCP
//	@Description	删除自定义MCP [EN] @Description Delete custom MCP
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.MCPIDReq	true	"mcpId"
//	@Success		200		{object}	response.Response{}
//	@Router			/mcp [delete]
func DeleteMCP(ctx *gin.Context) {
	var req request.MCPIDReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	gin_util.Response(ctx, nil, service.DeleteMCP(ctx, req.MCPID))
}

// GetMCPList
//
//	@Tags			mcp
//	@Summary		获取自定义MCP列表 [EN] @Summary Get a custom MCP list
//	@Description	获取自定义MCP列表 [EN] @Description Get custom MCP list
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"mcp名称" [EN] @Param name query string false "mcp name"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.MCPInfo}}
//	@Router			/mcp/list [get]
func GetMCPList(ctx *gin.Context) {
	resp, err := service.GetMCPList(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetMCPTools
//
//	@Tags			mcp
//	@Summary		获取MCP Tool列表 [EN] @Summary Get the MCP Tool list
//	@Description	获取MCP Tool列表 [EN] @Description Get the MCP Tool list
//	@Accept			json
//	@Produce		json
//	@Param			mcpId		query		string	false	"mcpId(和serverUrl传一个)" [EN] @Param mcpId query string false "mcpId (pass the same one as serverUrl)"
//	@Param			serverUrl	query		string	false	"serverUrl,就是sseUrl(和mcpId传一个)" [EN] @Param serverUrl query string false "serverUrl, which is sseUrl (pass the same as mcpId)"
//	@Success		200			{object}	response.Response{data=response.MCPToolList}
//	@Router			/mcp/tool/list [get]
func GetMCPTools(ctx *gin.Context) {
	resp, err := service.GetMCPToolList(ctx, ctx.Query("mcpId"), ctx.Query("serverUrl"))
	gin_util.Response(ctx, resp, err)
}

// GetMCPSelect
//
//	@Tags			mcp
//	@Summary		获取自定义MCP列表（用于下拉选择） [EN] @Summary Get a custom MCP list (for drop-down selection)
//	@Description	获取自定义MCP列表（用于下拉选择） [EN] @Description Get a custom MCP list (for drop-down selection)
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"mcp名称" [EN] @Param name query string false "mcp name"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.MCPSelect}}
//	@Router			/mcp/select [get]
func GetMCPSelect(ctx *gin.Context) {
	resp, err := service.GetMCPSelect(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetMCPActionList
//
//	@Tags			mcp
//	@Summary		获取MCP Action列表 [EN] @Summary Get the MCP Action list
//	@Description	获取MCP Action列表 [EN] @Description Get MCP Action list
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.MCPActionListReq	true	"mcp信息" [EN] @Param data query request.MCPActionListReq true "mcp information"
//	@Success		200		{object}	response.Response{data=response.MCPActionList}
//	@Router			/mcp/action/list [get]
func GetMCPActionList(ctx *gin.Context) {
	var req request.MCPActionListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetMCPActionList(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}
