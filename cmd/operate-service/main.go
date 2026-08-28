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

	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"

	"github.com/UnicomAI/wanwu/internal/operate-service/client/assistant"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/iam"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/operate-service/config"
	"github.com/UnicomAI/wanwu/internal/operate-service/server/grpc"
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
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
	flag.StringVar(&configFile, "config", "configs/microservice/operate-service/configs/config.yaml", "conf yaml file")
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
	// init time local
	if err := util.InitTimeLocal(); err != nil {
		log.Fatalf("init time local UTC8 err: %v", err)
	}

	if err := log.InitLog(config.Cfg().Log.Std, config.Cfg().Log.Level, config.Cfg().Log.Logs...); err != nil {
		log.Fatalf("init log err: %v", err)
	}

	// init tracer
	if err := trace_util.InitTracer("operate-service"); err != nil {
		log.Fatalf("init tracer err: %v", err)
	}

	db, err := db.New(config.Cfg().DB)
	if err != nil {
		log.Fatalf("init db err: %v", err)
	}

	// 消息中心的受众校验与 joinedAt 查询依赖 iam；懒连接、不阻塞启动
	iamCli, err := iam.NewClient(config.Cfg().Iam.Host)
	if err != nil {
		log.Fatalf("init iam client err: %v", err)
	}

	// 消息中心 actions 清洗依赖 assistant（agent 老 id → uuid）；懒连接、不阻塞启动
	assistantCli, err := assistant.NewClient(config.Cfg().Assistant.Host)
	if err != nil {
		log.Fatalf("init assistant client err: %v", err)
	}

	c, err := orm.NewClient(db, iamCli, assistantCli, config.Cfg().Notice)
	if err != nil {
		log.Fatalf("init client err: %v", err)
	}

	s, err := grpc.NewServer(config.Cfg(), c)
	if err != nil {
		log.Fatalf("init server err: %v", err)
	}
	if err := s.Start(ctx); err != nil {
		log.Fatalf("start grpc server err: %s", err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc

	// flush trace spans
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	trace_util.ShutdownTracer(shutdownCtx)

	s.Stop(ctx)
}

func versionPrint() {
	fmt.Printf("build_time: %s\n", buildTime)
	fmt.Printf("build_version: %s\n", buildVersion)
	fmt.Printf("git_commit_id: %s\n", gitCommitID)
	fmt.Printf("git branch: %s\n", gitBranch)
	fmt.Printf("runtime version: %s\n", runtime.Version())
	fmt.Printf("builder: %s\n", builder)
}
