package params_process

import (
	"context"
	"encoding/json"
	"errors"
	net_url "net/url"
	"time"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
)

type WorkflowIdListParams struct {
	WorkflowIDs []string `json:"workflow_ids"`
}

type WorkflowProcess struct {
}

func init() {
	AddServiceContainer(&WorkflowProcess{})
}

func (k *WorkflowProcess) ServiceType() ServiceType {
	return WorkflowType
}

func (k *WorkflowProcess) Prepare(agent *AgentInfo, prepareParams *AgentPrepareParams, clientInfo *ClientInfo, userQueryParams *UserQueryParams) error {
	workflows, err := buildWorkflowList(agent, clientInfo)
	if err != nil {
		return errors.New("GetAssistantWorkflowsByAssistantID error")
	}
	// workflow ids
	var workflowIDs = buildWorkflowIDList(workflows)
	if len(workflowIDs) == 0 {
		return nil
	}
	list, err := SearchWorkflowByIdList(userQueryParams.Ctx, &WorkflowIdListParams{WorkflowIDs: workflowIDs})
	if err != nil {
		return err
	}
	prepareParams.WorkflowList = list
	return nil
}
func (k *WorkflowProcess) Build(assistant *AgentInfo, prepareParams *AgentPrepareParams, agentChatParams *assistant_service.AgentDetail) error {
	if len(prepareParams.WorkflowList) > 0 {
		for _, schema := range prepareParams.WorkflowList {
			schemaByte, err := json.Marshal(schema)
			if err != nil {
				return err
			}
			//校验schema
			if err = openapi3_util.ValidateSchema(context.Background(), schemaByte); err != nil {
				return err
			}
			agentChatParams.ToolParams.PluginToolList, err = buildPluginList(agentChatParams.ToolParams.PluginToolList, schema, nil, "", "")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func buildWorkflowList(agent *AgentInfo, clientInfo *ClientInfo) ([]*model.AssistantWorkflow, error) {
	if agent.Draft {
		workflows, status := clientInfo.Cli.GetAssistantWorkflowsByAssistantID(context.Background(), agent.Assistant.ID)
		if status != nil {
			return nil, errors.New("GetAssistantWorkflowsByAssistantID error")
		}
		return workflows, nil
	}
	var workflows []*model.AssistantWorkflow
	if agent.AssistantSnapshot.AssistantWorkflowConfig != "" {
		if err := json.Unmarshal([]byte(agent.AssistantSnapshot.AssistantWorkflowConfig), &workflows); err != nil {
			return nil, errors.New("GetAssistantWorkflowsConfigByAssistantID error")
		}
	}
	return workflows, nil
}

// SearchWorkflowByIdList 批量搜索工作流详情
func SearchWorkflowByIdList(ctx context.Context, params *WorkflowIdListParams) ([]map[string]interface{}, error) {
	workflowConfig := config.Cfg().Workflow
	// workflow schemas
	url, _ := net_url.JoinPath(workflowConfig.Endpoint, workflowConfig.ListSchemaUri)
	reqBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	result, err := http_client.Default().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       reqBody,
		Timeout:    time.Minute,
		MonitorKey: "workflow_schema",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return nil, err
	}
	var schemas []map[string]interface{}
	if err = json.Unmarshal(result, &schemas); err != nil {
		return nil, err
	}
	return schemas, nil
}

// buildWorkflowIDList
func buildWorkflowIDList(workflows []*model.AssistantWorkflow) (workflowIDs []string) {
	for _, workflow := range workflows {
		if !workflow.Enable {
			continue
		}
		workflowIDs = append(workflowIDs, workflow.WorkflowId)
	}
	return workflowIDs
}
