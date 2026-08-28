package orm

import (
	"strings"

	"github.com/UnicomAI/wanwu/internal/operate-service/client/model"
)

// 本文件是 «vis»（可见性）与 «未读» 谓词的唯一来源，被全部 5 个读查询共用。
// 任何分支改动都只能改这里，三份拷贝一旦漂移可见性语义就会在不同接口间不一致。
//
// «vis» 按 AudienceType 拆成三个互斥分支 UNION ALL，各走各的索引，候选集从全表缩到
// "该用户真实可见"：
//
//	分支 A · 组织内（AT=2）：由 notice_message_orgs 的 (org_id, created_at) 前缀驱动范围扫，
//	                          join 回本体；joinedAt 屏蔽 + 排除本人
//	分支 B · 全局（AT=4）  ：joinedAt 屏蔽 + 排除本人（全员可见，无排除——排除语义已整体移除）
//	分支 C · 特定用户（AT IN (1,3)）：反向驱动，由 notice_message_audience(user_id,org_id) 前缀
//	                                 先取 message_id 再半连接回本体；冻结投递，不受 joinedAt 屏蔽
//
// 反连接保持 NOT EXISTS 写法（MySQL 8.0.17+ 自动做反连接变换），执行算法交给优化器
// 按基数选择：活跃用户期望嵌套循环覆盖探测，从不读用户期望 hash。实测选错才局部加 hint。
//
// 方言约束：pkg/db 支持 postgres，故这里只用 ANSI 语法（UNION ALL / NOT EXISTS / 行构造器），
// 布尔值走绑定参数而非 `= 1` 字面量，所有 upsert 走 clause.OnConflict 不手写 ON DUPLICATE KEY。

// visOpt 是 «vis» 的构造参数。
type visOpt struct {
	UserID string
	OrgID  string
	// JoinedAt 用户加入当前组织的时间（org_users.created_at）；
	// 新成员不追溯历史消息，仅作用于分支 A/B。iam 不可用时 fail-open 传 0。
	JoinedAt int64
	// Denied 成员关系不存在或已禁用（Exists=false 或 Active=false）。
	// 读侧据此 fail-closed 返回空结果——拒绝伪造 X-Org-Id 读取其他组织通知。
	// 仅 iam gRPC 调用本身故障时才 fail-open（JoinedAt=0 放行），二者不可混淆。
	Denied bool
	// WmAt/WmID 已读水位二元组，无水位行时为 (0,0)
	WmAt int64
	WmID int64
	// OnlyUnread 追加 «未读» 谓词（水位切点 + 无已读/已删状态行）
	OnlyUnread bool
	// SelectCols 各分支的投影列，如 "m.id" / "m.id, m.category, m.created_at"
	SelectCols string
}

// visUnionSQL 返回 «vis» 三分支 UNION ALL 的 SQL 与绑定参数。
// 调用方负责拼外层 SELECT / JOIN / WHERE / ORDER BY / LIMIT。
func visUnionSQL(o visOpt) (string, []interface{}) {
	cols := o.SelectCols
	if cols == "" {
		cols = "m.id"
	}

	var sb strings.Builder
	args := make([]interface{}, 0, 32)

	// --- 分支 A · 组织内（AT=2）---
	// 由 notice_message_orgs 驱动：org_id 前缀 + created_at 范围扫（read-diffusion 直读该组织消息），
	// join 本体取宽列；created_at 随行冗余存储、与本体同写同时刻，org 索引上可做水位范围截断。
	sb.WriteString("SELECT " + cols + " FROM notice_message_orgs o JOIN notice_messages m ON m.id = o.message_id WHERE o.org_id = ?")
	args = append(args, o.OrgID)
	sb.WriteString(" AND m.audience_type = ?")
	args = append(args, model.AudienceTypeOrg)
	sb.WriteString(" AND m.created_at >= ?")
	args = append(args, o.JoinedAt)
	sb.WriteString(" AND NOT (m.sender_id = ? AND m.sender_org_id = ?)")
	args = append(args, o.UserID, o.OrgID)
	if o.OnlyUnread {
		args = append(args, appendUnreadSQL(&sb, o, "o.created_at")...)
	}

	sb.WriteString(" UNION ALL ")

	// --- 分支 B · 全局（AT=4）---
	// 全员可见，无排除（排除语义已随产品设计整体移除）。
	sb.WriteString("SELECT " + cols + " FROM notice_messages m WHERE m.audience_type = ?")
	args = append(args, model.AudienceTypeGlobal)
	sb.WriteString(" AND m.created_at >= ?")
	args = append(args, o.JoinedAt)
	sb.WriteString(" AND NOT (m.sender_id = ? AND m.sender_org_id = ?)")
	args = append(args, o.UserID, o.OrgID)
	if o.OnlyUnread {
		args = append(args, appendUnreadSQL(&sb, o, "m.created_at")...)
	}

	sb.WriteString(" UNION ALL ")

	// --- 分支 C · 特定用户 + 点对点（AT IN (1,3)）---
	// 反向驱动：必须 EXISTS 半连接，禁止 INNER JOIN（否则本体成为驱动表，候选集退化为全表）
	sb.WriteString("SELECT " + cols + " FROM notice_messages m WHERE m.audience_type IN (?, ?)")
	args = append(args, model.AudienceTypePrivate, model.AudienceTypeSpecific)
	sb.WriteString(" AND EXISTS (SELECT 1 FROM notice_message_audience a" +
		" WHERE a.message_id = m.id AND a.user_id = ? AND a.org_id = ?)")
	args = append(args, o.UserID, o.OrgID)
	if o.OnlyUnread {
		// 分支 C 不加冗余下界：它由 notice_message_audience(user_id,org_id) 前缀驱动，
		// created_at 范围截断对它无收益（收益只在按 (AT,…,created_at,…) 复合索引范围扫的 A/B 支）
		args = append(args, appendUnreadSQL(&sb, o, "")...)
	}

	return sb.String(), args
}

// appendUnreadSQL 追加 «未读» 谓词，返回对应的绑定参数。
//
// boundCol 控制是否加冗余下界 `<col> >= :wm_at`：
//   - 分支 A 传 "o.created_at"（org 索引列，让水位截断作用在 notice_message_orgs 的范围扫上）
//   - 分支 B 传 "m.created_at"（本体索引列）
//   - 分支 C 传 ""（前缀驱动无收益）
//
// ⚠ 冗余下界对 A/B 支不可省略：行构造器的列序与索引键序不对齐，range 优化器不保证能从
// 行构造器推导出范围截断；少了它，活跃画像会从"只扫水位后的百来条"退化成全候选扫描
// （5~20ms → 50~150ms）。
func appendUnreadSQL(sb *strings.Builder, o visOpt, boundCol string) []interface{} {
	args := make([]interface{}, 0, 7)
	if boundCol != "" {
		sb.WriteString(" AND " + boundCol + " >= ?")
		args = append(args, o.WmAt)
	}
	// 水位是排序序列上的精确切点，与 ORDER BY 同键
	sb.WriteString(" AND (m.created_at, m.id) > (?, ?)")
	args = append(args, o.WmAt, o.WmID)
	sb.WriteString(" AND NOT EXISTS (SELECT 1 FROM notice_user_status s" +
		" WHERE s.message_id = m.id AND s.user_id = ? AND s.org_id = ?" +
		" AND (s.is_read = ? OR s.is_deleted = ?))")
	args = append(args, o.UserID, o.OrgID, true, true)
	return args
}
