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
	Server ServerConfig `json:"server" mapstructure:"server"`
	Log    LogConfig    `json:"log" mapstructure:"log"`
	DB     db.Config    `json:"db" mapstructure:"db"`
	Redis  redis.Config `json:"redis" mapstructure:"redis"`

	// --- microservice ---
	Iam       ServiceConfig `json:"iam" mapstructure:"iam"`
	Assistant ServiceConfig `json:"assistant" mapstructure:"assistant"`

	// --- 消息中心 ---
	Notice NoticeConfig `json:"notice" mapstructure:"notice"`
}

type ServerConfig struct {
	GrpcEndpoint   string `json:"grpc_endpoint" mapstructure:"grpc_endpoint"`
	MaxRecvMsgSize int    `json:"max_recv_msg_size" mapstructure:"max_recv_msg_size"`
}

type ServiceConfig struct {
	Host string `json:"host" mapstructure:"host"`
}

// NoticeConfig 消息中心配置
type NoticeConfig struct {
	// MaxAudienceSize 特定用户名单 S 的规模上限；超限拒绝生成该事件消息（打点告警），不阻塞业务主流程
	MaxAudienceSize int `json:"max_audience_size" mapstructure:"max_audience_size"`
}

// 消息中心配置默认值（配置缺省时生效）
const (
	DefaultNoticeMaxAudienceSize = 500
)

// Fill 用默认值补齐未配置项
func (c NoticeConfig) Fill() NoticeConfig {
	if c.MaxAudienceSize <= 0 {
		c.MaxAudienceSize = DefaultNoticeMaxAudienceSize
	}
	return c
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
