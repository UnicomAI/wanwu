package assistant

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/spf13/viper"
)

var (
	_mrpc *templateCfg
)

type templateCfg struct {
	Assistants []*Assistant `json:"assistants" mapstructure:"assistants"`
}

type Assistant struct {
	Category          string   `json:"category" mapstructure:"category"`
	TemplateId        string   `json:"templateId" mapstructure:"templateId"`
	Avatar            string   `json:"avatar" mapstructure:"avatar"`
	Name              string   `json:"name" mapstructure:"name"`
	Desc              string   `json:"desc" mapstructure:"desc"`
	Prologue          string   `json:"prologue" mapstructure:"prologue"`
	Instructions      string   `json:"instructions" mapstructure:"instructions"`
	RecommendQuestion []string `json:"recommendQuestion" mapstructure:"recommendQuestion"`
	Summary           string   `json:"summary" mapstructure:"summary"`
	Feature           string   `json:"feature" mapstructure:"feature"`
	Scenario          string   `json:"scenario" mapstructure:"scenario"`
	WorkflowDesc      string   `json:"workflowDesc" mapstructure:"workflowDesc"`

	AvatarKey string // minio avatar objectPath
}

func Init(ctx context.Context) error {
	if _mrpc != nil {
		log.Panicf("already init")
	}
	// Load configuration
	configPath := config.Cfg().AssistantTemplate.ConfigPath
	if _, err := os.Stat(configPath); err != nil {
		return err
	}
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		return err
	}
	cfg := &templateCfg{}
	if err := v.Unmarshal(cfg); err != nil {
		return err
	}
	// Process pictures
	for _, assistant := range cfg.Assistants {
		// Check Picture - Local
		b, err := os.ReadFile(assistant.Avatar)
		if err != nil {
			return fmt.Errorf("read %v err: %v", assistant.Avatar, err)
		}
		// Check pictures - minio
		objectName := path.Join("avatar/agent", path.Base(assistant.Avatar))
		objectPath := path.Join(minio.BucketCustom, objectName)
		if _, err = minio.Custom().GetObject(ctx, minio.BucketCustom, objectName); err != nil {
			log.Warnf("check minio %v err: %v", objectPath, err)
			if _, err = minio.Custom().PutObject(ctx, minio.BucketCustom, objectName, b); err != nil {
				return fmt.Errorf("upload minio %v err: %v", objectPath, err)
			}
		}
		assistant.AvatarKey = objectPath
	}
	_mrpc = cfg
	return nil
}

func Cfg() []*Assistant {
	if _mrpc != nil {
		return _mrpc.Assistants
	}
	return nil
}
