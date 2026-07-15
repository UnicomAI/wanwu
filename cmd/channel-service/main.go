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

	"github.com/UnicomAI/wanwu/internal/channel-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/channel-service/config"
	"github.com/UnicomAI/wanwu/internal/channel-service/server/grpc"
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/redis"
	"github.com/UnicomAI/wanwu/pkg/util"
)

var (
	configFile string
	isVersion  bool

	buildTime    string
	buildVersion string
	gitCommitID  string
	gitBranch    string
	builder      string
)

func main() {
	flag.StringVar(&configFile, "config", "configs/microservice/channel-service/configs/config.yaml", "conf yaml file")
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
	if err := trace_util.InitTracer("channel-service"); err != nil {
		log.Fatalf("init tracer err: %v", err)
	}

	if err := util.InitTimeLocal(); err != nil {
		log.Fatalf("init time local UTC8 err: %v", err)
	}

	// init redis
	if err := redis.InitOP(ctx, config.Cfg().Redis); err != nil {
		log.Fatalf("init redis err: %v", err)
	}

	// init db
	database, err := db.New(config.Cfg().DB)
	if err != nil {
		log.Fatalf("init db err: %v", err)
	}

	// init orm client
	c, err := orm.NewClient(database)
	if err != nil {
		log.Fatalf("init orm client err: %v", err)
	}

	// init grpc server（内部创建 adapter manager 和 chat handler）
	s, err := grpc.NewServer(config.Cfg(), c)
	if err != nil {
		log.Fatalf("init server err: %v", err)
	}
	if err := s.Start(ctx); err != nil {
		log.Fatalf("start grpc server err: %s", err)
	}

	log.Infof("channel-service started successfully")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc

	// graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	s.Stop(shutdownCtx)
	redis.OP().Stop()

	// flush trace spans
	trace_util.ShutdownTracer(shutdownCtx)
}

func versionPrint() {
	fmt.Printf("build_time: %s\n", buildTime)
	fmt.Printf("build_version: %s\n", buildVersion)
	fmt.Printf("git_commit_id: %s\n", gitCommitID)
	fmt.Printf("git branch: %s\n", gitBranch)
	fmt.Printf("runtime version: %s\n", runtime.Version())
	fmt.Printf("builder: %s\n", builder)
}
