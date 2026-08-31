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

	"github.com/UnicomAI/wanwu/internal/app-service/task"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/internal/app-service/client/assistant"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/app-service/config"
	async_task "github.com/UnicomAI/wanwu/internal/app-service/pkg/async-task"
	"github.com/UnicomAI/wanwu/internal/app-service/server/grpc"
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/UnicomAI/wanwu/pkg/redis"
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
	flag.StringVar(&configFile, "config", "configs/microservice/app-service/configs/config.yaml", "conf yaml file")
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
	if err := trace_util.InitTracer("app-service"); err != nil {
		log.Fatalf("init tracer err: %v", err)
	}

	if err := util.InitTimeLocal(); err != nil {
		log.Fatalf("init time local err: %v", err)
	}

	err := minio.InitSafety(ctx, minio.Config{
		Endpoint: config.Cfg().Minio.Endpoint,
		User:     config.Cfg().Minio.User,
		Password: config.Cfg().Minio.Password,
	}, config.Cfg().Minio.Bucket)
	if err != nil {
		log.Fatalf("init minio safety client err: %v", err)
	}

	// 初始化文件上传客户端
	if err := minio.InitApp(ctx, minio.Config{
		Endpoint: config.Cfg().Minio.Endpoint,
		User:     config.Cfg().Minio.User,
		Password: config.Cfg().Minio.Password,
	}, config.Cfg().Minio.PublicExportBucket); err != nil {
		log.Fatalf("init minio file upload client err: %v", err)
	}

	if err := redis.InitApp(ctx, config.Cfg().Redis); err != nil {
		log.Fatalf("init redis err: %v", err)
	}

	db, err := db.New(config.Cfg().DB)
	if err != nil {
		log.Fatalf("init db err: %v", err)
	}

	// 数据清洗依赖 assistant（agent 老 id → uuid）；懒连接、不阻塞启动
	assistantCli, err := assistant.NewClient(config.Cfg().Assistant.Host)
	if err != nil {
		log.Fatalf("init assistant client err: %v", err)
	}

	c, err := orm.NewClient(ctx, db, assistantCli)
	if err != nil {
		log.Fatalf("init client err: %v", err)
	}

	// 初始化下游微服务 gRPC client
	if err := task.Init(c); err != nil {
		log.Fatalf("init grpc client err: %v", err)
	}
	if err := async_task.InitAsync(ctx, db); err != nil {
		log.Fatalf("init async task err: %v", err)
	}
	if err := orm.CronInit(ctx, db); err != nil {
		log.Fatalf("init cron failed, err: %v", err)
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
	async_task.StopAsync()
	orm.CronStop()
	redis.StopApp()
}

func versionPrint() {
	fmt.Printf("build_time: %s\n", buildTime)
	fmt.Printf("build_version: %s\n", buildVersion)
	fmt.Printf("git_commit_id: %s\n", gitCommitID)
	fmt.Printf("git branch: %s\n", gitBranch)
	fmt.Printf("runtime version: %s\n", runtime.Version())
	fmt.Printf("builder: %s\n", builder)
}
