package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/UnicomAI/wanwu/internal/assistant-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	"github.com/UnicomAI/wanwu/internal/assistant-service/server/grpc"
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/es"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	"github.com/UnicomAI/wanwu/pkg/redis"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
)

var (
	configFile string
	isVersion  bool

	buildTime    string //编译时间
	buildVersion string //编译版本
	gitCommitID  string //git的commit id
	gitBranch    string //git branch
	builder      string //构建者
)

func main() {
	flag.StringVar(&configFile, "config", "configs/microservice/assistant-service/configs/config.yaml", "conf yaml file")
	flag.BoolVar(&isVersion, "v", false, "build message")
	flag.Parse()

	if isVersion {
		versionPrint()
		return
	}

	ctx := context.Background()

	flag.Parse()
	if err := config.LoadConfig(configFile); err != nil {
		log.Fatalf("init cfg err: %v", err)
	}

	if err := log.InitLog(config.Cfg().Log.Std, config.Cfg().Log.Level, config.Cfg().Log.Logs...); err != nil {
		log.Fatalf("init log err: %v", err)
	}

	// init tracer
	if err := trace_util.InitTracer("assistant-service"); err != nil {
		log.Fatalf("init tracer err: %v", err)
	}

	if err := redis.InitSys(ctx, config.Cfg().Redis); err != nil {
		log.Fatalf("init redis err: %v", err)
	}

	if err := redis.InitAssistant(ctx, config.Cfg().Redis); err != nil {
		log.Fatalf("init assistant redis err: %v", err)
	}

	if err := es.InitAssistant(ctx, config.Cfg().ES); err != nil {
		log.Fatalf("init es err: %v", err)
	}

	if err := es.InitAsistantChatHistoryIndexTemplate(ctx); err != nil {
		log.Fatalf("init es index template err: %v", err)
	}

	if err := es.InitWgaChatHistoryEventIndexTemplate(ctx); err != nil {
		log.Fatalf("init wga chat history index template err: %v", err)
	}
	minioConfig := config.Cfg().Minio
	if err := minio.InitFileUpload(ctx, minio.Config{
		Endpoint:    minioConfig.EndPoint,
		User:        minioConfig.User,
		Password:    minioConfig.Password,
		DownloadURL: minioConfig.DownloadURL,
	}); err != nil {
		log.Fatalf("init minio err: %v", err)
	}

	db, err := db.New(config.Cfg().DB)
	if err != nil {
		log.Fatalf("init db err: %v", err)
	}

	// init model provider
	mp.Init(config.Cfg().Server.CallbackUrl)

	c, err := orm.NewClient(db)
	if err != nil {
		log.Fatalf("init client err: %v", err)
	}

	s, err := grpc.NewServer(config.Cfg(), c)
	if err != nil {
		log.Fatalf("init server err: %v", err)
	}
	if err := s.Start(); err != nil {
		log.Fatalf("start grpc server err: %s", err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc

	// flush trace spans
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	trace_util.ShutdownTracer(shutdownCtx)

	s.Stop()
	redis.StopSys()
	redis.StopAssistant()
	es.StopAssistant()
}

func versionPrint() {
	fmt.Printf("build_time: %s\n", buildTime)
	fmt.Printf("build_version: %s\n", buildVersion)
	fmt.Printf("git_commit_id: %s\n", gitCommitID)
	fmt.Printf("git branch: %s\n", gitBranch)
	fmt.Printf("runtime version: %s\n", runtime.Version())
	fmt.Printf("builder: %s\n", builder)
}
