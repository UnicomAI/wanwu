package request

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/UnicomAI/wanwu/pkg/util"
)

// OpenAPI 智能体配置参数边界
const (
	agentPrologueMaxLen        = 100
	agentInstructionsMaxLen    = 20000
	agentRecommendQMaxCount    = 4
	agentRecommendQMaxLen      = 50
	agentRecommendPromptMaxLen = 5000

	agentMemoryRoundsMin     = 0
	agentMemoryRoundsMax     = 100
	agentRecommendHistoryMin = 0
	agentRecommendHistoryMax = 5

	agentMaxTokensMin     = 1
	agentMaxTokensMax     = 1048576 // 防滥用/溢出兜底，远高于现有模型实际上限
	agentKnowledgeTopKMax = 10

	agentWeightSumTolerance = 0.01 // 权重和=1 的浮点容差（滑块步长 0.1）
)

var (
	agentMatchTypeEnum  = map[string]bool{"vector": true, "text": true, "mix": true}
	agentRerankTypeEnum = map[string]bool{"rerank": true, "multimodal-rerank": true}
)

// validateAgentConfigUpdate 校验 OpenAPI 智能体配置更新请求的长度/范围/枚举/数组元素。
// 仅校验请求中实际提供的字段（指针为 nil 或 Config 为空时跳过，交由 handler 填默认值）。
func validateAgentConfigUpdate(req *OpenAPIAgentConfigUpdateRequest) error {
	// 开场白长度（按 Unicode 字符数）
	if req.Prologue != nil && utf8.RuneCountInString(*req.Prologue) > agentPrologueMaxLen {
		return fmt.Errorf("prologue length exceeds %d", agentPrologueMaxLen)
	}
	// 系统提示词长度
	if req.Instructions != nil && utf8.RuneCountInString(*req.Instructions) > agentInstructionsMaxLen {
		return fmt.Errorf("instructions length exceeds %d", agentInstructionsMaxLen)
	}
	// 推荐问数量
	if len(req.RecommendQuestion) > agentRecommendQMaxCount {
		return fmt.Errorf("recommendQuestion count exceeds %d", agentRecommendQMaxCount)
	}
	// 推荐问数组元素：不能为空（含 null→""）、单条长度
	for i, q := range req.RecommendQuestion {
		if strings.TrimSpace(q) == "" {
			return fmt.Errorf("recommendQuestion[%d] cannot be empty", i)
		}
		if utf8.RuneCountInString(q) > agentRecommendQMaxLen {
			return fmt.Errorf("recommendQuestion item length exceeds %d", agentRecommendQMaxLen)
		}
	}
	// 大模型采样参数范围（temperature/topP/penalty/maxTokens）
	if err := validateAgentModelConfig(req.ModelConfig); err != nil {
		return err
	}
	// Rerank 模型类型枚举（提供时必须是 rerank 类型）
	if req.RerankConfig != nil && req.RerankConfig.ModelType != "" && !agentRerankTypeEnum[req.RerankConfig.ModelType] {
		return fmt.Errorf("rerankConfig modelType %q invalid, expected rerank/multimodal-rerank", req.RerankConfig.ModelType)
	}
	// 视觉配置图片数量非负
	if req.VisionConfig != nil && req.VisionConfig.PicNum < 0 {
		return fmt.Errorf("visionConfig picNum must be >= 0")
	}
	// 记忆轮次范围
	if req.MemoryConfig != nil {
		if req.MemoryConfig.MaxHistoryLength < agentMemoryRoundsMin || req.MemoryConfig.MaxHistoryLength > agentMemoryRoundsMax {
			return fmt.Errorf("memory maxHistoryLength out of range [%d,%d]", agentMemoryRoundsMin, agentMemoryRoundsMax)
		}
	}
	// 追问配置：历史轮次范围、提示词长度、追问模型采样参数范围
	if req.RecommendConfig != nil {
		if req.RecommendConfig.MaxHistory < agentRecommendHistoryMin || req.RecommendConfig.MaxHistory > agentRecommendHistoryMax {
			return fmt.Errorf("recommend maxHistory out of range [%d,%d]", agentRecommendHistoryMin, agentRecommendHistoryMax)
		}
		if utf8.RuneCountInString(req.RecommendConfig.Prompt) > agentRecommendPromptMaxLen {
			return fmt.Errorf("recommend prompt length exceeds %d", agentRecommendPromptMaxLen)
		}
		if err := validateAgentModelConfig(&req.RecommendConfig.ModelConfig); err != nil {
			return err
		}
	}
	// 知识库召回参数（topK/threshold/matchType/权重和）与 id 去重
	if err := validateAgentKnowledgeConfig(req.KnowledgeBaseConfig); err != nil {
		return err
	}
	// 安全护栏：敏感词表数组元素 tableId 必填
	if req.SafetyConfig != nil {
		for i, t := range req.SafetyConfig.Tables {
			if t.TableId == "" {
				return fmt.Errorf("safety tables[%d] tableId required", i)
			}
		}
	}
	return nil
}

// validateAgentModelConfig 校验大模型采样参数范围，仅对请求中出现的 key 生效。
func validateAgentModelConfig(cfg *AppModelConfig) error {
	if cfg == nil || cfg.Config == nil {
		return nil
	}
	m, ok := cfg.Config.(map[string]any)
	if !ok {
		return fmt.Errorf("modelConfig.config must be an object")
	}
	ranges := []struct {
		key      string
		min, max float64
	}{
		{"temperature", 0, 2},
		{"topP", 0, 1},
		{"frequencyPenalty", -2, 2},
		{"presencePenalty", -2, 2},
		{"maxTokens", agentMaxTokensMin, agentMaxTokensMax},
	}
	for _, r := range ranges {
		if err := util.FloatFieldInRange(m, r.key, r.min, r.max); err != nil {
			return err
		}
	}
	return nil
}

// validateAgentKnowledgeConfig 仅在挂载了知识库时校验召回参数。
func validateAgentKnowledgeConfig(cfg *AppKnowledgebaseConfig) error {
	if cfg == nil || len(cfg.Knowledgebases) == 0 {
		return nil
	}
	// 知识库 id 去重（空 id 由 service 层统一过滤，这里只查重复）
	seen := make(map[string]struct{}, len(cfg.Knowledgebases))
	for _, kb := range cfg.Knowledgebases {
		if kb.ID == "" {
			continue
		}
		if _, dup := seen[kb.ID]; dup {
			return fmt.Errorf("duplicate knowledgebase id %q", kb.ID)
		}
		seen[kb.ID] = struct{}{}
	}
	c := cfg.Config
	// topK==0 视为未设置走默认；提供时必须在 (0,10]
	if c.TopK < 0 || c.TopK > agentKnowledgeTopKMax {
		return fmt.Errorf("knowledge topK out of range [1,%d]", agentKnowledgeTopKMax)
	}
	if c.Threshold < 0 || c.Threshold > 1 {
		return fmt.Errorf("knowledge threshold out of range [0,1]")
	}
	if c.MatchType != "" && !agentMatchTypeEnum[c.MatchType] {
		return fmt.Errorf("invalid matchType %q, expected vector/text/mix", c.MatchType)
	}
	// mix + 权重模式下，语义权重 + 关键词权重必须为 1
	if c.MatchType == "mix" && c.PriorityMatch == 1 {
		sum := float64(c.SemanticsPriority) + float64(c.KeywordPriority)
		if math.Abs(sum-1) > agentWeightSumTolerance {
			return fmt.Errorf("semanticsPriority + keywordPriority must equal 1, got %v", sum)
		}
	}
	return nil
}
