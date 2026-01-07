package model

import "github.com/UnicomAI/wanwu/pkg/db"

type ModelImported struct {
	ID             uint32      `gorm:"primary_key;auto_increment;not null;"`
	UUID           string      `gorm:"column:uuid;type:varchar(255);uniqueIndex:idx_unique_uuid;comment:模型uuid"`
	Provider       string      `gorm:"column:provider;index:idx_model_imported_provider_type_model,priority:1;type:varchar(100);comment:模型供应商"`
	ModelType      string      `gorm:"column:model_type;index:idx_model_imported_provider_type_model,priority:2;type:varchar(100);comment:模型类型"`
	Model          string      `gorm:"column:model;index:idx_model_imported_provider_type_model,priority:3;type:varchar(100);comment:模型名称"`
	DisplayName    string      `gorm:"column:display_name;idx:idx_model_imported_model_display_name;type:varchar(100);comment:模型显示名称"`
	ModelIconPath  string      `gorm:"column:model_icon_path;type:varchar(512);comment:模型图标路径"`
	IsActive       bool        `gorm:"column:is_active;default:true;comment:模型是否启用"`
	ProviderConfig db.LongText `gorm:"column:provider_config;comment:某供应商下的模型配置"`
	ModelDesc      db.LongText `gorm:"column:model_desc;comment:模型描述"`
	PublishDate    string      `gorm:"column:publish_date;type:varchar(100);comment:模型发布时间"`
	PublicModel
}
