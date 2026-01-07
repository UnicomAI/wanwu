package config

import (
	"strings"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg"
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/spf13/viper"
)

var config = Config{}

func init() {
	pkg.AddContainer(config)
}

func GetConfig() *Config {
	return &config
}

func (c Config) LoadType() string {
	return "configs"
}

func (c Config) Load() error {
	viper.SetConfigFile("configs/microservice/knowledge-service/configs/config.yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()
	viper.AllowEmptyEnv(true)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}
	config = cfg
	return nil
}

func (c Config) StopPriority() int {
	return pkg.DefaultPriority
}

func (c Config) Stop() error {
	return nil
}

// 系统配置，对应yml
// viper内置了mapstructure, yml文件用"-"区分单词, 转为驼峰方便

// 需要详细自定义配置项目
// 主要集中在系统配置、log配置、mysql、请求权限控制、请求速率限制
// 全局配置变量

type Config struct {
	Server             *Server             `mapstructure:"server" json:"server"`
	Log                LogConfig           `json:"log" mapstructure:"log"`
	AccessLog          LogConfig           `mapstructure:"access-log" json:"access-log" yaml:"access-log"`
	RpcLog             LogConfig           `mapstructure:"rpc-log" json:"rpc-log" yaml:"rpc-log"`
	DB                 db.Config           `json:"db" mapstructure:"db"`
	Minio              *MinioConfig        `mapstructure:"minio" json:"minio"`
	Topic              *TopicConfig        `mapstructure:"topic" json:"topic"`
	Kafka              *KafkaConfig        `mapstructure:"kafka" json:"kafka"`
	Redis              *RedisConfig        `mapstructure:"redis" json:"redis"` // 新增Redis配置
	UsageLimit         *UsageLimitConfig   `mapstructure:"usage-limit" json:"usageLimit"`
	RagServer          *RagServerConfig    `mapstructure:"rag-server" json:"ragServer"`
	KnowledgeDocConfig *KnowledgeDocConfig `json:"knowledge-doc-config" mapstructure:"knowledge-doc-config"`
	SplitterList       []*Splitter         `mapstructure:"splitters" json:"splitters" yaml:"splitters"`
}

type Splitter struct {
	Name  string `mapstructure:"name" json:"name" yaml:"name"`
	Value string `mapstructure:"value" json:"value" yaml:"value"`
}

type Server struct {
	GrpcEndpoint   string `mapstructure:"grpc-endpoint" json:"grpcEndpoint"`
	MaxRecvMsgSize int    `mapstructure:"max-recv-msg-size" json:"maxRecvMsgSize"`
}

type LogConfig struct {
	Std   bool         `json:"std" mapstructure:"std"`
	Level string       `json:"level" mapstructure:"level"`
	Logs  []log.Config `json:"logs" mapstructure:"logs"`
}

type MinioConfig struct {
	EndPoint           string `json:"endpoint" mapstructure:"endpoint"`
	KnowledgeDir       string `mapstructure:"knowledge-dir" json:"knowledge-dir"`
	KnowledgeExportDir string `mapstructure:"knowledge-export-dir" json:"knowledge-export-dir"`
	QAExportDir        string `mapstructure:"qa-export-dir" json:"qa-export-dir"`
	User               string `mapstructure:"user" json:"user"`
	Password           string `mapstructure:"password" json:"password"`
	Bucket             string `mapstructure:"bucket" json:"bucket"`
	PublicRagBucket    string `mapstructure:"public-rag-bucket" json:"public-rag-bucket"`
	PublicExportBucket string `mapstructure:"public-export-bucket" json:"public-export-bucket"`
}

type TopicConfig struct {
	UrlAnalysisTopic    string `mapstructure:"url-analysis-topic" json:"url-analysis-topic"`
	UrlImportTopic      string `mapstructure:"url-import-topic" json:"url-import-topic"`
	Topic               string `mapstructure:"topic" json:"topic"`
	KnowledgeGraphTopic string `mapstructure:"knowledge-graph-topic" json:"knowledge-graph-topic"`
}

type KafkaConfig struct {
	Enabled             bool   `mapstructure:"enabled" json:"enabled"`
	Addr                string `mapstructure:"addr" json:"addr"`
	User                string `mapstructure:"user" json:"user"`
	Password            string `mapstructure:"password" json:"password"`
	DefaultPartitionNum int32  `mapstructure:"default-partition-num" json:"defaultPartitionNum"`
}

type RedisConfig struct {
	Enabled      bool     `mapstructure:"enabled" json:"enabled"`
	Mode         string   `mapstructure:"mode" json:"mode"`              // standalone, sentinel, cluster
	Addr         []string `mapstructure:"addr" json:"addr"`              // 单节点: ["host:port"]; 哨兵/集群: ["host1:port1", "host2:port2", ...]
	MasterName   string   `mapstructure:"master-name" json:"masterName"` // 哨兵模式专用
	Password     string   `mapstructure:"password" json:"password"`
	DB           int      `mapstructure:"db" json:"db"` // 仅standalone模式有效
	PoolSize     int      `mapstructure:"pool-size" json:"poolSize"`
	MinIdleConns int      `mapstructure:"min-idle-conns" json:"minIdleConns"`
	MaxRetries   int      `mapstructure:"max-retries" json:"maxRetries"`
	DialTimeout  int      `mapstructure:"dial-timeout" json:"dialTimeout"`   // 秒
	ReadTimeout  int      `mapstructure:"read-timeout" json:"readTimeout"`   // 秒
	WriteTimeout int      `mapstructure:"write-timeout" json:"writeTimeout"` // 秒
	IdleTimeout  int      `mapstructure:"idle-timeout" json:"idleTimeout"`   // 秒
	// Stream 配置
	StreamMaxLen       int64 `mapstructure:"stream-max-len" json:"streamMaxLen"`              // Stream 最大长度
	StreamApproxMaxLen bool  `mapstructure:"stream-approx-max-len" json:"streamApproxMaxLen"` // 是否使用近似最大长度
}

type UsageLimitConfig struct {
	DocTotal                     int64  `mapstructure:"doc-total" json:"docTotal"`
	FileTypes                    string `mapstructure:"file-types" json:"fileTypes"`
	MaxFileSize                  int64  `mapstructure:"max-file-size" json:"maxFileSize"` //单位：字节
	CompressedFileType           string `mapstructure:"compressed-file-type" json:"compressedFileType"`
	MaxNumberOfFilesInCompressed int64  `mapstructure:"max-number-of-files-in-compressed" json:"maxNumberOfFilesInCompressed"`
	FileSizeLimit                int64  `mapstructure:"file-size-limit" json:"fileSizeLimit"`
	TxtSizeLimit                 int64  `mapstructure:"txt-size-limit" json:"txtSizeLimit"`
	DocxSizeLimit                int64  `mapstructure:"docx-size-limit" json:"docxSizeLimit"`
	PdfSizeLimit                 int64  `mapstructure:"pdf-size-limit" json:"pdfSizeLimit"`
	ExcelSizeLimit               int64  `mapstructure:"excel-size-limit" json:"excelSizeLimit"`
	CsvSizeLimit                 int64  `mapstructure:"csv-size-limit" json:"csvSizeLimit"`
	PptxSizeLimit                int64  `mapstructure:"pptx-size-limit" json:"pptxSizeLimit"`
	HtmlSizeLimit                int64  `mapstructure:"html-size-limit" json:"htmlSizeLimit"`
	CompressedSizeLimit          int64  `mapstructure:"compressed-size-limit" json:"compressedSizeLimit"`
	UploadConcurrentLimit        int64  `mapstructure:"upload-concurrent-limit" json:"uploadConcurrentLimit"`
	UrlCountLimit                int    `mapstructure:"url-count-limit" json:"url-count-limit"`
}

type KnowledgeDocConfig struct {
	DocLocalFilePath string `mapstructure:"doc-local-file-path" json:"doc-local-file-path"`
}

type RagServerConfig struct {
	Endpoint                  string `mapstructure:"endpoint" json:"endpoint"`
	ProxyPoint                string `mapstructure:"proxy-point" json:"proxy-point"`
	UrlImportEndpoint         string `mapstructure:"url-import-endpoint" json:"url-import-endpoint"`
	UrlAnalysisEndpoint       string `mapstructure:"url-analysis-endpoint" json:"url-analysis-endpoint"`
	InitKnowledgeUri          string `mapstructure:"init-knowledge-uri" json:"init-knowledge-uri"`
	UpdateKnowledgeUri        string `mapstructure:"update-knowledge-uri" json:"update-knowledge-uri"`
	DeleteKnowledgeUri        string `mapstructure:"delete-knowledge-uri" json:"delete-knowledge-uri"`
	KnowledgeHitUri           string `mapstructure:"knowledge-hit-uri" json:"knowledge-hit-uri"`
	GetDocSegmentUri          string `mapstructure:"get-doc-segment-uri" json:"get-doc-segment-uri"`
	GetDocChildSegmentUri     string `mapstructure:"get-child-content-uri" json:"get-child-content-uri"`
	DocSegmentUpdateStatusUri string `mapstructure:"doc-segment-update-status-uri" json:"doc-segment-update-status-uri"`
	DocDeleteUri              string `mapstructure:"doc-delete-uri" json:"doc-delete-uri"`
	UpdateFileMetasUri        string `mapstructure:"update-file-metas-uri" json:"update-file-metas-uri"`
	DocUrlImportUri           string `mapstructure:"doc-url-import-uri" json:"doc-url-import-uri"`
	DocUrlAnalysisUri         string `mapstructure:"doc-url-analysis-uri" json:"doc-url-analysis-uri"`
	KeywordsUri               string `mapstructure:"keywords-uri" json:"keywords-uri"`
	DocSegmentUpdateLabelsUri string `mapstructure:"doc-segment-update-labels-uri" json:"doc-segment-update-labels-uri"`
	DocSegmentCreateUri       string `mapstructure:"doc-segment-create-uri" json:"doc-segment-create-uri"`
	DocSegmentUpdateUri       string `mapstructure:"doc-segment-update-uri" json:"doc-segment-update-uri"`
	DocSegmentDeleteUri       string `mapstructure:"doc-segment-delete-uri" json:"doc-segment-delete-uri"`
	DocChildSegmentCreateUri  string `mapstructure:"doc-child-segment-create-uri" json:"doc-child-segment-create-uri"`
	DocChildSegmentDeleteUri  string `mapstructure:"doc-child-segment-delete-uri" json:"doc-child-segment-delete-uri"`
	DocChildSegmentUpdateUri  string `mapstructure:"doc-child-segment-update-uri" json:"doc-child-segment-update-uri"`
	BatchDeleteMetaKeyUri     string `mapstructure:"batch-delete-meta-key-uri" json:"batch-delete-meta-key-uri"`
	BatchRenameMetakeyUri     string `mapstructure:"batch-rename-meta-key-uri" json:"batch-rename-meta-key-uri"`
	GetCommunityReportListUri string `mapstructure:"get-community-report-list-uri" json:"get-community-report-list-uri"`
	BatchAddReportsUri        string `mapstructure:"batch-add-reports-uri" json:"batch-add-reports-uri"`
	UpdateReportUri           string `mapstructure:"update-report-uri" json:"update-report-uri"`
	BatchDeleteReportsUri     string `mapstructure:"batch-delete-reports-uri" json:"batch-delete-reports-uri"`
	KnowledgeGraphUri         string `mapstructure:"knowledge-graph-uri" json:"knowledge-graph-uri"`
	InitQABaseUri             string `mapstructure:"init-QA-base-uri" json:"init-QA-base-uri"`
	DeleteQABaseUri           string `mapstructure:"delete-QA-base-uri" json:"delete-QA-base-uri"`
	UpdateQAMetasUri          string `mapstructure:"update-QA-metas-uri" json:"update-QA-metas-uri"`
	DeleteQAMetaKeyUri        string `mapstructure:"delete-QA-meta-key-uri" json:"delete-QA-meta-key-uri"`
	RenameQAMetakeyUri        string `mapstructure:"rename-QA-meta-key-uri" json:"rename-QA-meta-key-uri"`
	BatchAddQAPairsUri        string `mapstructure:"batch-add-qa-pairs-uri" json:"batch-add-qa-pairs-uri"`
	UpdateQAPairUri           string `mapstructure:"update-qa-pair-uri" json:"update-qa-pair-uri"`
	UpdateQAPairStatusUri     string `mapstructure:"update-qa-pair-status-uri" json:"update-qa-pair-status-uri"`
	BatchDeleteQAPairsUri     string `mapstructure:"batch-delete-qa-pairs-uri" json:"batch-delete-qa-pairs-uri"`
	KnowledgeQAHitUri         string `mapstructure:"knowledge-QA-hit-uri" json:"knowledge-QA-hit-uri"`
	KnowledgeDocSearchUri     string `mapstructure:"knowledge-doc-search-uri" json:"knowledge-doc-search-uri"`
	Timeout                   int64  `mapstructure:"timeout" json:"timeout"`
}
