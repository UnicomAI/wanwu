package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const (
	EXEC_MODE_SET_PASSWORD = "set_password"
	EXEC_MODE_INIT_DATA    = "init_data"
)

func main() {
	// 1. 设置密码
	if err := setupTidb(EXEC_MODE_SET_PASSWORD); err != nil {
		log.Fatalf("setup tidb set password err: %v", err)
	}

	// 2. 初始化数据
	if err := setupTidb(EXEC_MODE_INIT_DATA); err != nil {
		log.Fatalf("setup tidb init data err: %v", err)
	}
}

func setupTidb(execMode string) error {
	// 0. 解析环境变量
	dbHost := getEnvWithDefault("WANWU_TIDB_HOST", "127.0.0.1")
	dbPort := getEnvWithDefault("WANWU_TIDB_PORT", "4000")
	dbUser := getEnvWithDefault("WANWU_TIDB_USER", "root")
	dbPassword := os.Getenv("WANWU_TIDB_PASSWORD")
	sqlFile := os.Getenv("WANWU_TIDB_SQL_FILE")

	if dbPassword != "" {
		if strings.ContainsAny(dbPassword, "'\";`") {
			return fmt.Errorf("invalid password: contains forbidden characters")
		}
		if len(dbPassword) > 128 {
			return fmt.Errorf("invalid password: too long")
		}
	}

	if sqlFile != "" {
		cleanPath := filepath.Clean(sqlFile)
		if strings.Contains(cleanPath, "..") {
			return fmt.Errorf("invalid sql file path: path traversal not allowed")
		}
		absPath, err := filepath.Abs(sqlFile)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for sql file: %w", err)
		}
		if !strings.HasPrefix(absPath, "/opt/") && !strings.HasPrefix(absPath, "/home/") {
			log.Printf("warning: sql file path may be unsafe: %s", sqlFile)
		}
	}

	// 1. 连接数据库
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&tls=false",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
	))
	if err != nil {
		return fmt.Errorf("init db exec mode %v err: %v", execMode, err)
	}
	defer func() { _ = db.Close() }()
	switch execMode {
	case EXEC_MODE_SET_PASSWORD:
		if err = db.Ping(); err != nil {
			// 尝试无密码连接
			db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&tls=false",
				dbUser,
				"",
				dbHost,
				dbPort,
			))
			if err != nil {
				return fmt.Errorf("init db exec mode %v without password err: %v", execMode, err)
			}
			defer func() { _ = db.Close() }()
		} else {
			log.Printf("already setup db, exec mode: %v", execMode)
			return nil
		}
	case EXEC_MODE_INIT_DATA:
		// do nothing
	default:
		return fmt.Errorf("invalid exec mode: %v", execMode)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("ping db exec mode %v err: %v", execMode, err)
	}

	// 2. 开启会话级多语句模式
	if _, err = db.Exec("SET tidb_multi_statement_mode = 'ON'"); err != nil {
		return fmt.Errorf("set tidb_multi_statement_mode err: %v", err)
	}

	switch execMode {
	case EXEC_MODE_SET_PASSWORD:
		// 使用参数化查询替代字符串拼接，防止 SQL 注入
		if _, err = db.Exec("USE mysql;ALTER USER 'root'@'%%' IDENTIFIED BY ?;FLUSH PRIVILEGES;SET GLOBAL tidb_skip_isolation_level_check = 1;", dbPassword); err != nil {
			return fmt.Errorf("exec sql err: %v", err)
		}
	case EXEC_MODE_INIT_DATA:
		sqlBytes, err := os.ReadFile(sqlFile)
		if err != nil {
			return fmt.Errorf("read sql file %v err: %v", sqlFile, err)
		}
		if _, err = db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("exec sql err: %v", err)
		}
	default:
		return fmt.Errorf("invalid exec mode: %v", execMode)
	}

	log.Printf("setup db success, exec mode: %v", execMode)
	return nil
}

func getEnvWithDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}
