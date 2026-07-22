package client

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/orm"
)

type IClient interface {
	CheckMCPExist(ctx context.Context, orgID, userID, mcpSquareID string) (bool, *errs.Status)
	GetMCP(ctx context.Context, mcpID uint32) (*model.MCPClient, *errs.Status)
	CreateMCP(ctx context.Context, mcp *model.MCPClient) *errs.Status
	UpdateMCP(ctx context.Context, mcp *model.MCPClient) *errs.Status
	DeleteMCP(ctx context.Context, mcpID uint32) *errs.Status
	ListMCPs(ctx context.Context, orgID, userID, name string) ([]*model.MCPClient, *errs.Status)
	ListMCPsAdmin(ctx context.Context, orgIDs, userIDs []string, name string, pageNo, pageSize int) ([]*model.MCPClient, int64, *errs.Status)
	ListMCPsByMCPIdList(ctx context.Context, mcpIDList []uint32) ([]*model.MCPClient, *errs.Status)

	CreateCustomTool(ctx context.Context, customTool *model.CustomTool) *errs.Status
	GetCustomTool(ctx context.Context, customTool *model.CustomTool) (*model.CustomTool, *errs.Status)
	ListCustomTools(ctx context.Context, orgID, userID, name string) ([]*model.CustomTool, *errs.Status)
	ListCustomToolsAdmin(ctx context.Context, orgIDs, userIDs []string, name string, pageNo, pageSize int) ([]*model.CustomTool, int64, *errs.Status)
	ListCustomToolsByCustomToolIDs(ctx context.Context, customToolIDs []uint32) ([]*model.CustomTool, *errs.Status)
	UpdateCustomTool(ctx context.Context, customTool *model.CustomTool) *errs.Status
	DeleteCustomTool(ctx context.Context, customToolID uint32) *errs.Status
	ListBuiltinTools(ctx context.Context, orgID, userID string) ([]*model.BuiltinTool, *errs.Status)
	ListBuiltinToolsBySquareIdList(ctx context.Context, squareList []string) ([]*model.BuiltinTool, *errs.Status)
	GetBuiltinTool(ctx context.Context, builtinTool *model.BuiltinTool) (*model.BuiltinTool, *errs.Status)

	UpdateBuiltinTool(ctx context.Context, builtinTool *model.BuiltinTool) *errs.Status
	CreateBuiltinTool(ctx context.Context, customTool *model.BuiltinTool) *errs.Status

	GetMCPServer(ctx context.Context, mcpServerId string) (*model.MCPServer, *errs.Status)
	CreateMCPServer(ctx context.Context, mcpServer *model.MCPServer) *errs.Status
	UpdateMCPServer(ctx context.Context, mcpServer *model.MCPServer) *errs.Status
	DeleteMCPServer(ctx context.Context, mcpServerId string) *errs.Status
	ListMCPServers(ctx context.Context, orgID, userID, name string) ([]*model.MCPServer, *errs.Status)
	ListMCPServersAdmin(ctx context.Context, orgIDs, userIDs []string, name string, pageNo, pageSize int) ([]*model.MCPServer, int64, *errs.Status)
	ListMCPServerByIdList(ctx context.Context, mcpServerIdList []string) ([]*model.MCPServer, *errs.Status)
	GetMCPServerTool(ctx context.Context, mcpServerToolId string) (*model.MCPServerTool, *errs.Status)
	CreateMCPServerTool(ctx context.Context, mcpServerTools []*model.MCPServerTool) *errs.Status
	UpdateMCPServerTool(ctx context.Context, mcpServerTool *model.MCPServerTool) *errs.Status
	DeleteMCPServerTool(ctx context.Context, mcpServerToolId string) *errs.Status
	ListMCPServerTools(ctx context.Context, mcpServerId string) ([]*model.MCPServerTool, *errs.Status)
	CountMCPServerTools(ctx context.Context, mcpServerId string) (int64, *errs.Status)

	//================CustomSkill================
	CreateCustomSkill(ctx context.Context, customSkill *model.CustomSkill) (string, *errs.Status)
	DeleteCustomSkill(ctx context.Context, skillId string) *errs.Status
	GetCustomSkill(ctx context.Context, skillId string) (*model.CustomSkill, *errs.Status)
	GetCustomSkillByPreviewThreadID(ctx context.Context, previewThreadId string) (*model.CustomSkill, *errs.Status)
	GetCustomSkillByWgaThreadID(ctx context.Context, wgaThreadID string) (*model.CustomSkill, *errs.Status)
	GetCustomSkillListByWgaThreadIDList(ctx context.Context, userId, orgId string, wgaThreadIDList []string) ([]*model.CustomSkill, *errs.Status)
	GetCustomSkillList(ctx context.Context, userId, orgId, name string) ([]*model.CustomSkill, int64, *errs.Status)
	GetCustomSkillPageList(ctx context.Context, userId, orgId []string, name string, pageNum, pageSize int) ([]*model.CustomSkill, int64, *errs.Status)
	GetCustomSkillBySkillIds(ctx context.Context, skillIds []string) ([]*model.CustomSkill, *errs.Status)
	UpdateCustomSkillBasicMeta(ctx context.Context, skillId, name, desc string) *errs.Status
	UpdateCustomSkillThreadMeta(ctx context.Context, skillId, wgaThreadId, previewThreadId string) *errs.Status
	CreateCustomSkillVar(ctx context.Context, userId, orgId string, variable *model.CustomSkillVariable) (uint32, *errs.Status)
	UpdateCustomSkillVar(ctx context.Context, userId, orgId string, id uint32, variable *model.CustomSkillVariable) *errs.Status
	DeleteCustomSkillVar(ctx context.Context, userId, orgId string, id uint32) *errs.Status
	GetCustomSkillVarByID(ctx context.Context, id uint32) (*model.CustomSkillVariable, *errs.Status)
	GetCustomSkillVars(ctx context.Context, userId, orgId, skillId string) ([]*model.CustomSkillVariable, *errs.Status)
	GetCustomSkillVarsBySkillIDs(ctx context.Context, userId, orgId string, skillIds []string) (map[string][]*model.CustomSkillVariable, *errs.Status)
	PublishCustomSkill(ctx context.Context, publish *model.CustomSkillPublish, snapshot *orm.CustomSkillPublishSnapshot) *errs.Status
	UpdatePublishCustomSkill(ctx context.Context, skillId, versionDesc string) *errs.Status
	GetPublishCustomSkillHistoryList(ctx context.Context, skillId string) ([]*model.CustomSkillPublish, int64, *errs.Status)
	GetPublishCustomSkillByLatest(ctx context.Context, skillId string) (*model.CustomSkillPublish, *errs.Status)
	GetPublishCustomSkillByVersion(ctx context.Context, skillId, version string) (*model.CustomSkillPublish, *errs.Status)
	GetPublishCustomSkillByIDList(ctx context.Context, skillIdList []string) ([]*model.CustomSkillPublish, *errs.Status)

	//================AcquiredSkill================
	CreateAcquiredSkill(ctx context.Context, acquiredSkill *model.AcquiredSkill) (string, *errs.Status)
	DeleteAcquiredSkill(ctx context.Context, acquiredSkillId string) *errs.Status
	GetAcquiredSkill(ctx context.Context, acquiredSkillId string) (*model.AcquiredSkill, *errs.Status)
	GetAcquiredSkillByIDList(ctx context.Context, acquiredSkillIdList []string) ([]*model.AcquiredSkill, *errs.Status)
	GetAcquiredSkillList(ctx context.Context, userId, orgId, name string) ([]*model.AcquiredSkill, int64, *errs.Status)
	CreateAcquiredSkillVar(ctx context.Context, userId, orgId string, variable *model.AcquiredSkillVariable) (uint32, *errs.Status)
	UpdateAcquiredSkillVar(ctx context.Context, userId, orgId string, id uint32, variable *model.AcquiredSkillVariable) *errs.Status
	DeleteAcquiredSkillVar(ctx context.Context, userId, orgId string, id uint32) *errs.Status
	GetAcquiredSkillVarByID(ctx context.Context, id uint32) (*model.AcquiredSkillVariable, *errs.Status)
	GetAcquiredSkillVars(ctx context.Context, userId, orgId, acquiredSkillId string) ([]*model.AcquiredSkillVariable, *errs.Status)
	GetAcquiredSkillVarsByAcquiredSkillIDs(ctx context.Context, userId, orgId string, acquiredSkillIDs []string) (map[string][]*model.AcquiredSkillVariable, *errs.Status)

	//================BuiltinSkill================
	CreateBuiltinSkillVar(ctx context.Context, userId, orgId string, variable *model.BuiltinSkillVariable) (uint32, *errs.Status)
	UpdateBuiltinSkillVar(ctx context.Context, userId, orgId string, id uint32, variable *model.BuiltinSkillVariable) *errs.Status
	DeleteBuiltinSkillVar(ctx context.Context, userId, orgId string, id uint32) *errs.Status
	GetBuiltinSkillVars(ctx context.Context, userId, orgId, skillId string) ([]*model.BuiltinSkillVariable, int64, *errs.Status)
	GetBuiltinSkillsVars(ctx context.Context, userId, orgId string, skillIds []string) (map[string][]*model.BuiltinSkillVariable, *errs.Status)

	//================SkillStatistic================
	IncrementCustomSkillDownloadCount(ctx context.Context, skillId string) *errs.Status
	IncrementCustomSkillAcquiredCount(ctx context.Context, skillId string) *errs.Status
	IncrementBuiltinSkillDownloadCount(ctx context.Context, skillId string) *errs.Status
	GetBuiltinSkillDownloadCounts(ctx context.Context, skillIds []string) (map[string]int32, *errs.Status)
}
