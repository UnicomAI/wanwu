package orm

import (
	"context"
	"errors"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	"gorm.io/gorm"
)

// NoticeUnreadCount 未读总数 + 分类角标
type NoticeUnreadCount struct {
	Total int32
	// ByCategory key 为 category（1 公告 / 2 产品服务 / 3 工单）；零值类别也要返回
	ByCategory map[int32]int32
}

// ListNoticeReq 整页列表入参（校验由 bff 参数层保证：keyword ≤50、pageSize ≤100）
type ListNoticeReq struct {
	UserID     string
	OrgID      string
	Category   int32 // 0 全部
	OnlyUnread bool
	Keyword    string
	PageNo     int32
	PageSize   int32
}

// NoticeItem 列表项
type NoticeItem struct {
	MessageID  int64
	Title      string
	Category   int32
	Content    string
	Actions    string // 原始 JSON，由 grpc 层解析为 []Action
	ReceivedAt int64
	IsRead     bool
}

// loadVisOpt 读路径前置元数据：每请求 2 次点查（joinedAt + 水位）。
//
// 成员关系判定分三态，fail-open 与 fail-closed 不可混淆：
//   - iam gRPC 调用失败 → fail-open（joinedAt=0 放行 + 告警）。代价仅为新成员短暂看到
//     加入前通知，优于读路径整体不可用。
//   - 调用成功但 Exists=false（用户不属于该组织）或 Active=false（成员关系被禁用）
//     → fail-closed（Denied=true）。读侧据此返回空结果，拒绝伪造 X-Org-Id 越权读取
//     其他组织通知。bff 全仓未校验 X-Org-Id 与用户归属，此处是消息中心自己的越权防线。
func (c *Client) loadVisOpt(ctx context.Context, userID, orgID string) (visOpt, *errs.Status) {
	o := visOpt{UserID: userID, OrgID: orgID}

	membership, err := c.iam.GetUserOrgMembership(ctx, userID, orgID)
	if err != nil {
		log.Errorf("notice: get membership (uid=%v, oid=%v) failed, fail-open with joinedAt=0: %v", userID, orgID, err)
	} else if membership.Exists {
		if !membership.Active {
			// 成员关系被禁用：拒绝，不追溯历史
			o.Denied = true
		} else {
			o.JoinedAt = membership.JoinedAt
		}
	} else {
		// 用户不属于该组织：拒绝（fail-closed，防伪造 X-Org-Id 越权）
		o.Denied = true
	}

	wmAt, wmID, status := c.getWatermark(ctx, userID, orgID)
	if status != nil {
		return o, status
	}
	o.WmAt, o.WmID = wmAt, wmID
	return o, nil
}

// getWatermark 读已读水位二元组；无行等价 (0,0)
func (c *Client) getWatermark(ctx context.Context, userID, orgID string) (int64, int64, *errs.Status) {
	var wm model.ReadWatermark
	err := c.db.WithContext(ctx).
		Where("user_id = ? AND org_id = ?", userID, orgID).
		First(&wm).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, nil
		}
		return 0, 0, toErrStatus("notice_watermark_get", err.Error())
	}
	return wm.WatermarkAt, wm.WatermarkID, nil
}

// GetNoticeUnreadCount 未读总数 + 分类角标。
// 三分支 UNION ALL 各带 «未读» 谓词，外层按 category 分组计数。
//
// 最终 SQL：
//
//	SELECT t.category AS category, COUNT(*) AS n
//	FROM (
//	    -- 分支 A · 组织内（AT=2）：由 notice_message_orgs 驱动（org_id 前缀范围扫）+ joinedAt 屏蔽 + 排除本人
//	    SELECT m.id AS id, m.category AS category
//	    FROM notice_message_orgs o JOIN notice_messages m ON m.id = o.message_id
//	    WHERE o.org_id = ?                            -- orgID
//	      AND m.audience_type = 2                     -- 冗余（org 行只属于 AT=2）
//	      AND m.created_at >= ?                       -- joinedAt（新成员不追溯）
//	      AND NOT (m.sender_id = ? AND m.sender_org_id = ?)   -- 排除本人 uid/oid
//	      AND o.created_at >= ?                       -- 冗余下界 wmAt（作用 org 索引范围扫，勿省）
//	      AND (m.created_at, m.id) > (?, ?)           -- 水位切点 (wmAt, wmID)
//	      AND NOT EXISTS (SELECT 1 FROM notice_user_status s
//	                      WHERE s.message_id = m.id AND s.user_id = ? AND s.org_id = ?
//	                        AND (s.is_read = ? OR s.is_deleted = ?))
//	    UNION ALL
//	    -- 分支 B · 全局（AT=4）：全员可见，无排除
//	    SELECT m.id AS id, m.category AS category
//	    FROM notice_messages m
//	    WHERE m.audience_type = 4
//	      AND m.created_at >= ?                       -- joinedAt
//	      AND NOT (m.sender_id = ? AND m.sender_org_id = ?)
//	      AND m.created_at >= ?                       -- 冗余下界 wmAt
//	      AND (m.created_at, m.id) > (?, ?)
//	      AND NOT EXISTS (SELECT 1 FROM notice_user_status s ...)
//	    UNION ALL
//	    -- 分支 C · 特定用户 + 点对点（AT IN (1,3)）：由 notice_message_audience 前缀反查，冻结投递无 joinedAt 屏蔽
//	    SELECT m.id AS id, m.category AS category
//	    FROM notice_messages m
//	    WHERE m.audience_type IN (1, 3)
//	      AND EXISTS (SELECT 1 FROM notice_message_audience a
//	                  WHERE a.message_id = m.id AND a.user_id = ? AND a.org_id = ?)
//	      AND (m.created_at, m.id) > (?, ?)          -- 分支 C 不加冗余下界（前缀驱动，无收益）
//	      AND NOT EXISTS (SELECT 1 FROM notice_user_status s ...)
//	) t
//	GROUP BY t.category
func (c *Client) GetNoticeUnreadCount(ctx context.Context, userID, orgID string) (*NoticeUnreadCount, *errs.Status) {
	ret := &NoticeUnreadCount{ByCategory: map[int32]int32{
		int32(model.NoticeCategoryAnnouncement):   0,
		int32(model.NoticeCategoryProductService): 0,
		int32(model.NoticeCategoryTicket):         0,
	}}

	o, status := c.loadVisOpt(ctx, userID, orgID)
	if status != nil {
		return nil, status
	}
	if o.Denied {
		return ret, nil // 非该组织成员 / 已禁用：返回零未读（fail-closed）
	}
	o.OnlyUnread = true
	o.SelectCols = "m.id, m.category"

	inner, args := visUnionSQL(o)
	var rows []struct {
		Category int32
		N        int64
	}
	if err := c.db.WithContext(ctx).
		Raw("SELECT t.category AS category, COUNT(*) AS n FROM ("+inner+") t GROUP BY t.category", args...).
		Scan(&rows).Error; err != nil {
		return nil, toErrStatus("notice_unread_count", err.Error())
	}
	for _, r := range rows {
		ret.ByCategory[r.Category] = int32(r.N)
		ret.Total += int32(r.N)
	}
	return ret, nil
}

// ListNotice 整页列表。已读+未读混排，每条带 isRead。
//
// isRead = 有已读状态行 OR 落在水位内——两者在 Go 侧合成，避免把布尔表达式当结果列
// 在不同方言/驱动下的取值差异。
// onlyUnread 时把 «未读» 谓词下推进三个分支（而非留在外层），这样水位截断能作用在
// 复合索引的范围扫上，候选集缩到"水位之后"，这是悬浮面板 5~20ms 的来源。
//
// 最终 SQL（vis 内层 = 上面 GetNoticeUnreadCount 注释的三分支 UNION ALL，此处不再重复）：
//
//	/* ── 路径 A · 无 keyword（延迟关联：先取本页 id 再回表宽列） ── */
//	-- ① COUNT：vis 已包含计数所需列，不再重复关联 notice_messages
//	SELECT COUNT(*)
//	FROM (<vis 三分支 UNION ALL> SELECT m.id AS id, m.created_at AS created_at, m.category AS category) v
//	WHERE NOT EXISTS (
//	    SELECT 1 FROM notice_user_status s
//	    WHERE s.message_id = v.id AND s.user_id = ? AND s.org_id = ? AND s.is_deleted = ?
//	)                                                                                   -- true
//	[AND v.category = ?]                                                               -- 可选
//
//	-- ② 内层：按 created_at/id 排序分页取本页 id（先过滤软删，避免已删消息占位）
//	SELECT v.id
//	FROM (<vis 三分支 UNION ALL>) v
//	LEFT JOIN notice_user_status ps ON ps.message_id = v.id AND ps.user_id = ? AND ps.org_id = ?
//	WHERE (ps.is_deleted IS NULL OR ps.is_deleted = ?)
//	[AND v.category = ?]
//	ORDER BY v.created_at DESC, v.id DESC
//	LIMIT ? OFFSET ?                                                                   -- pageSize/offset
//
//	-- ③ 外层：仅本页 id 回表取宽列 + is_read
//	SELECT m.id AS id, m.title AS title, m.category AS category, m.content AS content,
//	       m.actions AS actions, m.created_at AS received_at,
//	       COALESCE(s.is_read, FALSE) AS row_is_read
//	FROM (<步骤② 内层>) p
//	JOIN notice_messages m ON m.id = p.id
//	LEFT JOIN notice_user_status s ON s.message_id = m.id AND s.user_id = ? AND s.org_id = ?
//	WHERE (s.is_deleted IS NULL OR s.is_deleted = ?)
//	ORDER BY m.created_at DESC, m.id DESC
//
//	/* ── 路径 B · 有 keyword（title/content LIKE 需回表读，延迟关联无收益，直接 JOIN 过滤后分页） ── */
//	SELECT m.id AS id, m.title AS title, m.category AS category, m.content AS content,
//	       m.actions AS actions, m.created_at AS received_at,
//	       COALESCE(s.is_read, FALSE) AS row_is_read
//	FROM (<vis 三分支 UNION ALL>) v
//	JOIN notice_messages m ON m.id = v.id
//	LEFT JOIN notice_user_status s ON s.message_id = m.id AND s.user_id = ? AND s.org_id = ?
//	WHERE (s.is_deleted IS NULL OR s.is_deleted = ?)
//	[AND m.category = ?]
//	AND (m.title LIKE ? OR m.content LIKE ?)                                           -- %kw% 通配
//	ORDER BY m.created_at DESC, m.id DESC
//	LIMIT ? OFFSET ?
func (c *Client) ListNotice(ctx context.Context, req *ListNoticeReq) ([]*NoticeItem, int64, *errs.Status) {
	o, status := c.loadVisOpt(ctx, req.UserID, req.OrgID)
	if status != nil {
		return nil, 0, status
	}
	if o.Denied {
		return nil, 0, nil // 非该组织成员 / 已禁用：返回空列表（fail-closed）
	}
	o.OnlyUnread = req.OnlyUnread
	// 延迟关联：vis 输出 (id, created_at, category)，无 keyword 路径先关联当前用户 notice_user_status 过滤软删除，
	// 再按 created_at/id 排序分页取本页 id，最后回表读取 pageSize 行宽列；category 在内层过滤。
	// keyword 路径保持 JOIN 后过滤再分页，保证两条路径的可见性和分页口径一致。
	// ReadAllNotice("m.id,m.created_at") 与 GetNoticeUnreadCount("m.id,m.category") 的并集。
	o.SelectCols = "m.id, m.created_at, m.category"
	inner, innerArgs := visUnionSQL(o)

	// 外层过滤条件（onlyUnread 已下推进分支，此处不再重复）
	where := " WHERE (s.is_deleted IS NULL OR s.is_deleted = ?)"
	filterArgs := []interface{}{false}
	if req.Category != 0 {
		where += " AND m.category = ?"
		filterArgs = append(filterArgs, req.Category)
	}
	if req.Keyword != "" {
		// 用户输入按普通字符匹配：转义 % _ \ 后在 Go 侧拼通配符，
		// SQL 里只写 LIKE ?（避免 CONCAT 的方言差异；EscapeLike 依赖默认转义符 \，无需 ESCAPE 子句）
		kw := "%" + db.EscapeLike(req.Keyword) + "%"
		where += " AND (m.title LIKE ? OR m.content LIKE ?)"
		filterArgs = append(filterArgs, kw, kw)
	}

	from := " FROM (" + inner + ") v" +
		" JOIN notice_messages m ON m.id = v.id" +
		" LEFT JOIN notice_user_status s ON s.message_id = m.id AND s.user_id = ? AND s.org_id = ?"
	joinArgs := []interface{}{req.UserID, req.OrgID}

	// total：无 keyword 时 vis 已包含 id/category，直接计数并用反连接排除软删除，
	// 避免重复关联 notice_messages 和 LEFT JOIN notice_user_status。keyword 路径仍需
	// 回表匹配 title/content，保持原查询不变。
	var total int64
	var countSQL string
	var countArgs []interface{}
	if req.Keyword == "" {
		countSQL, countArgs = noticeListNoKeywordCountQuery(inner, innerArgs, req)
	} else {
		countSQL = "SELECT COUNT(*)" + from + where
		countArgs = concatArgs(innerArgs, joinArgs, filterArgs)
	}
	if err := c.db.WithContext(ctx).
		Raw(countSQL, countArgs...).
		Scan(&total).Error; err != nil {
		return nil, 0, toErrStatus("notice_list", err.Error())
	}
	if total == 0 {
		return nil, 0, nil
	}

	pageNo, pageSize := req.PageNo, req.PageSize
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (pageNo - 1) * pageSize

	var rows []struct {
		ID         int64
		Title      string
		Category   int32
		Content    string
		Actions    string
		ReceivedAt int64
		RowIsRead  bool
	}
	// 宽列回表投影（外层 SELECT 列表，两条路径共用，列别名与 rows 结构体字段一一对应）
	listSelect := "SELECT m.id AS id, m.title AS title, m.category AS category, m.content AS content," +
		" m.actions AS actions, m.created_at AS received_at," +
		" COALESCE(s.is_read, FALSE) AS row_is_read"

	if req.Keyword != "" {
		// 路径 B：keyword，保持现状（不延迟关联）。title/content LIKE 是中缀匹配，必须回表读
		// 全候选，延迟关联无收益；文档已接受 200~500ms，ngram 全文索引后续再做。
		// 占位符顺序与现状逐字一致：innerArgs + joinArgs + filterArgs + [pageSize, offset]
		listArgs := concatArgs(innerArgs, joinArgs, filterArgs, []interface{}{pageSize, offset})
		if err := c.db.WithContext(ctx).Raw(
			listSelect+from+where+
				" ORDER BY m.created_at DESC, m.id DESC LIMIT ? OFFSET ?",
			listArgs...).Scan(&rows).Error; err != nil {
			return nil, 0, toErrStatus("notice_list", err.Error())
		}
	} else {
		// 路径 A：无 keyword。vis 已输出 (id, created_at, category)，分页前关联当前用户 user_status
		// 过滤软删除，避免已删消息占用分页位置；再按 created_at/id 排序分页取本页 id。
		// category 下推内层用 v.category 过滤，外层只回表本页宽列并读取 notice_user_status.is_read。
		innerPage := "SELECT v.id FROM (" + inner + ") v" +
			" LEFT JOIN notice_user_status ps ON ps.message_id = v.id AND ps.user_id = ? AND ps.org_id = ?" +
			" WHERE (ps.is_deleted IS NULL OR ps.is_deleted = ?)"
		innerPageArgs := innerArgs // concatArgs 不修改输入，innerArgs 仍可被 COUNT 复用
		innerPageArgs = concatArgs(
			innerPageArgs,
			[]interface{}{req.UserID, req.OrgID, false},
		)
		if req.Category != 0 {
			innerPage += " AND v.category = ?"
			innerPageArgs = concatArgs(innerPageArgs, []interface{}{req.Category})
		}
		innerPage += " ORDER BY v.created_at DESC, v.id DESC LIMIT ? OFFSET ?"

		// 外层：仅对本页 id 回表取宽列并读取 notice_user_status.is_read。WHERE 只剩 is_deleted 谓词
		// （category 已在内层应用），不复用 where/filterArgs 避免重复。
		outerFrom := " FROM (" + innerPage + ") p" +
			" JOIN notice_messages m ON m.id = p.id" +
			" LEFT JOIN notice_user_status s ON s.message_id = m.id AND s.user_id = ? AND s.org_id = ?"
		outerWhere := " WHERE (s.is_deleted IS NULL OR s.is_deleted = ?)"

		// 占位符顺序：[innerArgs][UserID,OrgID,false][category?][pageSize,offset][UserID,OrgID][false]
		listArgs := concatArgs(
			innerPageArgs,
			[]interface{}{pageSize, offset},
			[]interface{}{req.UserID, req.OrgID},
			[]interface{}{false},
		)
		if err := c.db.WithContext(ctx).Raw(
			listSelect+outerFrom+outerWhere+
				" ORDER BY m.created_at DESC, m.id DESC",
			listArgs...).Scan(&rows).Error; err != nil {
			return nil, 0, toErrStatus("notice_list", err.Error())
		}
	}

	ret := make([]*NoticeItem, 0, len(rows))
	for _, r := range rows {
		ret = append(ret, &NoticeItem{
			MessageID:  r.ID,
			Title:      r.Title,
			Category:   r.Category,
			Content:    r.Content,
			Actions:    r.Actions,
			ReceivedAt: r.ReceivedAt,
			// 已读 = 有已读状态行 OR 落在水位内（≤ 水位二元组）
			IsRead: r.RowIsRead || !afterWatermark(r.ReceivedAt, r.ID, o.WmAt, o.WmID),
		})
	}
	return ret, total, nil
}

// noticeListNoKeywordCountQuery 构造无 keyword 列表的 total 查询。
// vis 的三个分支互斥且已投影 id/category，因此不需要再次关联消息本体；状态表唯一键
// (user_id, org_id, message_id) 保证 NOT EXISTS 与原 LEFT JOIN 软删除过滤语义一致。
func noticeListNoKeywordCountQuery(inner string, innerArgs []interface{}, req *ListNoticeReq) (string, []interface{}) {
	query := "SELECT COUNT(*) FROM (" + inner + ") v" +
		" WHERE NOT EXISTS (SELECT 1 FROM notice_user_status s" +
		" WHERE s.message_id = v.id AND s.user_id = ? AND s.org_id = ? AND s.is_deleted = ?)"
	args := concatArgs(innerArgs, []interface{}{req.UserID, req.OrgID, true})
	if req.Category != 0 {
		query += " AND v.category = ?"
		args = append(args, req.Category)
	}
	return query, args
}

// afterWatermark 判定 (createdAt, id) 是否严格大于水位二元组（字典序）
func afterWatermark(createdAt, id, wmAt, wmID int64) bool {
	if createdAt != wmAt {
		return createdAt > wmAt
	}
	return id > wmID
}

func concatArgs(groups ...[]interface{}) []interface{} {
	var n int
	for _, g := range groups {
		n += len(g)
	}
	ret := make([]interface{}, 0, n)
	for _, g := range groups {
		ret = append(ret, g...)
	}
	return ret
}

// filterVisibleMessageIDs 可见性过滤：返回入参 ids 中对该 (uid, oid) 真正可见的子集。
// 不可见 ID 静默跳过（防越权 + 防枚举）。
//
// 注意 visOpt 由调用方在**事务外**准备好再传进来：loadVisOpt 内含一次 iam grpc 点查，
// 放进事务会让 DB 事务跨网络调用、平白拉长持锁时间。
func filterVisibleMessageIDs(db *gorm.DB, o visOpt, ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	o.SelectCols = "m.id"
	inner, args := visUnionSQL(o)
	args = append(args, ids)

	var visible []int64
	if err := db.Raw("SELECT v.id FROM ("+inner+") v WHERE v.id IN ?", args...).Scan(&visible).Error; err != nil {
		return nil, err
	}
	return visible, nil
}
