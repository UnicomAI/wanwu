package config

import (
	"fmt"
	"net/url"
	"os"

	oauth2_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/oauth2-util"
	"github.com/UnicomAI/wanwu/pkg/i18n"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/UnicomAI/wanwu/pkg/redis"
	rsa_util "github.com/UnicomAI/wanwu/pkg/rsa-util"
	"github.com/UnicomAI/wanwu/pkg/util"
)

var (
	_c *Config
)

type Config struct {
	Server            ServerConfig               `json:"server" mapstructure:"server"`
	Log               LogConfig                  `json:"log" mapstructure:"log"`
	JWT               JWTConfig                  `json:"jwt" mapstructure:"jwt"`
	OAuth             OAuthConfig                `json:"oauth" mapstructure:"oauth"`
	Decrypt           DecryptPasswd              `json:"decrypt-passwd" mapstructure:"decrypt-passwd"`
	RSAKey            RSAKeyConfig               `json:"rsa-key" mapstructure:"rsa-key"`
	I18n              i18n.Config                `json:"i18n" mapstructure:"i18n"`
	CustomInfo        CustomInfoConfig           `json:"custom-info" mapstructure:"custom-info"`
	DocCenter         DocCenterConfig            `json:"doc-center" mapstructure:"doc-center"`
	DefaultIcon       DefaultIconConfig          `json:"default-icon" mapstructure:"default-icon"`
	AssistantTemplate AssistantTemplateConfig    `json:"assistant-template" mapstructure:"assistant-template"`
	WorkflowTemplate  WorkflowTemplatePathConfig `json:"workflow-template" mapstructure:"workflow-template"`
	PromptTemplate    PromptTemplatePathConfig   `json:"prompt-template" mapstructure:"prompt-template"`
	SkillsTemplate    SkillsTemplatePathConfig   `json:"skills-template" mapstructure:"skills-template"`
	WorkflowTemplates []*WorkflowTemplateConfig  `json:"workflows" mapstructure:"workflows"`
	PromptTemplates   []*PromptTempConfig        `json:"prompts" mapstructure:"prompts"`
	AgentSkills       []*SkillsConfig            `json:"skills" mapstructure:"skills"`
	PromptEngineering PromptEngineeringConfig    `json:"prompt-engineering" mapstructure:"prompt-engineering"`
	SkillCreator      SkillCreatorConfig         `json:"skill-creator" mapstructure:"skill-creator"`
	SkillCreatorPath  SkillCreatorPathConfig     `json:"skill-creator-path" mapstructure:"skill-creator-path"`
	RecommendModels   RecommendModelConfig       `json:"recommend_models" mapstructure:"recommend_models"`
	// middleware
	Minio minio.Config `json:"minio" mapstructure:"minio"`
	Redis redis.Config `json:"redis" mapstructure:"redis"`
	// microservice
	Iam                 ServiceConfig         `json:"iam" mapstructure:"iam"`
	Model               ServiceModelConfig    `json:"model" mapstructure:"model"`
	MCP                 ServiceConfig         `json:"mcp" mapstructure:"mcp"`
	App                 ServiceConfig         `json:"app" mapstructure:"app"`
	Knowledge           ServiceConfig         `json:"knowledge" mapstructure:"knowledge"`
	Rag                 ServiceConfig         `json:"rag" mapstructure:"rag"`
	Assistant           ServiceConfig         `json:"assistant" mapstructure:"assistant"`
	Operate             ServiceConfig         `json:"operate" mapstructure:"operate"`
	AgentService        ServiceConfig         `json:"agent-service" mapstructure:"agent-service"`
	Channel             ServiceConfig         `json:"channel" mapstructure:"channel"`
	RagKnowledgeConfig  RagKnowledgeConfig    `json:"rag-knowledge" mapstructure:"rag-knowledge"`
	DifyKnowledgeConfig DifyKnowledgeConfig   `json:"dify-knowledge" mapstructure:"dify-knowledge"`
	Workflow            WorkflowServiceConfig `json:"workflow" mapstructure:"workflow"`
	WgaSandbox          WgaSandboxConfig      `json:"wga-sandbox" mapstructure:"wga-sandbox"`
	Ontology            OntologyServiceConfig `json:"ontology" mapstructure:"ontology"`
}

type ServerConfig struct {
	Host         string `json:"host" mapstructure:"host"`
	Port         int    `json:"port" mapstructure:"port"`
	CallbackPort int    `json:"callback_port" mapstructure:"callback_port"`
	WebBaseUrl   string `json:"web_base_url" mapstructure:"web_base_url"`
	ApiBaseUrl   string `json:"api_base_url" mapstructure:"api_base_url"`
	AppOpenUrl   string `json:"app_open_base_url" mapstructure:"app_open_base_url"`
	CallbackUrl  string `json:"callback_url" mapstructure:"callback_url"`
}

type ServiceModelConfig struct {
	Host                     string `json:"host" mapstructure:"host"`
	PngTestFilePath          string `json:"png_test_file_path" mapstructure:"png_test_file_path"`
	PdfTestFilePath          string `json:"pdf_test_file_path" mapstructure:"pdf_test_file_path"`
	AsrTestFilePath          string `json:"asr_test_file_path" mapstructure:"asr_test_file_path"`
	RecommendModelConfigPath string `json:"recommend_model_config_path" mapstructure:"recommend_model_config_path"`
}

type RecommendModelItemLLM struct {
	Model       string   `mapstructure:"model" json:"model"`
	DisplayName string   `mapstructure:"display_name" json:"displayName"`
	Tags        []string `mapstructure:"tags" json:"tags"`
}

type RecommendModelItemOCR struct {
	Model            string   `mapstructure:"model" json:"model"`
	DisplayName      string   `mapstructure:"display_name" json:"displayName"`
	SupportFileTypes []string `mapstructure:"support_filetypes" json:"supportFileTypes"`
}

type RecommendModelItem struct {
	Model       string `mapstructure:"model" json:"model"`
	DisplayName string `mapstructure:"display_name" json:"displayName"`
}

type LogConfig struct {
	Std   bool         `json:"std" mapstructure:"std"`
	Level string       `json:"level" mapstructure:"level"`
	Logs  []log.Config `json:"logs" mapstructure:"logs"`
}

type JWTConfig struct {
	SigningKey string `json:"signing-key" mapstructure:"signing-key"`
}

type OAuthSwitch struct {
	OAuthSwitch int `json:"switch" mapstructure:"switch"`
}

type OAuthJWTConfig struct {
	RSAPrivateKeyPath string `json:"private_key_path" mapstructure:"private_key_path"`
	RSAPublicKeyPath  string `json:"public_key_path" mapstructure:"public_key_path"`
}

type OAuthConfig struct {
	Switch int                   `json:"switch" mapstructure:"switch"`
	RSA    oauth2_util.RSAConfig `json:"rsa" mapstructure:"rsa"`
}

type DecryptPasswd struct {
	IV  string `json:"iv" mapstructure:"iv"`
	Key string `json:"key" mapstructure:"key"`
}

// RSAKeyConfig RSA密钥配置
type RSAKeyConfig struct {
	// 私钥文件路径（权限应为 600）
	PrivateKeyPath string `json:"private_key_path" mapstructure:"private_key_path"`
	// 公钥文件路径（权限可为 644）
	PublicKeyPath string `json:"public_key_path" mapstructure:"public_key_path"`
	// 密钥轮换周期（天），默认 90 天
	KeyRotationDays int `json:"key_rotation_days" mapstructure:"key_rotation_days"`
}

type ServiceConfig struct {
	Host string `json:"host" mapstructure:"host"`
}

type RagKnowledgeConfig struct {
	Endpoint               string `json:"endpoint" mapstructure:"endpoint"`
	ChatEndpoint           string `json:"chat-endpoint" mapstructure:"chat-endpoint"`
	UploadEndpoint         string `json:"upload-endpoint" mapstructure:"upload-endpoint"`
	SearchKnowledgeBaseUri string `json:"search-knowledge-base-uri" mapstructure:"search-knowledge-base-uri"`
	SearchKnowTimeout      int    `json:"search-know-timeout" mapstructure:"search-know-timeout"`
	SearchQABaseUri        string `json:"search-qa-base-uri" mapstructure:"search-qa-base-uri"`
	KnowledgeChatUri       string `json:"knowledge-chat-uri" mapstructure:"knowledge-chat-uri"`
	UploadUri              string `json:"upload-uri" mapstructure:"upload-uri"`
	UploadBucket           string `json:"upload-bucket" mapstructure:"upload-bucket"`
	UploadTime             int64  `json:"upload-timeout" mapstructure:"upload-timeout"` //单位s
}

type DifyKnowledgeConfig struct {
	SearchKnowledgeBaseUri string `json:"search-knowledge-base-uri" mapstructure:"search-knowledge-base-uri"`
	SearchKnowTimeout      int    `json:"search-know-timeout" mapstructure:"search-know-timeout"`
}

type WgaSandboxConfig struct {
	Sandbox         WgaSandboxSandboxConfig `json:"sandbox" mapstructure:"sandbox"`
	SandboxOntology WgaSandboxSandboxConfig `json:"sandbox-ontology" mapstructure:"sandbox-ontology"`
}

type WgaSandboxSandboxConfig struct {
	Type      string `json:"type" mapstructure:"type"`
	Host      string `json:"host" mapstructure:"host"`
	ImageName string `json:"image-name" mapstructure:"image-name"`
}

type OntologyServiceConfig struct {
	Enable                        int                            `json:"enable" mapstructure:"enable"`
	Endpoint                      string                         `json:"endpoint" mapstructure:"endpoint"`
	KnowledgeNetworkListUri       string                         `json:"knowledge_network_list_uri" mapstructure:"knowledge_network_list_uri"`
	DigitalEmployeeListUri        string                         `json:"digital_employee_list_uri" mapstructure:"digital_employee_list_uri"`
	DigitalEmployeeInfoUri        string                         `json:"digital_employee_info_uri" mapstructure:"digital_employee_info_uri"`
	DigitalEmployeePromptTemplate string                         `json:"digital_employee_prompt_template" mapstructure:"digital_employee_prompt_template"`
	BuiltinSkills                 []string                       `yaml:"builtin_skills" json:"builtin_skills" mapstructure:"builtin_skills"`
	SmartDataSkills               []OntologySmartDataSkillConfig `json:"smart_data_skills" mapstructure:"smart_data_skills"`
}

type OntologySmartDataSkillConfig struct {
	Name      string `json:"name" mapstructure:"name"`
	Desc      string `json:"desc" mapstructure:"desc"`
	SkillPath string `json:"skillPath" mapstructure:"skillPath"`
}

type WorkflowTemplatePathConfig struct {
	ServerMode string `json:"server_mode" mapstructure:"server_mode"`
	ConfigPath string `json:"configPath" mapstructure:"configPath"`

	GlobalWebListUrl string `json:"global_web_list_url" mapstructure:"global_web_list_url"`

	ListUrl      string `json:"list_url" mapstructure:"list_url"`
	DownloadUrl  string `json:"download_url" mapstructure:"download_url"`
	DetailUrl    string `json:"detail_url" mapstructure:"detail_url"`
	RecommendUrl string `json:"recommend_url" mapstructure:"recommend_url"`
}

type PromptTemplatePathConfig struct {
	ConfigPath string `json:"configPath" mapstructure:"configPath"`
}

type SkillsTemplatePathConfig struct {
	ConfigPath string `json:"configPath" mapstructure:"configPath"`
}

type SkillCreatorPathConfig struct {
	ConfigPath string `json:"configPath" mapstructure:"configPath"`
}

type PromptTempConfig struct {
	TemplateId string `json:"templateId" mapstructure:"templateId"`
	Category   string `json:"category" mapstructure:"category"`
	Avatar     string `json:"avatar"`
	Name       string `json:"name"`
	Desc       string `json:"desc" mapstructure:"desc"`
	Author     string `json:"author" mapstructure:"author"`
	Prompt     string `json:"prompt" mapstructure:"prompt"`
}

type PromptEngineeringConfig struct {
	Optimization string `json:"optimization" mapstructure:"optimization"`
	Evaluation   string `json:"evaluation" mapstructure:"evaluation"`
}

type RecommendModelsByProvider struct {
	Provider       string                  `mapstructure:"provider" json:"provider"`
	Llm            []RecommendModelItemLLM `mapstructure:"llm" json:"llm"`
	Embedding      []RecommendModelItem    `mapstructure:"embedding" json:"embedding"`
	MultiEmbedding []RecommendModelItem    `mapstructure:"multimodal-embedding" json:"multimodal-embedding"`
	Rerank         []RecommendModelItem    `mapstructure:"rerank" json:"rerank"`
	MultiRerank    []RecommendModelItem    `mapstructure:"multimodal-rerank" json:"multimodal-rerank"`
	Ocr            []RecommendModelItemOCR `mapstructure:"ocr" json:"ocr"`
	Gui            []RecommendModelItem    `mapstructure:"gui" json:"gui"`
	PdfParser      []RecommendModelItem    `mapstructure:"pdf-parser" json:"pdf-parser"`
	SyncAsr        []RecommendModelItem    `mapstructure:"sync-asr" json:"sync-asr"`
}

type RecommendModelConfig []RecommendModelsByProvider

type WorkflowServiceConfig struct {
	Endpoint           string `json:"endpoint" mapstructure:"endpoint"`
	MinioProxyEndpoint string `json:"minio_proxy_endpoint" mapstructure:"minio_proxy_endpoint"`
	MinioProxyPrefix   string `json:"minio_proxy_prefix" mapstructure:"minio_proxy_prefix"`
	// general
	ListUri    string `json:"list_uri" mapstructure:"list_uri"`
	CreateUri  string `json:"create_uri" mapstructure:"create_uri"`
	DeleteUri  string `json:"delete_uri" mapstructure:"delete_uri"`
	CopyUri    string `json:"copy_uri" mapstructure:"copy_uri"`
	ExportUri  string `json:"export_uri" mapstructure:"export_uri"`
	ImportUri  string `json:"import_uri" mapstructure:"import_uri"`
	ConvertUri string `json:"convert_uri" mapstructure:"convert_uri"`
	// schema
	ListSchemaUri string `json:"list_schema_uri" mapstructure:"list_schema_uri"`
	// run
	WorkflowRunByOpenapiUri     string `json:"workflow_run_by_openapi_uri" mapstructure:"workflow_run_by_openapi_uri"`
	WorkflowRunLatestVersionUri string `json:"workflow_run_latest_version_uri" mapstructure:"workflow_run_latest_version_uri"`
	GetProcessUri               string `json:"get_process_uri" mapstructure:"get_process_uri"`
	ChatflowRunByOpenapiUri     string `json:"chatflow_run_by_openapi_uri" mapstructure:"chatflow_run_by_openapi_uri"`
	// conversation
	CreateChatflowConversationUri string `json:"create_chatflow_conversation_uri" mapstructure:"create_chatflow_conversation_uri"`
	GetConversationMessageListUri string `json:"get_conversation_message_list_uri" mapstructure:"get_conversation_message_list_uri"`
	DeleteConversationUri         string `json:"delete_conversation_uri" mapstructure:"delete_conversation_uri"`
	GetDraftIntelligenceListUri   string `json:"get_draft_intelligence_list_uri" mapstructure:"get_draft_intelligence_list_uri"`
	GetDraftIntelligenceInfoUri   string `json:"get_draft_intelligence_info_uri" mapstructure:"get_draft_intelligence_info_uri"`
	GetProjectConversationUri     string `json:"get_project_conversation_uri" mapstructure:"get_project_conversation_uri"`
	GetProjectConversationListUri string `json:"get_project_conversation_list_uri" mapstructure:"get_project_conversation_list_uri"`
	// upload
	UploadActionUri string `json:"upload_action_uri" mapstructure:"upload_action_uri"`
	UploadCommonUri string `json:"upload_common_uri" mapstructure:"upload_common_uri"`
	SignImgUri      string `json:"sign_img_uri" mapstructure:"sign_img_uri"`
	// version
	PublishUri           string               `json:"publish_uri" mapstructure:"publish_uri"`
	VersionListUri       string               `json:"version_list_uri" mapstructure:"version_list_uri"`
	MultiVersionListUri  string               `json:"multi_version_list_uri" mapstructure:"multi_version_list_uri"`
	UpdateVersionDescUri string               `json:"update_version_desc_uri" mapstructure:"update_version_desc_uri"`
	RollbackUri          string               `json:"rollback_uri" mapstructure:"rollback_uri"`
	ModelParams          []WorkflowModelParam `json:"model_params" mapstructure:"model_params"`
}

type WorkflowModelParam struct {
	Name       string `json:"name" mapstructure:"name"`
	Desc       string `json:"desc" mapstructure:"desc"`
	Label      string `json:"label" mapstructure:"label"`
	Type       int    `json:"type" mapstructure:"type"`
	Precision  int    `json:"precision" mapstructure:"precision"`
	Min        string `json:"min" mapstructure:"min"`
	Max        string `json:"max" mapstructure:"max"`
	APIMode    int    `json:"api_mode" mapstructure:"api_mode"`
	CustomFlag bool   `json:"custom_flag" mapstructure:"custom_flag"`

	ParamClass WorkflowModelParamClass      `json:"param_class" mapstructure:"param_class"`
	DefaultVal WorkflowModelParamDefaultVal `json:"default_val" mapstructure:"default_val"`
	Options    []WorkflowModelParamOption   `json:"options" mapstructure:"options"`
}

type WorkflowModelParamOption struct {
	Label string `json:"label" mapstructure:"label"`
	Value string `json:"value" mapstructure:"value"`
}

type WorkflowModelParamClass struct {
	ClassID int    `json:"class_id" mapstructure:"class_id"`
	Label   string `json:"label" mapstructure:"label"`
}

type WorkflowModelParamDefaultVal struct {
	Precise    string `json:"precise" mapstructure:"precise"`
	Balance    string `json:"balance" mapstructure:"balance"`
	Creative   string `json:"creative" mapstructure:"creative"`
	DefaultVal string `json:"default_val" mapstructure:"default_val"`
}

type AssistantTemplateConfig struct {
	ConfigPath string `json:"configPath" mapstructure:"configPath"`
}

type DocCenterConfig struct {
	FrontendPrefix string          `json:"frontend_prefix" mapstructure:"frontend_prefix"`
	Links          []DocLinkConfig `json:"links" mapstructure:"links"`
	docs           map[string]string
}

type DocLinkConfig struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

type CustomInfoConfig struct {
	DefaultMode          string        `json:"default_mode" mapstructure:"default_mode"`
	Modes                []CustomTheme `json:"modes" mapstructure:"modes"`
	Version              string        `json:"version" mapstructure:"version"`
	ReleaseNotesPath     string        `json:"release_notes_path" mapstructure:"release_notes_path"`
	ReleaseNotes         string        `json:"-" mapstructure:"-"`
	RegisterByEmail      int           `json:"register_by_email" mapstructure:"register_by_email"`
	ResetPasswordByEmail int           `json:"reset_password_by_email" mapstructure:"reset_password_by_email"`
	LoginByEmail         int           `json:"login_by_email" mapstructure:"login_by_email"`
	UserPhoneRequired    int           `json:"user_phone_required" mapstructure:"user_phone_required"`
	PasswordRSAEncrypt   int           `json:"password_rsa_encrypt" mapstructure:"password_rsa_encrypt"` // 密码RSA加密开关 0-关闭 1-开启
}

type CustomTheme struct {
	Mode         string             `json:"mode" mapstructure:"mode"`
	Login        CustomLogin        `json:"login" mapstructure:"login"`
	Home         CustomHome         `json:"home" mapstructure:"home"`
	Tab          CustomTab          `json:"tab" mapstructure:"tab"`
	About        CustomAbout        `json:"about" mapstructure:"about"`
	GeneralAgent CustomGeneralAgent `json:"general_agent" mapstructure:"general_agent"`
}

type CustomLogin struct {
	BackgroundPath   string `json:"background_path" mapstructure:"background_path"`
	LogoPath         string `json:"logo_path" mapstructure:"logo_path"`
	LoginButtonColor string `json:"login_button_color" mapstructure:"login_button_color"`
	WelcomeText      string `json:"welcome_text" mapstructure:"welcome_text"`
	PlatformDesc     string `json:"platform_desc" mapstructure:"platform_desc"`
}

type CustomHome struct {
	LogoPath        string `json:"logo_path" mapstructure:"logo_path"`
	Title           string `json:"title" mapstructure:"title"`
	BackgroundColor string `json:"background_color" mapstructure:"background_color"`
}

type CustomTab struct {
	TabTitle    string `json:"title" mapstructure:"title"`
	TabLogoPath string `json:"logo_path" mapstructure:"logo_path"`
}

type CustomAbout struct {
	LogoPath  string `json:"logo_path" mapstructure:"logo_path"`
	Copyright string `json:"copyright" mapstructure:"copyright"`
}

type CustomGeneralAgent struct {
	LogoPath    string `json:"logo_path" mapstructure:"logo_path"`
	WelcomeText string `json:"welcome_text" mapstructure:"welcome_text"`
}

type DefaultIconConfig struct {
	UserIcon         string `json:"user" mapstructure:"user"`
	RagIcon          string `json:"rag" mapstructure:"rag"`
	AgentIcon        string `json:"agent" mapstructure:"agent"`
	WorkflowIcon     string `json:"workflow" mapstructure:"workflow"`
	ChatflowIcon     string `json:"chatflow" mapstructure:"chatflow"`
	McpCustomIcon    string `json:"mcpCustom" mapstructure:"mcpCustom"`
	McpServerIcon    string `json:"mcpServer" mapstructure:"mcpServer"`
	ToolIcon         string `json:"tool" mapstructure:"tool"`
	PromptIcon       string `json:"prompt" mapstructure:"prompt"`
	SkillIcon        string `json:"skill" mapstructure:"skill"`
	CustomSkillIcon  string `json:"skillCustom" mapstructure:"skillCustom"`
	KnowledgeIcon    string `json:"knowledge" mapstructure:"knowledge"`
	QAIcon           string `json:"qa" mapstructure:"qa"`
	ModelIcon        string `json:"model" mapstructure:"model"`
	GeneralAgentIcon string `json:"generalAgent" mapstructure:"generalAgent"`
	OrgIcon          string `json:"org" mapstructure:"org"`
	RoleIcon         string `json:"role" mapstructure:"role"`
	SystemIcon       string `json:"system" mapstructure:"system"`
	AdminIcon        string `json:"admin" mapstructure:"admin"`
}

func LoadConfig(in string) error {
	_c = &Config{}
	if err := util.LoadConfig(in, _c); err != nil {
		return err
	}
	_c.DocCenter.docs = make(map[string]string)
	for _, link := range _c.DocCenter.Links {
		url, _ := url.JoinPath(_c.Server.WebBaseUrl, _c.DocCenter.FrontendPrefix, url.PathEscape(link.Val))
		_c.DocCenter.docs[link.Key] = url
	}
	// 加载工作流模板配置
	if err := util.LoadConfig(_c.WorkflowTemplate.ConfigPath, _c); err != nil {
		return fmt.Errorf("load workflow template config err: %v", err)
	}
	for _, wt := range _c.WorkflowTemplates {
		if err := wt.load(); err != nil {
			return err
		}
	}
	// 加载提示词模板配置
	promptIn := _c.PromptTemplate.ConfigPath
	if err := util.LoadConfig(promptIn, _c); err != nil {
		return fmt.Errorf("load prompt template config err: %v", err)
	}
	// 加载skills模板配置
	skillsIn := _c.SkillsTemplate.ConfigPath
	if err := util.LoadConfig(skillsIn, _c); err != nil {
		return fmt.Errorf("load skills template config err: %v", err)
	}
	for _, st := range _c.AgentSkills {
		if err := st.load(); err != nil {
			return err
		}
	}
	// 加载 skill-creator 配置
	skillCreatorIn := _c.SkillCreatorPath.ConfigPath
	if err := util.LoadConfig(skillCreatorIn, _c); err != nil {
		return fmt.Errorf("load skill-creator config err: %v", err)
	}
	// 加载推荐模型配置
	recommendModelPath := _c.Model.RecommendModelConfigPath
	if recommendModelPath != "" {
		if err := util.LoadConfig(recommendModelPath, _c); err != nil {
			return fmt.Errorf("load recommend model config err: %v", err)
		}
	}
	// 加载 release notes
	if _c.CustomInfo.ReleaseNotesPath != "" {
		data, err := os.ReadFile(_c.CustomInfo.ReleaseNotesPath)
		if err != nil {
			log.Warnf("load release notes from %s err: %v", _c.CustomInfo.ReleaseNotesPath, err)
		} else {
			_c.CustomInfo.ReleaseNotes = string(data)
		}
	} else {
		log.Warnf("release_notes_path is empty, skip loading release notes")
	}
	// 初始化RSA密钥管理器
	if err := rsa_util.InitKeyManager(rsa_util.Config{
		PrivateKeyPath:  _c.RSAKey.PrivateKeyPath,
		PublicKeyPath:   _c.RSAKey.PublicKeyPath,
		KeyRotationDays: _c.RSAKey.KeyRotationDays,
	}); err != nil {
		return fmt.Errorf("init rsa key manager err: %v", err)
	}
	return nil
}

func Cfg() *Config {
	if _c == nil {
		log.Panicf("cfg nil")
	}
	return _c
}

func (c *Config) WorkflowTemp(templateId string) (WorkflowTemplateConfig, bool) {
	for _, wtf := range c.WorkflowTemplates {
		if wtf.TemplateId == templateId {
			return *wtf, true
		}
	}
	return WorkflowTemplateConfig{}, false
}

func (c *Config) PromptTemp(templateId string) (PromptTempConfig, bool) {
	for _, ptf := range c.PromptTemplates {
		if ptf.TemplateId == templateId {
			return *ptf, true
		}
	}
	return PromptTempConfig{}, false
}

func (c *Config) AgentSkill(skillId string) (SkillsConfig, bool) {
	for _, stf := range c.AgentSkills {
		if stf.SkillId == skillId {
			return *stf, true
		}
	}
	return SkillsConfig{}, false
}

// GetDocs 返回 docs 的深拷贝
func (d *DocCenterConfig) GetDocs() map[string]string {
	if d.docs == nil {
		return nil
	}
	// 深拷贝
	result := make(map[string]string, len(d.docs))
	for k, v := range d.docs {
		result[k] = v
	}
	return result
}
