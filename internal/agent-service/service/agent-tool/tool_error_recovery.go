package agent_tool

import (
	"github.com/UnicomAI/wanwu/internal/agent-service/model/response"
	agent_util "github.com/UnicomAI/wanwu/internal/agent-service/pkg/util"
	tool_recovery "github.com/UnicomAI/wanwu/internal/agent-service/service/agent-tool/tool-recovery"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

func newToolErrorRecoveryMiddleware(skillTools []tool.BaseTool) compose.ToolMiddleware {
	// Use registered IDs, since some models require hashed skill names.
	skillNames := make(map[string]bool, len(skillTools))
	for _, t := range skillTools {
		if skill, ok := t.(*skillTool); ok {
			skillNames[skill.info.Name] = true
		}
	}
	return tool_recovery.NewMiddleware(func(name string, err error) (string, error) {
		log.Warnf("agent tool %s failed; returning error to model: %v", name, err)
		content, marshalErr := response.ToolErrResp(err)
		if marshalErr != nil {
			return "", marshalErr
		}
		if skillNames[name] {
			return agent_util.BuildToolFinishMessage(content), nil
		}
		return content, nil
	})
}
