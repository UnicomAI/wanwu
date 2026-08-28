package orm

import (
	"context"
	"fmt"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/iam"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AppNoticeReq 应用消息写入入参（发布/改范围/发新版本/删除）。
// oldScope/newScope 已由 bff 归一化为 ScopeXxx 常量。
type AppNoticeReq struct {
	EventID     string // 幂等键：源事件唯一标识
	AppType     string // agent/workflow/chatflow/rag/skill（SubjectType 取此值）
	AppID       string
	AppName     string
	OrgID       string // 应用归属组织
	OldScope    string // 删除分支（newScope=""）的下线受众范围；发布分支不使用
	NewScope    string
	Version     string // 非空 → common 收「版本升级」；空 → common 不发
	SenderID    string
	SenderOrgID string
}

// ModelNoticeReq 模型消息写入入参（导入/删除/停用启用）。
type ModelNoticeReq struct {
	EventID     string // 幂等键：源事件唯一标识
	ModelID     string
	ModelName   string
	OrgID       string // 模型归属组织
	OldScope    string
	NewScope    string
	SenderID    string
	SenderOrgID string
}

// KnowledgeNoticeReq 知识库消息写入入参（共享/取消共享/改权限/删除）。
type KnowledgeNoticeReq struct {
	EventID       string // 幂等键：源事件唯一标识
	KnowledgeID   string
	KnowledgeName string
	Gained        []iam.UserOrg // → shared
	Lost          []iam.UserOrg // → unshared / deleted（资源删除时）
	Changed       []iam.UserOrg // → perm_changed（权限变更）
	SenderID      string
	SenderOrgID   string
	// ChangedDetail perm_changed 分支的补充信息：变更后的权限名/状态描述
	ChangedDetail string
	// Variant 文案变体（一段式）：shared/unshared/deleted/perm_changed；仅 lost 维度用 deleted 区分删除
	Variant string
	Actions []Action
}

// CreateNoticeResult 写入结果
type CreateNoticeResult struct {
	MessagesCreated     int32
	AudienceRowsCreated int32
	IgnoredCount        int32
}

// pendingMessage 一条待写入的本体 + 它的受众规则
type pendingMessage struct {
	spec         *audienceSpec
	eventVariant string
	subjectType  string
	title        string
	content      string
	actions      []Action
	category     int8
}

// resourceTypeLabels 资源类型的中文文案标签
var resourceTypeLabels = map[string]string{
	MsgTypeAgent:     "智能体",
	MsgTypeWorkflow:  "工作流",
	MsgTypeChatflow:  "对话流",
	MsgTypeRag:       "知识问答",
	MsgTypeSkill:     "Skill",
	MsgTypeModel:     "模型",
	MsgTypeKnowledge: "知识库",
}

func resourceLabel(resourceType string) string {
	if label, ok := resourceTypeLabels[resourceType]; ok {
		return label
	}
	return "资源"
}

// CreateAppNotice 应用消息：发布 / 发新版本 / 删除。
//
//   - 发新版本（newScope 非空 + version 非空）→ 整个新范围收「已上线」（无差分、无排除）
//   - 应用删除（newScope=""）→ 删除前 oldScope 全体收「已下线」（应用删除仍发下线；
//     范围变更导致的收件人减少不发下线）
//   - 只改范围不发版本（newScope 非空 + version=""）→ 不发
//
// oldScope 仅删除分支（newScope=""）使用，发布分支不参与受众计算。
func (c *Client) CreateAppNotice(ctx context.Context, req *AppNoticeReq) (*CreateNoticeResult, *errs.Status) {
	// 应用删除：oldScope 全体收「已下线」
	if req.NewScope == "" {
		spec := scopeSpec(req.OldScope, req.OrgID)
		if spec == nil {
			// private/空范围 → 无受众，跳过
			return c.commitNotice(ctx, req.EventID, req.SenderID, req.SenderOrgID, nil, 0)
		}
		p := c.buildScopeMessage(req.AppName, req.AppType, req.AppID, "", req.AppType, spec, model.EventVariantOffline)
		return c.commitNotice(ctx, req.EventID, req.SenderID, req.SenderOrgID, []*pendingMessage{p}, 0)
	}
	// 只改范围不发版本 → 不发
	if req.Version == "" {
		return c.commitNotice(ctx, req.EventID, req.SenderID, req.SenderOrgID, nil, 0)
	}
	spec := scopeSpec(req.NewScope, req.OrgID)
	if spec == nil {
		// private/空范围 → 无受众，跳过
		return c.commitNotice(ctx, req.EventID, req.SenderID, req.SenderOrgID, nil, 0)
	}
	p := c.buildScopeMessage(req.AppName, req.AppType, req.AppID, req.Version, req.AppType, spec, model.EventVariantOnline)
	return c.commitNotice(ctx, req.EventID, req.SenderID, req.SenderOrgID, []*pendingMessage{p}, 0)
}

// CreateModelNotice 模型消息：导入/启用 → 范围收「上线」；删除/停用 → 范围收「下线」。
func (c *Client) CreateModelNotice(ctx context.Context, req *ModelNoticeReq) (*CreateNoticeResult, *errs.Status) {
	var spec *audienceSpec
	var variant string
	if req.NewScope == "" {
		// 删除/停用：旧范围全体收「下线」
		spec = scopeSpec(req.OldScope, req.OrgID)
		variant = model.EventVariantOffline
	} else {
		// 导入/启用：新范围全体收「上线」
		spec = scopeSpec(req.NewScope, req.OrgID)
		variant = model.EventVariantOnline
	}
	if spec == nil {
		// 私有/空范围 → 无受众，跳过
		return c.commitNotice(ctx, req.EventID, req.SenderID, req.SenderOrgID, nil, 0)
	}
	p := c.buildScopeMessage(req.ModelName, MsgTypeModel, req.ModelID, "", MsgTypeModel, spec, variant)
	return c.commitNotice(ctx, req.EventID, req.SenderID, req.SenderOrgID, []*pendingMessage{p}, 0)
}

// buildScopeMessage 范围模板：应用 + 模型共用。参数化入参——不接收整个 req，
// 各 rpc 从自己的 req 提取字段传入；resourceType 区分 app/model（cta、action 差异）。
// version 仅应用发布非空（online 文案带版本号）；模型/删除传 ""（无版本）。
func (c *Client) buildScopeMessage(name, resourceType, resourceID, version, subjectType string, spec *audienceSpec, variant string) *pendingMessage {
	label := resourceLabel(resourceType)
	p := &pendingMessage{
		spec:         spec,
		eventVariant: variant,
		subjectType:  subjectType,
		category:     model.NoticeCategoryProductService,
	}
	switch variant {
	case model.EventVariantOnline:
		p.title = name + " 已上线"
		// 模型上线文案与知识库一致用「查看」；应用仍用「立即体验」
		cta := "立即体验"
		if resourceType == MsgTypeModel {
			cta = "查看"
		}
		// 应用发布带版本号（version 已含 v 前缀：智能体《X》v2.0.1已上线）；模型无版本，保持「已上线」
		if version != "" {
			p.content = fmt.Sprintf("%s《%s》%s已上线，点击$${%s}$$", label, name, version, cta)
		} else {
			p.content = fmt.Sprintf("%s《%s》已上线，点击$${%s}$$", label, name, cta)
		}
		p.actions = []Action{resourceAction(resourceType, resourceID)}
	case model.EventVariantOffline:
		// 资源已不可见，无处可跳
		p.title = name + " 已下线"
		p.content = fmt.Sprintf("%s《%s》已下线", label, name)
	}
	return p
}

// resourceAction 构造资源详情页跳转：应用五类跳探索广场详情，模型跳模型体验页
func resourceAction(resourceType, resourceID string) Action {
	if resourceType == MsgTypeModel {
		return Action{
			MsgType:      MsgTypeModel,
			ActionType:   ActionTypeLink,
			ActionParams: map[string]interface{}{"modelId": resourceID},
		}
	}
	return Action{
		MsgType:    resourceType,
		ActionType: ActionTypeLink,
		ActionParams: map[string]interface{}{
			"appId":   resourceID,
			"appType": resourceType,
		},
	}
}

// CreateKnowledgeNotice 知识库消息：受众是调用方持有的原始事实（显式名单），
// 载体为特定用户名单，恒落 notice_message_audience 正向行。
func (c *Client) CreateKnowledgeNotice(ctx context.Context, req *KnowledgeNoticeReq) (*CreateNoticeResult, *errs.Status) {
	var ignored int32
	gained, n, exceeded, err := c.cleanAudiencePairs(ctx, req.Gained, req.SenderID, req.SenderOrgID)
	ignored += n
	if err != nil {
		// IAM 校验失败：整事件 fail-closed，不投递任何分组（杜绝部分成功部分丢失）
		return &CreateNoticeResult{IgnoredCount: ignored}, nil
	}
	if exceeded {
		return &CreateNoticeResult{IgnoredCount: ignored}, nil
	}
	lost, n, exceeded, err := c.cleanAudiencePairs(ctx, req.Lost, req.SenderID, req.SenderOrgID)
	ignored += n
	if err != nil {
		return &CreateNoticeResult{IgnoredCount: ignored}, nil
	}
	if exceeded {
		return &CreateNoticeResult{IgnoredCount: ignored}, nil
	}
	changed, n, exceeded, err := c.cleanAudiencePairs(ctx, req.Changed, req.SenderID, req.SenderOrgID)
	ignored += n
	if err != nil {
		return &CreateNoticeResult{IgnoredCount: ignored}, nil
	}
	if exceeded {
		return &CreateNoticeResult{IgnoredCount: ignored}, nil
	}

	// 一段式：variant 直接承载文案语义（shared/unshared/deleted/perm_changed）。
	// 三个差分维度与文案变体一一对应：
	//   gained  → shared（新增可见）
	//   lost    → unshared（默认，取消共享）/ deleted（资源删除，调用方显式传 req.Variant）
	//   changed → perm_changed（权限/状态变更）
	// 各维度变体互异 ⇒ (event_id, event_variant) 不冲突，即使一次请求 gained+lost 并存。
	var pendings []*pendingMessage
	if spec := specificSpec(gained); spec != nil {
		pendings = append(pendings, c.buildAudienceMessage(req, spec, model.EventVariantShared))
	}
	if spec := specificSpec(lost); spec != nil {
		lossVariant := model.EventVariantUnshared
		if req.Variant == model.EventVariantDeleted {
			lossVariant = model.EventVariantDeleted
		}
		pendings = append(pendings, c.buildAudienceMessage(req, spec, lossVariant))
	}
	if spec := specificSpec(changed); spec != nil {
		pendings = append(pendings, c.buildAudienceMessage(req, spec, model.EventVariantPermChanged))
	}
	return c.commitNotice(ctx, req.EventID, req.SenderID, req.SenderOrgID, pendings, ignored)
}

func (c *Client) buildAudienceMessage(req *KnowledgeNoticeReq, spec *audienceSpec, variant string) *pendingMessage {
	label := resourceLabel(MsgTypeKnowledge)
	p := &pendingMessage{
		spec:         spec,
		eventVariant: variant,
		subjectType:  MsgTypeKnowledge,
		category:     model.NoticeCategoryProductService,
	}
	viewAction := Action{
		MsgType:      MsgTypeKnowledge,
		ActionType:   ActionTypeLink,
		ActionParams: map[string]interface{}{"knowledgeId": req.KnowledgeID},
	}
	// 调用方显式传了 actions 就用它（未来工单等复用本 rpc 时需要自定义跳转）
	useCallerActions := len(req.Actions) > 0

	switch variant {
	case model.EventVariantShared:
		p.title = fmt.Sprintf("%s《%s》已共享给你", label, req.KnowledgeName)
		p.content = p.title + "，点击$${查看}$$"
		p.actions = []Action{viewAction}
	case model.EventVariantDeleted:
		p.title = fmt.Sprintf("%s《%s》已被删除", label, req.KnowledgeName)
		p.content = p.title
	case model.EventVariantUnshared:
		p.title = fmt.Sprintf("《%s》的共享已取消", req.KnowledgeName)
		p.content = p.title
	case model.EventVariantPermChanged:
		p.title = fmt.Sprintf("你对《%s》的权限已变更为%s", req.KnowledgeName, req.ChangedDetail)
		p.content = p.title
	default:
		// 空/未知 variant 兜底：通用「有变更」文案，避免空消息进列表（best-effort 不吞消息）
		p.title = fmt.Sprintf("%s《%s》有变更", label, req.KnowledgeName)
		p.content = p.title
	}
	if useCallerActions {
		p.actions = req.Actions
	}
	return p
}

// commitNotice 执行写入顺序协议。
//
//  1. MessageID 预生成（雪花，不依赖 INSERT 自增回填）
//  2. 受众行先写（org/audience，靠各表 uk 幂等重写）——
//     本体行尚不存在时这些行对任何查询绝对不可见（«vis» 全部由 notice_messages 驱动）；
//     中断残留的孤儿行不可见、永久保留（不做后台清理）
//  3. 本体最后写：一次事件的本体在同一事务内 INSERT，
//     提交瞬间消息带完整受众原子可见，"要么都收到、要么都没有"
//
// ⚠ 此顺序是原子可见性协议，禁止改回"先父后子"：先写本体会让消息在受众行落完前
// 短暂可见性不完整（如组织行未落时组织内成员看不到本应收到的消息）。
//
// 重试按"本体是否已存在"区分：存在则跳过（上次已成功），不存在则名单幂等重写后补本体。
// 已有本体的消息禁止再无条件改写其名单。
func (c *Client) commitNotice(ctx context.Context, eventID, senderID, senderOrgID string, pendings []*pendingMessage, ignored int32) (*CreateNoticeResult, *errs.Status) {
	result := &CreateNoticeResult{IgnoredCount: ignored}
	if len(pendings) == 0 {
		// 受众剔除后为空集 / 范围无变化 / 名单超限：跳过整条消息
		return result, nil
	}
	if eventID == "" {
		return nil, toErrStatus("notice_create", "eventId empty")
	}

	// 重试幂等：本体已存在说明上次已成功，直接跳过（不再改写名单）
	var existing int64
	if err := c.db.WithContext(ctx).Model(&model.NoticeMessage{}).
		Where("event_id = ?", eventID).Count(&existing).Error; err != nil {
		return nil, toErrStatus("notice_create", err.Error())
	}
	if existing > 0 {
		log.Infof("notice: event %v already committed (%v messages), skip", eventID, existing)
		return result, nil
	}

	now := time.Now().UnixMilli()
	bodies := make([]*model.NoticeMessage, 0, len(pendings))
	for _, p := range pendings {
		messageID := util.MustI64(util.NewID()) // 步骤 1：预生成
		actionsJSON, dropped := sanitizeActions(p.actions)
		if dropped > 0 {
			// 校验不过的 action 被剔除，对应 $${}$$ 占位符由前端降级为纯文本
			log.Errorf("notice: %v action(s) dropped by schema check (event=%v, variant=%v)", dropped, eventID, p.eventVariant)
		}
		bodies = append(bodies, &model.NoticeMessage{
			ID:           messageID,
			EventID:      eventID,
			EventVariant: p.eventVariant,
			SubjectType:  p.subjectType,
			CreatedAt:    now,
			UpdatedAt:    now,
			Category:     p.category,
			Title:        p.title,
			Content:      p.content,
			AudienceType: p.spec.AudienceType,
			Actions:      actionsJSON,
			SenderID:     senderID,
			SenderOrgID:  senderOrgID,
		})

		// 步骤 2：名单行先写
		rows, err := c.writeAudienceRows(ctx, messageID, now, p.spec)
		if err != nil {
			return nil, toErrStatus("notice_create", err.Error())
		}
		result.AudienceRowsCreated += rows
	}

	// 步骤 3：本体最后写，同一事务提交一次事件的全部差分本体。
	// DoNothing 让 (event_id, event_variant) 冲突变成静默跳过而非报错——
	// 并发重试冲过前面的存在性预检时，"重复插入视作已成功"（幂等），
	// 且不依赖各方言的重复键错误码。
	tx := c.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "event_id"}, {Name: "event_variant"}},
		DoNothing: true,
	}).Create(&bodies)
	if tx.Error != nil {
		return nil, toErrStatus("notice_create", tx.Error.Error())
	}
	result.MessagesCreated = int32(tx.RowsAffected)
	return result, nil
}

// writeAudienceRows 写受众行（幂等重写）：
//   - Orgs（AT=2 组织内）→ notice_message_orgs 正向组织行（不计入对外 audienceRowsCreated）
//   - Pairs（AT=3 名单）→ notice_message_audience 正向名单行（计入 audienceRowsCreated）
//
// 排除名单概念已随产品设计整体移除，无反向写入分支。
func (c *Client) writeAudienceRows(ctx context.Context, messageID, now int64, spec *audienceSpec) (int32, error) {
	// created_at = created_at 的自赋值让重复写成为无副作用的幂等操作；
	// 仅 orgs 分支使用（notice_message_orgs.created_at 读侧水位截断仍存在）
	keepCreatedAt := clause.Assignments(map[string]interface{}{"created_at": gorm.Expr("created_at")})

	var audienceRows int32

	if len(spec.Orgs) > 0 {
		rows := make([]model.MessageOrg, 0, len(spec.Orgs))
		for _, orgID := range spec.Orgs {
			rows = append(rows, model.MessageOrg{
				MessageID: messageID, OrgID: orgID, CreatedAt: now,
			})
		}
		if err := c.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "message_id"}, {Name: "org_id"},
			},
			DoUpdates: keepCreatedAt,
		}).Create(&rows).Error; err != nil {
			return 0, err
		}
	}

	if len(spec.Pairs) > 0 {
		rows := make([]model.MessageAudience, 0, len(spec.Pairs))
		for _, p := range spec.Pairs {
			rows = append(rows, model.MessageAudience{
				MessageID: messageID, UserID: p.UserID, OrgID: p.OrgID,
			})
		}
		// audience 无 created_at 列，冲突时无列可更新，DoNothing 幂等跳过
		if err := c.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "user_id"}, {Name: "org_id"}, {Name: "message_id"},
			},
			DoNothing: true,
		}).Create(&rows).Error; err != nil {
			return 0, err
		}
		audienceRows = int32(len(rows))
	}
	return audienceRows, nil
}
