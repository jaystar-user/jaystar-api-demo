package main

import (
	"jaystar/internal/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	appServer := server.NewAppServer()

	appServer.Init()

	appServer.Run()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	appServer.Stop()
}
