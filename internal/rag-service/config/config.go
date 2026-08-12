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
	Server    ServerConfig    `json:"server" mapstructure:"server"`
	Log       LogConfig       `json:"log" mapstructure:"log"`
	DB        db.Config       `json:"db" mapstructure:"db"`
	Redis     redis.Config    `json:"redis" mapstructure:"redis"`
	ES        es.Config       `json:"es" mapstructure:"es"`
	Knowledge Knowledge       `mapstructure:"knowledge" json:"knowledge" yaml:"knowledge"`
	RagServer RagServerConfig `mapstructure:"ragServer" json:"ragServer"`
	Minio     *MinioConfig    `mapstructure:"minio" json:"minio"`
}

type MinioConfig struct {
	EndPoint    string `json:"endpoint" mapstructure:"endpoint"`
	User        string `mapstructure:"user" json:"user"`
	Password    string `mapstructure:"password" json:"password"`
	DownloadURL string `json:"download_url" mapstructure:"download_url"`
	Bucket      string `mapstructure:"bucket" json:"bucket"`
}

type RagServerConfig struct {
	ChatEndpoint string `mapstructure:"chatEndpoint" json:"chatEndpoint"`
	ChatUrl      string `mapstructure:"chatUrl" json:"chatUrl"`
	QASearchUrl  string `mapstructure:"qaSearchUrl" json:"qaSearchUrl"`
}

type Knowledge struct {
	Host string `mapstructure:"host" json:"host" yaml:"host"`
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
