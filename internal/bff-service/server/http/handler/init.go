package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/router/callback"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/router/openapi"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/router/openurl"
	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/router/v1"
	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/middleware"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/gin-gonic/gin"
)

var (
	httpServ, callbackServ *http.Server
)

func Start(ctx context.Context) {

	// router
	gin.ForceConsoleColor()

	// engin
	r := trace_util.NewTracerGin("bff-service")
	rcb := trace_util.NewTracerGin("bff-service-callback")

	// middleware
	middleware.Init(r)

	// validator
	if err := gin_util.InitValidator(); err != nil {
		log.Fatalf("init gin validator err: %v", err)
	}

	// v1
	v1.Register(r.Group("/v1"))
	// v2
	// v3
	// ..
	// openapi v1
	openapi.Register(r.Group("/openapi/v1"))
	// openurl v1
	openurl.Register(r.Group("/openurl/v1"))
	// callback v1
	callback.Register(rcb.Group("/callback/v1"))

	// service
	if err := service.Init(); err != nil {
		log.Fatalf("init service err: %v", err)
	}

	// doc-center
	if err := service.InitDocCenter(); err != nil {
		log.Fatalf("init doc-center err: %v", err)
	}

	// start mcp server
	if err := service.StartMCPServer(ctx); err != nil {
		log.Fatalf("start mcp server err: %v", err)
	}

	// start callback http server
	callbackServ = &http.Server{
		Addr:    ":" + strconv.Itoa(config.Cfg().Server.CallbackPort),
		Handler: rcb,
	}
	go func() {
		if err := callbackServ.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("callback fatal: %v", err)
		}
	}()
	log.Infof("callback listen on: %v", config.Cfg().Server.CallbackPort)
	// start http server
	httpServ = &http.Server{
		Addr:    ":" + strconv.Itoa(config.Cfg().Server.Port),
		Handler: r,
	}
	go func() {
		if err := httpServ.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server fatal: %v", err)
		}
	}()
	log.Infof("server listen on: %v", config.Cfg().Server.Port)

}

func Stop(ctx context.Context) {
	log.Infof("closing http server...")
	// stop http server
	cancelCtx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	if err := httpServ.Shutdown(cancelCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	} else {
		log.Infof("close http server gracefully")
	}
	if err := callbackServ.Shutdown(cancelCtx); err != nil {
		log.Fatalf("callback server forced to shutdown: %v", err)
	} else {
		log.Infof("close callback server gracefully")
	}
}
