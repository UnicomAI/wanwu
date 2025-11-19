package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/pkg/ahocorasick"
	mcp_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/mcp-util"
	oauth2_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/oauth2-util"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler"
	"github.com/UnicomAI/wanwu/pkg/i18n"
	jwt_util "github.com/UnicomAI/wanwu/pkg/jwt-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	"github.com/UnicomAI/wanwu/pkg/redis"
	"github.com/UnicomAI/wanwu/pkg/util"
)

var (
	configFile string
	isVersion  bool

	buildTime    string //编译时间 (build time) [EN] build time
	buildVersion string //编译版本 (build version) [EN] build version
	gitCommitID  string //git的commit id (git commit id) [EN] git commit id (git commit id)
	gitBranch    string //git branch (git branch)
	builder      string //构建者 (builder) [EN] builder
)

func main() {
	flag.StringVar(&configFile, "config", "configs/microservice/bff-service/configs/config.yaml", "conf yaml file")
	flag.BoolVar(&isVersion, "v", false, "build message")
	flag.Parse()

	if isVersion {
		versionPrint()
		return
	}

	ctx := context.Background()

	// config
	flag.Parse()
	if err := config.LoadConfig(configFile); err != nil {
		log.Fatalf("init cfg err: %v", err)
	}

	// init log
	if err := log.InitLog(config.Cfg().Log.Std, config.Cfg().Log.Level, config.Cfg().Log.Logs...); err != nil {
		log.Fatalf("init log err: %v", err)
	}

	// init time local
	if err := util.InitTimeLocal(); err != nil {
		log.Fatalf("init time local UTC8 err: %v", err)
	}

	// init i18n
	if err := i18n.Init(config.Cfg().I18n); err != nil {
		log.Fatalf("init i18n err: %v", err)
	}

	// init aho-corasick
	if err := ahocorasick.Init(true); err != nil {
		log.Fatalf("init aho-corasick err: %v", err)
	}

	// init jwt
	if err := jwt_util.InitUserJWT(config.Cfg().JWT.SigningKey); err != nil {
		log.Errorf("init jwt err: %v", err)
	}

	// init minio: custom
	if err := minio.InitCustom(ctx, config.Cfg().Minio); err != nil {
		log.Fatalf("init minio err: %v", err)
	}

	// init minio: fileupload
	if err := minio.InitFileUpload(ctx, config.Cfg().Minio); err != nil {
		log.Fatalf("init minio err: %v", err)
	}

	// init redis
	if err := redis.InitOP(ctx, config.Cfg().Redis); err != nil {
		log.Fatalf("init redis err: %v", err)
	}

	// init oauth2
	if config.Cfg().OAuth.Switch != 0 {
		issuer, err := url.JoinPath(config.Cfg().Server.ApiBaseUrl, "/openapi/v1")
		if err != nil {
			log.Errorf("init oauth issuer err: %v", err)
		}
		if err := oauth2_util.Init(redis.OP().Cli(), config.Cfg().OAuth.RSA, issuer, config.Cfg().JWT.SigningKey); err != nil {
			log.Fatalf("init oauth err: %v", err)
		}
	}

	// init model provider
	mp.Init(config.Cfg().Server.CallbackUrl)

	// init mcp server
	if err := mcp_util.Init(ctx); err != nil {
		log.Fatalf("init mcp server err: %v", err)
	}

	// start http handler
	handler.Start(ctx)

	// shutdown
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc

	// stop http handler
	handler.Stop(ctx)
	ahocorasick.Stop()
	redis.OP().Stop()
}

func versionPrint() {
	fmt.Printf("build_time: %s\n", buildTime)
	fmt.Printf("build_version: %s\n", buildVersion)
	fmt.Printf("git_commit_id: %s\n", gitCommitID)
	fmt.Printf("git branch: %s\n", gitBranch)
	fmt.Printf("runtime version: %s\n", runtime.Version())
	fmt.Printf("builder: %s\n", builder)
}
