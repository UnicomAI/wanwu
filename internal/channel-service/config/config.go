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
	Server   ServerConfig   `json:"server" mapstructure:"server"`
	Callback CallbackConfig `json:"callback" mapstructure:"callback"`
	Log      LogConfig      `json:"log" mapstructure:"log"`
	DB       db.Config      `json:"db" mapstructure:"db"`
	Redis    redis.Config   `json:"redis" mapstructure:"redis"`
	BFF      BFFConfig      `json:"bff" mapstructure:"bff"`
	DingTalk DingTalkConfig `json:"dingtalk" mapstructure:"dingtalk"`
	WeChat   WeChatConfig   `json:"wechat" mapstructure:"wechat"`
}

type ServerConfig struct {
	GrpcEndpoint   string `json:"grpc_endpoint" mapstructure:"grpc_endpoint"`
	MaxRecvMsgSize int    `json:"max_recv_msg_size" mapstructure:"max_recv_msg_size"`
}

// CallbackConfig HTTP 回调服务器配置（用于接收平台 Webhook 推送）
type CallbackConfig struct {
	Endpoint string `json:"endpoint" mapstructure:"endpoint"`
}

type LogConfig struct {
	Std   bool         `json:"std" mapstructure:"std"`
	Level string       `json:"level" mapstructure:"level"`
	Logs  []log.Config `json:"logs" mapstructure:"logs"`
}

// BFFConfig 万悟 BFF 服务地址，用于代理调用 OpenAPI
type BFFConfig struct {
	ApiBaseUrl string `json:"api_base_url" mapstructure:"api_base_url"`
}

// DingTalkConfig 钉钉通道配置
type DingTalkConfig struct {
	// StreamMode 默认使用 Stream 模式
	StreamMode bool `json:"stream_mode" mapstructure:"stream_mode"`
	// OAuthCallbackURL 扫码登录回调地址
	OAuthCallbackURL string `json:"oauth_callback_url" mapstructure:"oauth_callback_url"`
}

// WeChatConfig 微信通道配置
type WeChatConfig struct {
	BaseURL string `json:"base_url" mapstructure:"base_url"`
	// CallbackURL 消息回调地址
	CallbackURL string `json:"callback_url" mapstructure:"callback_url"`
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
