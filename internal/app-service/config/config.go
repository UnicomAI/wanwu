package config

import (
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/redis"
	"github.com/UnicomAI/wanwu/pkg/util"
)

var (
	_c *Config
)

type Config struct {
	Server    ServerConfig  `json:"server" mapstructure:"server"`
	Log       LogConfig     `json:"log" mapstructure:"log"`
	DB        db.Config     `json:"db" mapstructure:"db"`
	Redis     redis.Config  `json:"redis" mapstructure:"redis"`
	Minio     MinioConfig   `json:"minio" mapstructure:"minio"`         // 新增 MinIO 配置
	Assistant ServiceConfig `json:"assistant" mapstructure:"assistant"` // 数据清洗依赖 assistant（agent 老 id → uuid）
}

// ServiceConfig 微服务 grpc 连接配置。
type ServiceConfig struct {
	Host string `json:"host" mapstructure:"host"`
}

type ServerConfig struct {
	GrpcEndpoint   string `json:"grpc_endpoint" mapstructure:"grpc_endpoint"`
	MaxRecvMsgSize int    `json:"max_recv_msg_size" mapstructure:"max_recv_msg_size"`
}

type LogConfig struct {
	Std   bool         `json:"std" mapstructure:"std"`
	Level string       `json:"level" mapstructure:"level"`
	Logs  []log.Config `json:"logs" mapstructure:"logs"`
}

type DBConfig struct {
	Name string `json:"name" mapstructure:"name"`
}

type MinioConfig struct {
	Endpoint string `json:"endpoint" mapstructure:"endpoint"`
	User     string `json:"user" mapstructure:"user"`
	Password string `json:"password" mapstructure:"password"`
	Bucket   string `json:"bucket" mapstructure:"bucket"` // 安全模块的 bucket
}

func LoadConfig(in string) error {
	_c = &Config{}
	return util.LoadConfig(in, _c)
}

func Cfg() *Config {
	if _c == nil {
		log.Panicf("cfg nil")
	}
	return _c
}
