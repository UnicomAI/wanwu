package db

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	DBName     string     `json:"name" mapstructure:"name"`
	MySQL      ConnConfig `json:"mysql" mapstructure:"mysql"`
	PostgreSQL ConnConfig `json:"postgres" mapstructure:"postgres"`
	TiDB       ConnConfig `json:"tidb" mapstructure:"tidb"`
	OceanBase  ConnConfig `json:"oceanbase" mapstructure:"oceanbase"`
}

type ConnConfig struct {
	Address      string `json:"address" mapstructure:"address"`
	User         string `json:"user" mapstructure:"user"`
	Password     string `json:"password" mapstructure:"password"`
	Database     string `json:"database" mapstructure:"database"`
	MaxOpenConns int    `json:"max_open_conns" mapstructure:"max_open_conns"`
	MaxIdleConns int    `json:"max_idle_conns" mapstructure:"max_idle_conns"`
	LogMode      bool   `json:"log_mode" mapstructure:"log_mode"`
}

func New(cfg Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch cfg.DBName {
	case "mysql", "tidb", "oceanbase":
		var connCfg ConnConfig
		switch cfg.DBName {
		case "mysql":
			connCfg = cfg.MySQL
		case "tidb":
			connCfg = cfg.TiDB
		case "oceanbase":
			connCfg = cfg.OceanBase
		}
		db, err = gorm.Open(mysql.Open(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=%t&loc=%s",
			connCfg.User,
			connCfg.Password,
			connCfg.Address,
			connCfg.Database,
			true,
			// "Asia/Shanghai",
			"Local")), buildGormConfig(connCfg))
		if err != nil {
			break
		}
		err = setPoolParam(db, connCfg.MaxOpenConns, connCfg.MaxIdleConns)
	case "postgres":
		db, err = gorm.Open(postgres.Open(fmt.Sprintf("postgres://%s:%s@%s/%s",
			cfg.PostgreSQL.User,
			cfg.PostgreSQL.Password,
			cfg.PostgreSQL.Address,
			cfg.PostgreSQL.Database,
		)), &gorm.Config{})
		if err != nil {
			break
		}
		err = setPoolParam(db, cfg.PostgreSQL.MaxOpenConns, cfg.PostgreSQL.MaxIdleConns)
	default:
		err = fmt.Errorf("invalid db %v", cfg.DBName)
	}
	if err != nil {
		return nil, err
	}
	return db, err
}

// Build gorm log configuration
func buildGormConfig(cfg ConnConfig) *gorm.Config {
	var gormConfig *gorm.Config
	if cfg.LogMode { //Determine whether to enable logs based on configuration
		gormConfig = &gorm.Config{
			Logger:                                   logger.Default.LogMode(logger.Info),
			DisableForeignKeyConstraintWhenMigrating: true,
		}
	} else {
		gormConfig = &gorm.Config{
			Logger:                                   logger.Default.LogMode(logger.Silent),
			DisableForeignKeyConstraintWhenMigrating: true,
		}
	}
	return gormConfig
}

func setPoolParam(db *gorm.DB, maxOpenConn, maxIdleConn int) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(maxOpenConn)
	sqlDB.SetMaxIdleConns(maxIdleConn)
	return nil
}
