package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/HotPotatoC/kvstore-rewrite/config"
	"github.com/HotPotatoC/kvstore-rewrite/logger"
	"github.com/HotPotatoC/kvstore-rewrite/server"
)

var (
	debug     = flag.Bool("debug", false, "enable debug mode")
	cfg       = flag.String("config", "", "config file")
)

func init() {
	flag.BoolVar(debug, "d", false, "enable debug mode")
	flag.StringVar(cfg, "c", "", "config file")
}

func main() {
	flag.Parse()
	logger.Init(*debug)

	if err := config.Load(*cfg); err != nil {
		logger.S().Fatal("load config failed: ", err)
	}

	srv, err := server.New()
	if err != nil {
		logger.S().Fatal("failed instantiating server: ", err)
	}

	logger.S().Info("Starting kvstore server")
	go func() {
		if err := srv.Run(); err != nil {
			logger.S().Fatal("failed to run server", err)
		}
	}()

	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	logger.S().Warn("Shutting down kvstore server")

	srv.Stop()
}
