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

// WorkflowUser represents the user table in the workflow database
type WorkflowUser struct {
	ID         uint64 `gorm:"primaryKey;autoIncrement"`
	Name       string `gorm:"column:name;not null;default:''"`
	UniqueName string `gorm:"column:unique_name;not null;default:'';uniqueIndex:uniq_unique_name"`
	Email      string `gorm:"column:email;not null;default:'';uniqueIndex:uniq_email"`
	Password   string `gorm:"column:password;not null;default:''"`
	CreatedAt  uint64 `gorm:"column:created_at;not null;default:0"`
	UpdatedAt  uint64 `gorm:"column:updated_at;not null;default:0"`
}

func (WorkflowUser) TableName() string {
	return "user"
}

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

	// First, ensure the user exists in opencoze.user table
	if err := c.syncUserToWorkflowUserTable(ctx, userID); err != nil {
		log.Warnf("Failed to sync user %d to workflow user table (continuing anyway): %v", userID, err)
		// Don't fail - continue to sync space_user even if user sync fails
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

// syncUserToWorkflowUserTable syncs a user to the opencoze.user table
func (c *Client) syncUserToWorkflowUserTable(ctx context.Context, userID uint32) error {
	if workflowDB == nil {
		return fmt.Errorf("workflow database not initialized")
	}

	// Get user info from IAM database
	var iamUser struct {
		ID    uint32
		Name  string
		Email string
	}

	err := c.db.WithContext(ctx).
		Table("users").
		Select("id, name, email").
		Where("id = ? AND status = ?", userID, true).
		First(&iamUser).Error

	if err != nil {
		return fmt.Errorf("failed to get user from IAM database: %w", err)
	}

	nowMs := uint64(time.Now().UnixMilli())

	// Prepare email - use a default if empty
	email := iamUser.Email
	if email == "" {
		email = fmt.Sprintf("user%d@local.com", userID)
	}

	workflowUser := &WorkflowUser{
		ID:         uint64(userID),
		Name:       iamUser.Name,
		UniqueName: iamUser.Name,
		Email:      email,
		Password:   "", // Empty password - auth is handled by IAM service
		CreatedAt:  nowMs,
		UpdatedAt:  nowMs,
	}

	// Use FirstOrCreate to insert or update
	result := workflowDB.WithContext(ctx).
		Where("id = ?", userID).
		Assign(map[string]interface{}{
			"name":        iamUser.Name,
			"unique_name": iamUser.Name,
			"email":       email,
			"updated_at":  nowMs,
		}).
		FirstOrCreate(workflowUser)

	if result.Error != nil {
		return fmt.Errorf("failed to sync user to workflow user table: %w", result.Error)
	}

	log.Infof("Successfully synced user %d to workflow user table", userID)
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

