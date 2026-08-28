package config

import (
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/es"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/redis"
	"github.com/UnicomAI/wanwu/pkg/util"
)

var (
	_c *Config
)

type Config struct {
	Server       ServerConfig  `json:"server" mapstructure:"server"`
	Log          LogConfig     `json:"log" mapstructure:"log"`
	DB           db.Config     `json:"db" mapstructure:"db"`
	Redis        redis.Config  `json:"redis" mapstructure:"redis"`
	ES           es.Config     `json:"es" mapstructure:"es"`
	Assistant    Assistant     `json:"assistant" mapstructure:"assistant"`
	Minio        *MinioConfig  `mapstructure:"minio" json:"minio"`
	Knowledge    Knowledge     `mapstructure:"knowledge" json:"knowledge" yaml:"knowledge"`
	MCP          Mcp           `mapstructure:"mcp" json:"mcp"`
	Workflow     Workflow      `mapstructure:"workflow" json:"workflow"`
	Skill        Skill         `mapstructure:"skill" json:"skill"`
	BuiltinTools []BuiltinTool `mapstructure:"builtin_tools" json:"builtin_tools" yaml:"builtin_tools"`
}

// BuiltinTool 创建智能体时自动绑定的内置工具配置
type BuiltinTool struct {
	ToolId     string `mapstructure:"tool_id" json:"tool_id" yaml:"tool_id"`
	ToolType   string `mapstructure:"tool_type" json:"tool_type" yaml:"tool_type"`
	ActionName string `mapstructure:"action_name" json:"action_name" yaml:"action_name"`
}

type Knowledge struct {
	Host string `mapstructure:"host" json:"host" yaml:"host"`
}

type Mcp struct {
	Host string `mapstructure:"host" json:"host" yaml:"host"`
}

type Workflow struct {
	Endpoint      string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`
	ListSchemaUri string `mapstructure:"list_schema_uri" json:"list_schema_uri" yaml:"list_schema_uri"`
}

type Skill struct {
	Endpoint             string `mapstructure:"endpoint" json:"endpoint" yaml:"endpoint"`
	BuiltInSkillListUri  string `mapstructure:"builtin_skill_list_uri" json:"builtin_skill_list_uri" yaml:"builtin_skill_list_uri"`
	CustomSkillListUri   string `mapstructure:"custom_skill_list_uri" json:"custom_skill_list_uri" yaml:"custom_skill_list_uri"`
	AcquiredSkillListUri string `mapstructure:"acquired_skill_list_uri" json:"acquired_skill_list_uri" yaml:"acquired_skill_list_uri"`
}

type ServerConfig struct {
	GrpcEndpoint   string `json:"grpc_endpoint" mapstructure:"grpc_endpoint"`
	MaxRecvMsgSize int    `json:"max_recv_msg_size" mapstructure:"max_recv_msg_size"`
	CallbackUrl    string `json:"callback_url" mapstructure:"callback_url"`
}

type LogConfig struct {
	Std   bool         `json:"std" mapstructure:"std"`
	Level string       `json:"level" mapstructure:"level"`
	Logs  []log.Config `json:"logs" mapstructure:"logs"`
}

type DBConfig struct {
	Name string `json:"name" mapstructure:"name"`
}

type Assistant struct {
	SseUrl            string `mapstructure:"sse-url" json:"sse-url" yaml:"sse-url"`
	NewSseUrl         string `mapstructure:"new-sse-url" json:"new-sse-url" yaml:"new-sse-url"`
	MultiAgentChatUrl string `mapstructure:"multi-agent-chat-url" json:"multi-agent-chat-url" yaml:"multi-agent-chat-url"`
	MaxPicNum         int32  `mapstructure:"max_pic_num" json:"max_pic_num" yaml:"max_pic_num"`
}

type MinioConfig struct {
	EndPoint    string `json:"endpoint" mapstructure:"endpoint"`
	User        string `mapstructure:"user" json:"user"`
	Password    string `mapstructure:"password" json:"password"`
	DownloadURL string `json:"download_url" mapstructure:"download_url"`
	Bucket      string `mapstructure:"bucket" json:"bucket"`
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
