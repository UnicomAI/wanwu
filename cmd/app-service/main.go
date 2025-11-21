package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/UnicomAI/wanwu/internal/app-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/app-service/config"
	"github.com/UnicomAI/wanwu/internal/app-service/server/grpc"
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
)

var (
	configFile string
	isVersion  bool

	buildTime    string // Build time
	buildVersion string // Build version
	gitCommitID  string // Git commit ID
	gitBranch    string // Git branch
	builder      string // Builder
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

	err := minio.InitSafety(ctx, minio.Config{
		Endpoint: config.Cfg().Minio.Endpoint,
		User:     config.Cfg().Minio.User,
		Password: config.Cfg().Minio.Password,
	}, config.Cfg().Minio.Bucket)
	if err != nil {
		log.Fatalf("init minio safety client err: %v", err)
	}

	db, err := db.New(config.Cfg().DB)
	if err != nil {
		log.Fatalf("init db err: %v", err)
	}

	c, err := orm.NewClient(db)
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
