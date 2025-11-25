package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/UnicomAI/wanwu/internal/iam-service/client/orm"
	"github.com/UnicomAI/wanwu/internal/iam-service/config"
	smtp_util "github.com/UnicomAI/wanwu/internal/iam-service/pkg/util/smtp-util"
	"github.com/UnicomAI/wanwu/internal/iam-service/server/grpc"
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/redis"
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
	flag.StringVar(&configFile, "config", "configs/microservice/iam-service/configs/config.yaml", "conf yaml file")
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

	if err := smtp_util.Init(config.Cfg().SMTP); err != nil {
		log.Fatalf("init email err: %v", err)
	}

	if err := redis.InitIAM(ctx, config.Cfg().Redis); err != nil {
		log.Fatalf("init redis err: %v", err)
	}

	db, err := db.New(config.Cfg().DB)
	if err != nil {
		log.Fatalf("init db err: %v", err)
	}

	c, err := orm.NewClient(db)
	if err != nil {
		log.Fatalf("init client err: %v", err)
	}

	// Initialize workflow database connection for space_user sync
	// Get MySQL connection details from config
	var mysqlAddress, mysqlPassword string
	switch config.Cfg().DB.DBName {
	case "mysql":
		mysqlAddress = config.Cfg().DB.MySQL.Address
		mysqlPassword = config.Cfg().DB.MySQL.Password
	case "tidb":
		mysqlAddress = config.Cfg().DB.TiDB.Address
		mysqlPassword = config.Cfg().DB.TiDB.Password
	case "oceanbase":
		mysqlAddress = config.Cfg().DB.OceanBase.Address
		mysqlPassword = config.Cfg().DB.OceanBase.Password
	}

	if mysqlAddress != "" && mysqlPassword != "" {
		if err := orm.InitWorkflowDB(mysqlAddress, mysqlPassword); err != nil {
			log.Warnf("init workflow db connection err (space_user sync will be disabled): %v", err)
		}
	} else {
		log.Warnf("MySQL connection details not available, workflow space_user sync will be disabled")
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
	s.Stop()
	redis.StopIAM()
}

func versionPrint() {
	fmt.Printf("build_time: %s\n", buildTime)
	fmt.Printf("build_version: %s\n", buildVersion)
	fmt.Printf("git_commit_id: %s\n", gitCommitID)
	fmt.Printf("git branch: %s\n", gitBranch)
	fmt.Printf("runtime version: %s\n", runtime.Version())
	fmt.Printf("builder: %s\n", builder)
}
