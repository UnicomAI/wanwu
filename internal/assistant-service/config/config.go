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
	Server    ServerConfig `json:"server" mapstructure:"server"`
	Log       LogConfig    `json:"log" mapstructure:"log"`
	DB        db.Config    `json:"db" mapstructure:"db"`
	Redis     redis.Config `json:"redis" mapstructure:"redis"`
	Assistant Assistant    `json:"assistant" mapstructure:"assistant"`
	Minio     *MinioConfig `mapstructure:"minio" json:"minio"`
	Knowledge Knowledge    `mapstructure:"knowledge" json:"knowledge" yaml:"knowledge"`
	MCP       Mcp          `mapstructure:"mcp" json:"mcp"`
	Workflow  Workflow     `mapstructure:"workflow" json:"workflow"`
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
	SseUrl    string `mapstructure:"sse-url" json:"sse-url" yaml:"sse-url"`
	NewSseUrl string `mapstructure:"new-sse-url" json:"new-sse-url" yaml:"new-sse-url"`
	MaxPicNum int32  `mapstructure:"max_pic_num" json:"max_pic_num" yaml:"max_pic_num"`
}

type MinioConfig struct {
	EndPoint string `json:"endpoint" mapstructure:"endpoint"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
	Bucket   string `mapstructure:"bucket" json:"bucket"`
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
