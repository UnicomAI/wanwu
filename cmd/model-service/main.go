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

	"github.com/UnicomAI/wanwu/internal/model-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/model-service/config"
	"github.com/UnicomAI/wanwu/internal/model-service/server/grpc"
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
)

var (
	configFile   = flag.String("config", "configs/microservice/model-service/configs/config.yaml", "model-service config")
	isVersion    bool
	buildTime    string //编译时间
	buildVersion string //编译版本
	gitCommitID  string //git的commit id
	gitBranch    string //git branch
	builder      string //构建者

)

func main() {
	//打印编译信息
	flag.BoolVar(&isVersion, "v", false, "编译信息")
	flag.Parse()

	if len(os.Args) > 1 && !isVersion {
		flag.Usage()
		return
	}

	if isVersion {
		versionPrint()
		return
	}

	ctx := context.Background()

	flag.Parse()
	if err := config.LoadConfig(*configFile); err != nil {
		log.Fatalf("init cfg err: %v", err)
	}

	if err := log.InitLog(config.Cfg().Log.Std, config.Cfg().Log.Level, config.Cfg().Log.Logs...); err != nil {
		log.Fatalf("init log err: %v", err)
	}

	// init tracer
	if err := trace_util.InitTracer("model-service"); err != nil {
		log.Fatalf("init tracer err: %v", err)
	}

	db, err := db.New(config.Cfg().DB)
	if err != nil {
		log.Fatalf("init db failed, err: %v", err)
	}

	c, err := orm.NewClient(ctx, db)
	if err != nil {
		log.Fatalf("init client failed, err: %v", err)
	}

	s, err := grpc.NewServer(config.Cfg(), c)
	if err != nil {
		log.Fatalf("init server failed, err: %v", err)
	}
	if err := s.Start(ctx); err != nil {
		log.Fatalf("start grpc server failed, err: %s", err)
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
