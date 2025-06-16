package server

import (
	"context"
	"github.com/SeanZhenggg/go-utils/logger"
	"github.com/gin-gonic/gin"
	"jaystar/internal/app/job"
	"jaystar/internal/app/web"
	"jaystar/internal/config"
	"log"
	"net/http"
	"time"
)

type appServer struct {
	gin       *gin.Engine  `wire:"-"`
	server    *http.Server `wire:"-"`
	iWebApp   web.IWebApp
	job       job.IJob
	configEnv config.IConfigEnv
	logger    logger.ILogger
}

func (app *appServer) Init() {
	if app.configEnv.GetAppEnv() == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	app.gin = gin.New()
	app.gin.Use(gin.Recovery())

	app.iWebApp.Init(app.gin)
	app.job.Init()
}

func (app *appServer) Run() {
	app.job.Start()

	port := app.configEnv.GetHttpConfig().Port
	address := ":" + port

	app.server = &http.Server{
		Addr:         address,
		Handler:      app.gin.Handler(),
		ErrorLog:     log.Default(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		if err := app.server.ListenAndServe(); err != http.ErrServerClosed {
			app.logger.Error(context.TODO(), "server listen", err)
		}
	}()
	// TLS connection
	//err := app.gin.RunTLS(address, "./localhost+2.pem", "./localhost+2-key.pem")
}

func (app *appServer) Stop() {
	app.logger.Info(context.TODO(), "server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	const waitingProcesses = 2
	ch := make(chan struct{}, waitingProcesses)

	go func() {
		app.job.Stop(ctx)
		ch <- struct{}{}
	}()

	go func() {
		if err := app.server.Shutdown(ctx); err != nil {
			app.logger.Error(ctx, "server Shutdown error", err)
		}
		ch <- struct{}{}
	}()

	for i := 0; i < waitingProcesses; i++ {
		<-ch
	}

	app.logger.Info(context.TODO(), "server stopped")
}
