package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/HotPotatoC/kvstore/internal/config"
	"github.com/HotPotatoC/kvstore/internal/logger"
	"github.com/HotPotatoC/kvstore/internal/server"
	"github.com/HotPotatoC/kvstore/internal/util"
	"github.com/HotPotatoC/kvstore/internal/version"
	"github.com/panjf2000/gnet"
	"github.com/spf13/viper"

	"net/http"
	_ "net/http/pprof"
)

var (
	debug = flag.Bool("debug", false, "Debug mode")
	cfg   = flag.String("cfg", "", "kvstore yaml configuration path")
)

func init() {
	flag.BoolVar(debug, "d", false, "Debug mode")
	flag.StringVar(cfg, "c", "", "kvstore yaml configuration path")
}

func main() {
	flag.Parse()

	if err := config.Load(*cfg); err != nil {
		logger.S().Fatalf("failed loading config file: %v", err)
	}

	logger.Init(*debug)

	server := server.New(version.Version, version.Build)

	if *debug {
		logger.S().Debug("-=-=-=-=-=-= Running in debug mode =-=-=-=-=-=-")
		go func() {
			logger.S().Debugf("Pprof started -> http://%s:%d/debug/pprof",
				viper.GetString("server.host"),
				viper.GetInt("server.port")+1)

			if err := http.ListenAndServe(
				fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.port")+1), nil); err != nil {
				logger.S().Fatalf("pprof failed: %v", err)
			}
		}()
	}

	go func() {
		if err := server.Start(); err != nil {
			logger.S().Fatal(err)
		}
	}()

	recv := <-util.WaitForSignals(os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	logger.S().Debugf("received interrupt signal: %s", recv)

	gnet.Stop(context.Background(), fmt.Sprintf("%s://%s:%d",
		viper.GetString("server.protocol"),
		viper.GetString("server.host"),
		viper.GetInt("server.port")))
}
