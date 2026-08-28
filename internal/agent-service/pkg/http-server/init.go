package http_server

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/UnicomAI/wanwu/internal/agent-service/pkg"
	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/config"
	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/http-server/middleware"
	_ "github.com/UnicomAI/wanwu/internal/agent-service/pkg/tracer"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/gin-gonic/gin"
)

var (
	httpServ *http.Server
)

var ginHttpClient = GinHttpServer{}

type GinHttpServer struct {
	GinEngine *gin.Engine
}

func init() {
	pkg.AddContainer(ginHttpClient)
}

func (c GinHttpServer) LoadType() string {
	return "http-server"
}

func (c GinHttpServer) Load() error {
	// validator
	if err := gin_util.InitValidator(); err != nil {
		log.Fatalf("init gin validator err: %v", err)
	}

	// router
	gin.ForceConsoleColor()
	ginHttpClient.GinEngine = trace_util.NewTracerGin("agent-service")
	//初始化路由
	err := InitGinGroup(ginHttpClient.GinEngine, middleware.Record)
	if err != nil {
		return err
	}
	// start http server
	httpServ = &http.Server{
		Addr:    ":" + strconv.Itoa(config.GetConfig().Server.Port),
		Handler: ginHttpClient.GinEngine,
	}
	go func() {
		if err := httpServ.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server fatal: %v", err)
		}
	}()
	//log.Infof("server listen on: %v", config.Cfg().Server.Port)
	return nil
}

func (c GinHttpServer) Stop() error {
	log.Infof("closing http server...")
	// stop http server
	cancelCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := httpServ.Shutdown(cancelCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	} else {
		log.Infof("close http server gracefully")
	}
	return nil
}

func (c GinHttpServer) StopPriority() int {
	return pkg.DefaultPriority
}
