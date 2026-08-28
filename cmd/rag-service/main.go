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

	"github.com/UnicomAI/wanwu/internal/rag-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/rag-service/config"
	"github.com/UnicomAI/wanwu/internal/rag-service/server/grpc"
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/es"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/UnicomAI/wanwu/pkg/redis"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
)

var (
	configFile   = flag.String("config", "configs/microservice/rag-service/configs/config.yaml", "rag-service config")
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

	if err := trace_util.InitTracer("rag-service"); err != nil {
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

	initRagES(ctx)

	initRagMinio(ctx)

	initRagRedis(ctx)

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
	es.StopRag()
	s.Stop(ctx)
}

// initRagRedis redis 用来接 bff 递过来的安全护栏回复
func initRagRedis(ctx context.Context) {
	if err := redis.InitRag(ctx, config.Cfg().Redis); err != nil {
		log.Fatalf("init rag redis err: %v", err)
	}
}

func initRagES(ctx context.Context) {
	if err := es.InitRag(ctx, config.Cfg().ES); err != nil {
		log.Fatalf("init es err: %v", err)
	}
	if err := es.InitRagChatHistoryIndexTemplate(ctx); err != nil {
		log.Fatalf("init es index template err: %v", err)
	}
}

func initRagMinio(ctx context.Context) {
	minioConfig := config.Cfg().Minio
	if minioConfig == nil {
		log.Fatalf("init minio err: minio config empty")
	}
	if err := minio.InitFileUpload(ctx, minio.Config{
		Endpoint:    minioConfig.EndPoint,
		User:        minioConfig.User,
		Password:    minioConfig.Password,
		DownloadURL: minioConfig.DownloadURL,
	}); err != nil {
		log.Fatalf("init minio err: %v", err)
	}
}

func versionPrint() {
	fmt.Printf("build_time: %s\n", buildTime)
	fmt.Printf("build_version: %s\n", buildVersion)
	fmt.Printf("git_commit_id: %s\n", gitCommitID)
	fmt.Printf("git branch: %s\n", gitBranch)
	fmt.Printf("runtime version: %s\n", runtime.Version())
	fmt.Printf("builder: %s\n", builder)
}
