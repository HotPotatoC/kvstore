package main

import (
	"flag"
	"fmt"

	"github.com/HotPotatoC/kvstore/cmd/kvstore_server/internal/server"
	"github.com/HotPotatoC/kvstore/pkg/logger"
	"go.uber.org/zap"

	"net/http"
	_ "net/http/pprof"
)

var Version = "dev"
var Build = "dev"

var log *zap.SugaredLogger

var (
	host  = flag.String("host", "0.0.0.0", "KVStore server host")
	port  = flag.Int("port", 7275, "KVStore server port")
	debug = flag.Bool("debug", false, "Debug mode")
)

func init() {
	flag.StringVar(host, "h", "0.0.0.0", "KVStore server host")
	flag.IntVar(port, "p", 7275, "KVStore server port")
	flag.BoolVar(debug, "d", false, "Debug mode")
}

func init() {
	log = logger.NewLogger()
}

func main() {
	flag.Parse()

	server := server.New(Version, Build)

	if *debug {
		log.Info("-=-=-=-=-=-= Running in debug mode =-=-=-=-=-=-")
		go func() {
			log.Infof("Pprof started -> http://%s:%d/debug/pprof", *host, *port+1)
			if err := http.ListenAndServe(fmt.Sprintf("%s:%d", *host, *port+1), nil); err != nil {
				log.Fatalf("pprof failed: %v", err)
			}
		}()
	}

	server.Start(*host, *port)
}
