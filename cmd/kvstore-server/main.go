package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/HotPotatoC/kvstore-rewrite/config"
	"github.com/HotPotatoC/kvstore-rewrite/logger"
	"github.com/HotPotatoC/kvstore-rewrite/server"
	"github.com/panjf2000/gnet"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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

	logger.S().Info("Shutting down kvstore server")

	for _, addr := range viper.GetStringSlice("server.addrs") {
		if err := gnet.Stop(context.Background(), fmt.Sprintf("%s:%d", addr, viper.GetInt("server.port"))); err != nil {
			logger.S().Error("failed to stop server", zap.String("addr", addr), err)
		}
	}
}
