package agent_tool

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/internal/agent-service/model/response"
	agent_config "github.com/UnicomAI/wanwu/internal/agent-service/pkg/config"
	agent_util "github.com/UnicomAI/wanwu/internal/agent-service/pkg/util"
	minio_service "github.com/UnicomAI/wanwu/internal/agent-service/service/minio-service"
	service_model "github.com/UnicomAI/wanwu/internal/agent-service/service/service-model"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	safe_go_util "github.com/UnicomAI/wanwu/pkg/safe-go-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	wga_sandbox "github.com/UnicomAI/wanwu/pkg/wga-sandbox"
	wga_sandbox_converter "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-converter"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

// 使用 map 存储特殊字符，查找效率 O(1)
var specialNameSet = map[rune]bool{
	'(':  true,
	'（':  true,
	'[':  true,
	'【':  true,
	'{':  true,
	'《':  true,
	'<':  true,
	'“':  true,
	'\'': true,
}

type SkillParams struct {
	UploadFileUrl []string `json:"uploadFileUrl"`
}

// skillTool 实现了 tool.StreamableTool 接口
type skillTool struct {
	skillName   string // 处理后的技能名称
	info        *schema.ToolInfo
	Skill       *request.SkillToolInfo
	userQuery   string
	uploadFile  []string
	agentName   string
	chatInfo    *service_model.AgentChatInfo
	modelParams *request.ModelParams // 智能体配置里的 LLM 参数，透传到沙箱内 chat model
}

type skillRunEnv struct {
	runId          string
	inputDir       string
	outputDir      string
	skillDir       string
	runDir         string
	enableThinking bool //是否输出智能体思考过程
	// variables：本次 skill 执行注入到 sandbox 的自定义变量集合。
	// 严格约束：仅供 buildSkillOptions 透传到 wga_sandbox_option.Skill.Variables，不得用于日志/SKILL.md。
	variables []wga_sandbox_option.SkillVariable
}

type SkillExecutor struct {
	skill    *skillTool
	ctx      context.Context
	messages []adk.Message
	runEnv   *skillRunEnv
	err      error
}

// GetToolsFromSkills 根据技能列表创建工具列表
func GetToolsFromSkills(ctx context.Context, skillToolList []*request.SkillToolInfo, query, agentName string, uploadFile []string, chatInfo *service_model.AgentChatInfo, changeToolName bool, modelParams *request.ModelParams) ([]tool.BaseTool, map[string]*request.ToolConfig, error) {
	if len(skillToolList) == 0 {
		return nil, nil, nil
	}
	var toolMap = make(map[string]*request.ToolConfig)
	rs := lo.Map(skillToolList, func(skill *request.SkillToolInfo, index int) tool.BaseTool {
		skillToolInfo := buildSkillTool(skill, query, agentName, uploadFile, chatInfo, changeToolName, modelParams)
		var toolId = skillToolInfo.info.Name
		toolMap[toolId] = &request.ToolConfig{
			Avatar:   skillToolInfo.Skill.Avatar,
			ToolName: skillToolInfo.skillName,
			ToolID:   toolId,
		}
		return skillToolInfo
	})
	return rs, toolMap, nil
}

// Info 返回工具的元信息
func (t *skillTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	marshal, _ := json.Marshal(t.info)
	log.Infof("skillTool %v", string(marshal))
	return t.info, nil
}

//// InvokableRun 执行工具
//// 需要测试上下文数据传递，和文件传递
//func (t *skillTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
//	conv := wga_sandbox_converter.NewEinoConverter(wga_sandbox_option.RunnerTypeOpencode)
//	run, err := createExecutor(ctx, t).initEnv().initMessages().run(conv, skillOutputProcessor)
//	if err != nil {
//		return "", err
//	}
//	return run, nil
//}

func (t *skillTool) StreamableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (*schema.StreamReader[string], error) {
	log.Infof("skillTool StreamableRun %s", argumentsInJSON)
	return createExecutor(ctx, t).initEnv(argumentsInJSON).initMessages().runStream(wga_sandbox_option.RunnerTypeEinoChatModel)
}

// buildSkillTool 构建技能工具
func buildSkillTool(skill *request.SkillToolInfo, query, agentName string, uploadFile []string, chatInfo *service_model.AgentChatInfo, changeToolName bool, modelParams *request.ModelParams) *skillTool {
	toolInfo, skillName := buildToolInfo(skill, changeToolName)
	return &skillTool{
		skillName:   skillName,
		info:        toolInfo,
		Skill:       skill,
		userQuery:   query,
		agentName:   agentName,
		chatInfo:    chatInfo,
		uploadFile:  uploadFile,
		modelParams: modelParams,
	}
}

func buildToolInfo(skill *request.SkillToolInfo, changeToolName bool) (*schema.ToolInfo, string) {
	skillName := agent_util.AgentSkillPrefix + TrimBeforeSpecial(skill.Name, specialNameSet)
	var changeName = skillName
	if changeToolName {
		changeName = agent_util.MD5(skillName)
	}
	toolInfo := &schema.ToolInfo{
		Name: changeName,
		Desc: skill.Desc,
	}
	templateConfig := agent_config.GetToolTemplateConfig()
	skillConfig, _ := templateConfig.GetToolByID(agent_config.SkillTool)
	if skillConfig != nil {
		apiSchema, _ := GetEnioToolsFromOpenAPISchema(context.Background(), skillConfig)
		if len(apiSchema) > 0 {
			toolInfo.ParamsOneOf = apiSchema[0].ParamsOneOf
		}
	}
	return toolInfo, skillName
}

func TrimBeforeSpecial(s string, specialSet map[rune]bool) string {
	for i, r := range s {
		if specialSet[r] {
			return s[:i]
		}
	}
	return s
}

// 创建技能执行器
func createExecutor(ctx context.Context, skill *skillTool) *SkillExecutor {
	return &SkillExecutor{
		skill: skill,
		ctx:   ctx,
	}
}

func (s *SkillExecutor) initEnv(argumentsInJSON string) *SkillExecutor {
	params := &SkillParams{}
	_ = json.Unmarshal([]byte(argumentsInJSON), params)
	//初始化skill运行环境
	skill := s.skill.Skill
	var runId = util.MD5([]byte(skill.Name)) + "-" + uuid.New().String()
	dir, err := CreateSkillDir(runId, skill, s.skill.uploadFile, params)
	if err != nil {
		s.err = err
		return s
	}
	s.runEnv = &skillRunEnv{
		runId:          runId,
		inputDir:       dir.InputDir,
		outputDir:      dir.OutputDir,
		skillDir:       dir.SkillDir,
		runDir:         dir.RunDir,
		enableThinking: false,
		variables:      toSandboxSkillVariables(skill.Variables),
	}
	return s
}

// toSandboxSkillVariables 把 agent-svc 本包 SkillVariable 映射成 wga-sandbox option 形态。
// 仅做字段拷贝；任何过滤/校验放到 sandbox 层统一处理，避免双向维护两套规则。
func toSandboxSkillVariables(vars []request.SkillVariable) []wga_sandbox_option.SkillVariable {
	if len(vars) == 0 {
		return nil
	}
	out := make([]wga_sandbox_option.SkillVariable, 0, len(vars))
	for _, v := range vars {
		out = append(out, wga_sandbox_option.SkillVariable{
			Name:          v.Name,
			Description:   v.Desc,
			VariableKey:   v.VariableKey,
			VariableValue: v.VariableValue,
		})
	}
	return out
}

// clearEnv 清理运行环境，删除 runDir 下的所有数据
func (s *SkillExecutor) clearEnv() {
	if s.runEnv != nil && s.runEnv.runDir != "" {
		log.Errorf("skillTool clearEnv %s", s.runEnv.runDir)
		err := os.RemoveAll(s.runEnv.runDir)
		if err != nil {
			log.Errorf("skillTool clearEnv error: %v", err)
		}
	}
}

func (s *SkillExecutor) initMessages() *SkillExecutor {
	history, _ := agent_util.GetEnioReactChatHistory(s.ctx, s.skill.agentName)
	if len(history) == 0 {
		history = agent_util.UserMessage(s.skill.userQuery)
	}
	s.messages = history
	return s
}

func (s *SkillExecutor) runStream(runnerType wga_sandbox_option.RunnerType) (*schema.StreamReader[string], error) {
	if s.err != nil {
		return nil, s.err
	}
	sr, sw := schema.Pipe[string](1)
	conv := wga_sandbox_converter.NewEinoConverter(runnerType)
	skillOpts := buildSkillOptions(buildModeConfig(s.skill.chatInfo, s.skill.modelParams), s.runEnv, s.messages, runnerType)
	//执行调用
	_, jsonCh, err := wga_sandbox.Run(s.ctx, skillOpts...)
	var errResp string
	if err != nil {
		log.Errorf("skillTool strem run error: %v", err)
		errResp, err = response.ToolErrResp(err)
		if err != nil {
			return nil, errors.New(agent_util.WgaErr)
		}
	}

	safe_go_util.SafeGo(func() {
		defer func() {
			sw.Close()
			// 清理运行环境
			s.clearEnv()
		}()
		if len(errResp) == 0 {
			for ch := range jsonCh {
				if conv != nil {
					contentList, err1 := conv.Convert(ch)
					if err1 != nil || len(contentList) == 0 {
						log.Errorf("skillTool Run error: %v", err1)
						continue
					}
					for _, message := range contentList {
						if message.ResponseMeta != nil && message.ResponseMeta.FinishReason == "stop" {
							//自己手动添加结束，不需要沙箱的结束
							message.ResponseMeta.FinishReason = agent_util.AgentSkillWgaStop
						}
						marshal, err2 := json.Marshal(message)
						if err2 != nil {
							log.Errorf("skillTool Run error: %v", err1)
							continue
						}
						log.Infof("skillTool StreamableRun result %s", string(marshal))
						sw.Send(string(marshal), nil)
					}
				}
			}
			processor, extra, _ := skillOutputProcessor(s.runEnv.outputDir)
			if len(processor) > 0 {
				sw.Send(agent_util.BuildAssistantMessage(processor, extra), nil)
			}
		}

		sw.Send(agent_util.BuildToolFinishMessage(errResp), nil)
	})

	return sr, nil
}

// downloadInputFile 下载输入文件
func downloadInputFile(inputDir string, skillParams *SkillParams) {
	if len(skillParams.UploadFileUrl) > 0 {
		for _, file := range skillParams.UploadFileUrl {
			localFilePath := filepath.Join(inputDir, filepath.Base(file))
			err := minio_service.DownloadFileToLocal(context.Background(), file, localFilePath)
			if err != nil {
				log.Errorf("skillTool downloadInputFile error: %v", err)
			}
		}
	}
}

func skillOutputProcessor(outputDir string) (string, map[string]any, error) {
	var builder = &strings.Builder{}
	//结果文件处理
	fileList, err := uploadResultFile(outputDir)
	if err != nil {
		log.Errorf("skillTool uploadResultFile error: %v", err)
		return "", nil, err
	}
	var extra = map[string]any{}
	if len(fileList) > 0 {
		builder.WriteString("\n 生成文件结果如下-ReplaceLocalFile:")
		for _, file := range fileList {
			builder.WriteString(util.MdImageFile(file.FileName, file.FilePath))
		}
		extra["fileList"] = fileList
	}
	return builder.String(), extra, nil
}

// 上传结果文件
func uploadResultFile(outputDir string) ([]*response.DownloadFileInfo, error) {
	list, err := util.FindDirAndFileList(outputDir, false, true, true)
	if err != nil {
		return nil, err
	}
	var fileList []*response.DownloadFileInfo
	for _, fileInfo := range list {
		path := buildFilePath(fileInfo)
		log.Infof("uploadResultFile %s", path)
		fileName, minioPath, fileSize, err1 := minio_service.UploadLocalFile(context.Background(), uuid.New().String(), filepath.Base(path), path, true)
		if err1 != nil {
			return nil, err1
		}
		fileList = append(fileList, &response.DownloadFileInfo{
			FileName: fileName,
			FilePath: minioPath,
			FileSize: fileSize,
			CreateAt: util.Time2Str(time.Now().UnixMilli()),
		})
	}
	return fileList, nil
}

func buildFilePath(info *util.FileInfo) string {
	if !info.IsDir {
		return info.FilePath
	}
	dirZipName := buildDirZipName(info.FilePath)
	err := util.ZipDirToLocal(info.FilePath, dirZipName)
	if err != nil {
		log.Errorf("skillTool uploadResultFile error: %v", err)
		return ""
	}
	return dirZipName
}

func buildDirZipName(path string) string {
	return filepath.Base(path) + ".zip"
}

func buildModeConfig(chatInfo *service_model.AgentChatInfo, modelParams *request.ModelParams) *wga_sandbox_option.ModelConfig {
	modelInfo := chatInfo.ModelInfo
	modelConfig := wga_sandbox_option.ModelConfig{
		Provider:     modelInfo.Provider,
		ProviderName: modelInfo.Provider,
		Model:        modelInfo.Model,
		ModelName:    modelInfo.DisplayName,
		Params:       toLLMParams(modelParams),
	}
	endpoint := mp.ToModelEndpoint(modelInfo.ModelId, modelInfo.Model)
	for k, v := range endpoint {
		if k == "model_url" {
			modelConfig.BaseURL = v.(string)
			break
		}
	}
	return &modelConfig
}

// toLLMParams 把 agent-service 内部的 request.ModelParams（*float32/*int 指针形态，
// 非 nil 即启用）转成沙箱 ModelConfig.Params 期望的 *mp_common.LLMParams（float64+bool 开关）。
// 指针为 nil 时对应 *Enable=false（该参数不启用，沙箱内用模型默认值）。
// ThinkingEnable 不设：保持 nil，沙箱内 NewNoReasonChatModel 继续硬编码 enable_thinking:false。
func toLLMParams(p *request.ModelParams) *mp_common.LLMParams {
	if p == nil {
		return nil
	}
	out := &mp_common.LLMParams{}
	if p.Temperature != nil {
		out.TemperatureEnable = true
		out.Temperature = float64(*p.Temperature)
	}
	if p.TopP != nil {
		out.TopPEnable = true
		out.TopP = float64(*p.TopP)
	}
	if p.FrequencyPenalty != nil {
		out.FrequencyPenaltyEnable = true
		out.FrequencyPenalty = float64(*p.FrequencyPenalty)
	}
	if p.PresencePenalty != nil {
		out.PresencePenaltyEnable = true
		out.PresencePenalty = float64(*p.PresencePenalty)
	}
	if p.MaxTokens != nil {
		out.MaxTokensEnable = true
		out.MaxTokens = int32(*p.MaxTokens)
	}
	return out
}
func buildSkillOptions(modelConfig *wga_sandbox_option.ModelConfig, runEnv *skillRunEnv, messages []adk.Message, runnerType wga_sandbox_option.RunnerType) []wga_sandbox_option.Option {
	skillCreatorCfg := config.SkillCreatorConfig{
		EnableThinking: runEnv.enableThinking,
		Skills: []config.SkillConfig{{
			Dir: runEnv.skillDir,
		}},
	}
	opts := []wga_sandbox_option.Option{
		wga_sandbox_option.WithRunSession(wga_sandbox_option.RunSession{RunID: runEnv.runId}),
		wga_sandbox_option.WithModelConfig(*modelConfig),
		wga_sandbox_option.WithOutputDir(runEnv.outputDir),
		wga_sandbox_option.WithMessages(messages),
		wga_sandbox_option.WithEnableThinking(skillCreatorCfg.EnableThinking),
		wga_sandbox_option.WithRunnerType(runnerType),
	}

	if runEnv.inputDir != "" {
		opts = append(opts, wga_sandbox_option.WithInputDir(filepath.Clean(runEnv.inputDir)+"/."))
	}

	if skillCreatorCfg.Instruction != "" {
		opts = append(opts, wga_sandbox_option.WithInstruction(skillCreatorCfg.Instruction))
	}

	sandboxCfg := agent_config.GetConfig().WgaSandbox.Sandbox
	switch sandboxCfg.Type {
	case string(wga_sandbox_option.SandboxTypeOneshot):
		opts = append(opts, wga_sandbox_option.WithSandbox(wga_sandbox_option.SandboxOneshot(sandboxCfg.ImageName)))
	default:
		opts = append(opts, wga_sandbox_option.WithSandbox(wga_sandbox_option.SandboxReuse(sandboxCfg.Host)))
	}

	if len(skillCreatorCfg.Skills) > 0 {
		skills := make([]wga_sandbox_option.Skill, len(skillCreatorCfg.Skills))
		for i, skill := range skillCreatorCfg.Skills {
			skills[i] = wga_sandbox_option.Skill{
				Dir: skill.Dir,
				// 当前 skillCreatorCfg.Skills 由 runEnv.skillDir 单条构造（见上文），
				// runEnv.variables 即为该 skill 的变量集合，直接复用即可。
				Variables: runEnv.variables,
			}
		}
		opts = append(opts, wga_sandbox_option.WithSkills(skills))
	}

	return opts
}
