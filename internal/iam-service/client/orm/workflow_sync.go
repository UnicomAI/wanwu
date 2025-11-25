package orm

import (
	"context"
	"fmt"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/pkg/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// WorkflowSpaceUser represents the space_user table in the workflow database
type WorkflowSpaceUser struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement"`
	SpaceID   uint64 `gorm:"column:space_id;not null;default:0;index:uniq_space_user,unique"`
	UserID    uint64 `gorm:"column:user_id;not null;default:0;index:idx_user_id;index:uniq_space_user,unique"`
	RoleType  int    `gorm:"column:role_type;not null;default:3"` // 1=owner, 2=admin, 3=member
	CreatedAt uint64 `gorm:"column:created_at;not null;default:0"`
	UpdatedAt uint64 `gorm:"column:updated_at;not null;default:0"`
}

func (WorkflowSpaceUser) TableName() string {
	return "space_user"
}

var workflowDB *gorm.DB

// InitWorkflowDB initializes the connection to the workflow database (opencoze)
func InitWorkflowDB(mysqlAddress, mysqlPassword string) error {
	dsn := fmt.Sprintf("root:%s@tcp(%s)/opencoze?charset=utf8mb4&parseTime=true&loc=Local",
		mysqlPassword,
		mysqlAddress,
	)
	
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: nil, // Use default logger
	})
	if err != nil {
		return fmt.Errorf("failed to connect to workflow database: %w", err)
	}

	// Set connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	workflowDB = db
	log.Infof("Successfully connected to workflow database (opencoze)")
	return nil
}

// SyncUserToWorkflowSpace syncs a user to the workflow space_user table
// roleType: 1=owner, 2=admin, 3=member
func (c *Client) SyncUserToWorkflowSpace(ctx context.Context, userID, spaceID uint32, roleType int) *errs.Status {
	if workflowDB == nil {
		log.Warnf("Workflow database not initialized, skipping space_user sync for user %d, space %d", userID, spaceID)
		return nil // Don't fail the main operation if workflow sync is not available
	}

	nowMs := uint64(time.Now().UnixMilli())
	
	spaceUser := &WorkflowSpaceUser{
		SpaceID:   uint64(spaceID),
		UserID:    uint64(userID),
		RoleType:  roleType,
		CreatedAt: nowMs,
		UpdatedAt: nowMs,
	}

	// Use ON DUPLICATE KEY UPDATE logic via GORM
	result := workflowDB.WithContext(ctx).
		Where("space_id = ? AND user_id = ?", spaceID, userID).
		Assign(map[string]interface{}{
			"role_type":  roleType,
			"updated_at": nowMs,
		}).
		FirstOrCreate(spaceUser)

	if result.Error != nil {
		log.Errorf("Failed to sync user %d to workflow space %d: %v", userID, spaceID, result.Error)
		return toErrStatus("iam_workflow_space_user_sync", fmt.Sprintf("user_id=%d, space_id=%d, error=%v", userID, spaceID, result.Error))
	}

	log.Infof("Successfully synced user %d to workflow space %d with role %d", userID, spaceID, roleType)
	return nil
}

// RemoveUserFromWorkflowSpace removes a user from the workflow space_user table
func (c *Client) RemoveUserFromWorkflowSpace(ctx context.Context, userID, spaceID uint32) *errs.Status {
	if workflowDB == nil {
		log.Warnf("Workflow database not initialized, skipping space_user removal for user %d, space %d", userID, spaceID)
		return nil
	}

	result := workflowDB.WithContext(ctx).
		Where("space_id = ? AND user_id = ?", spaceID, userID).
		Delete(&WorkflowSpaceUser{})

	if result.Error != nil {
		log.Errorf("Failed to remove user %d from workflow space %d: %v", userID, spaceID, result.Error)
		return toErrStatus("iam_workflow_space_user_remove", fmt.Sprintf("user_id=%d, space_id=%d, error=%v", userID, spaceID, result.Error))
	}

	log.Infof("Successfully removed user %d from workflow space %d", userID, spaceID)
	return nil
}

