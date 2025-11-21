package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg"
	_ "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/minio"
	_ "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/tool-init"
	_ "github.com/UnicomAI/wanwu/internal/knowledge-service/server/grpc"
	_ "github.com/UnicomAI/wanwu/internal/knowledge-service/task"
)

var (
	isVersion    bool
	buildTime    string //编译时间 (build time) [EN] build time
	buildVersion string //编译版本 (build version) [EN] build version
	gitCommitID  string //git的commit id (git commit id) [EN] git commit id (git commit id)
	gitBranch    string //git branch (git branch)
	builder      string //构建者 (builder) [EN] builder
)

func main() {
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
	err := pkg.InitAllService()
	if err != nil {
		panic(err)
	}
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc
	pkg.StopAllService()
}

func versionPrint() {
	fmt.Printf("build_time: %s\n", buildTime)
	fmt.Printf("build_version: %s\n", buildVersion)
	fmt.Printf("git_commit_id: %s\n", gitCommitID)
	fmt.Printf("git branch: %s\n", gitBranch)
	fmt.Printf("runtime version: %s\n", runtime.Version())
	fmt.Printf("builder: %s\n", builder)
}
