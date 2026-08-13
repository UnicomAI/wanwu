package orm

import (
	"context"
	"errors"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// userStatusConflictCols notice_user_status 的幂等键（uk_notice_user_status_user_org_msg）
var userStatusConflictCols = []clause.Column{
	{Name: "user_id"}, {Name: "org_id"}, {Name: "message_id"},
}

// ReadNotice 单条已读：可见性校验后 upsert notice_user_status。
// 重复对已读消息调用是幂等的。
func (c *Client) ReadNotice(ctx context.Context, userID, orgID string, messageID int64) *errs.Status {
	o, status := c.loadVisOpt(ctx, userID, orgID)
	if status != nil {
		return status
	}
	if o.Denied {
		return nil // 非该组织成员 / 已禁用：静默跳过（fail-closed，与不可见语义一致）
	}
	visible, err := filterVisibleMessageIDs(c.db.WithContext(ctx), o, []int64{messageID})
	if err != nil {
		return toErrStatus("notice_read", err.Error())
	}
	if len(visible) == 0 {
		return nil // 不可见静默跳过（防越权、防枚举）
	}
	// 单行 upsert 本身是原子的，无需显式事务。
	// IsRead 与 IsDeleted 正交：只更新已读两列，不碰删除状态
	if err := c.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   userStatusConflictCols,
		DoUpdates: clause.AssignmentColumns([]string{"is_read", "read_at"}),
	}).Create(&model.UserStatus{
		UserID:    userID,
		OrgID:     orgID,
		MessageID: messageID,
		IsRead:    true,
		ReadAt:    time.Now().UnixMilli(),
	}).Error; err != nil {
		return toErrStatus("notice_read", err.Error())
	}
	return nil
}

// DeleteNotice 批量软删：单事务内先做可见性过滤，
// 再一条多值 upsert 置 IsDeleted。全部成功或整体回滚，不存在部分删除。
// 返回实际删除条数（不可见/不存在的 ID 静默跳过、不计入）。
func (c *Client) DeleteNotice(ctx context.Context, userID, orgID string, messageIDs []int64) (int32, *errs.Status) {
	if len(messageIDs) == 0 {
		return 0, nil
	}
	// 可见性过滤放在事务外：它内含一次 iam grpc 点查，不该被 DB 事务包住
	o, status := c.loadVisOpt(ctx, userID, orgID)
	if status != nil {
		return 0, status
	}
	if o.Denied {
		return 0, nil // 非该组织成员 / 已禁用：无可删消息（fail-closed）
	}

	var affected int32
	status = c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		visible, err := filterVisibleMessageIDs(tx, o, messageIDs)
		if err != nil {
			return toErrStatus("notice_delete", err.Error())
		}
		if len(visible) == 0 {
			return nil
		}
		now := time.Now().UnixMilli()
		rows := make([]model.UserStatus, 0, len(visible))
		for _, id := range visible {
			rows = append(rows, model.UserStatus{
				UserID:    userID,
				OrgID:     orgID,
				MessageID: id,
				IsDeleted: true,
				DeletedAt: now,
			})
		}
		// IsRead 与 IsDeleted 正交：只更新删除两列，不覆盖已读状态
		if err := tx.Clauses(clause.OnConflict{
			Columns:   userStatusConflictCols,
			DoUpdates: clause.AssignmentColumns([]string{"is_deleted", "deleted_at"}),
		}).Create(&rows).Error; err != nil {
			return toErrStatus("notice_delete", err.Error())
		}
		affected = int32(len(visible))
		return nil
	})
	if status != nil {
		return 0, status
	}
	return affected, nil
}

// ReadAllNotice 一键已读：作用于当前组织上下文的全部未读，
// 不分类别，与列表筛选无关。
//
// 不逐条写 notice_user_status（N 条未读 = N 次写会退化到数秒级），而是走水位线：
//  1. 从同一行取快照边界（ORDER BY created_at DESC, id DESC LIMIT 1）——
//     禁止分列 MAX(created_at)/MAX(id) 拼伪游标，那会造出一个从未存在过的边界；
//  2. 事务内 SELECT ... FOR UPDATE 锁水位行；
//  3. 成对字典序比较后成对更新，水位只前进——重试天然幂等、并发不倒退。
//
// 已知限制（已拍板接受）：请求瞬间存在 ≤1s 的事务晚提交窗口，
// 边界上的个别消息可能被一并标为已读（消息仍在列表可见）。
func (c *Client) ReadAllNotice(ctx context.Context, userID, orgID string) *errs.Status {
	o, status := c.loadVisOpt(ctx, userID, orgID)
	if status != nil {
		return status
	}
	if o.Denied {
		return nil // 非该组织成员 / 已禁用：无未读可清（fail-closed）
	}
	o.OnlyUnread = true
	o.SelectCols = "m.id, m.created_at"
	inner, args := visUnionSQL(o)

	// 步骤 1：同一行取快照边界
	var boundary struct {
		ID        int64
		CreatedAt int64
	}
	err := c.db.WithContext(ctx).
		Raw("SELECT t.id AS id, t.created_at AS created_at FROM ("+inner+") t"+
			" ORDER BY t.created_at DESC, t.id DESC LIMIT 1", args...).
		Scan(&boundary).Error
	if err != nil {
		return toErrStatus("notice_read_all", err.Error())
	}
	if boundary.ID == 0 && boundary.CreatedAt == 0 {
		return nil // 无未读，直接返回
	}

	// 步骤 2+3：锁行 + 成对字典序单调更新
	return c.transaction(ctx, func(tx *gorm.DB) *errs.Status {
		var wm model.ReadWatermark
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND org_id = ?", userID, orgID).
			First(&wm).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return toErrStatus("notice_read_all", err.Error())
			}
			if createErr := tx.Create(&model.ReadWatermark{
				UserID:      userID,
				OrgID:       orgID,
				WatermarkAt: boundary.CreatedAt,
				WatermarkID: boundary.ID,
			}).Error; createErr != nil {
				return toErrStatus("notice_read_all", createErr.Error())
			}
			return nil
		}
		if !afterWatermark(boundary.CreatedAt, boundary.ID, wm.WatermarkAt, wm.WatermarkID) {
			return nil // 水位只前进
		}
		if err := tx.Model(&model.ReadWatermark{}).
			Where("user_id = ? AND org_id = ?", userID, orgID).
			Updates(map[string]interface{}{
				"watermark_at": boundary.CreatedAt,
				"watermark_id": boundary.ID,
				"updated_at":   time.Now().UnixMilli(),
			}).Error; err != nil {
			return toErrStatus("notice_read_all", err.Error())
		}
		return nil
	})
}
