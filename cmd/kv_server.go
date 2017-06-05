package main

import (
	"flag"
	"fmt"
	"junoKvServer"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var addr = flag.String("addr", ":8379", "listen address")
var usePprof = flag.Bool("pprof", true, "enable pprof")
var pprofPort = flag.Int("pprof_port", 6060, "pprof http port")

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	if *usePprof {
		go func() {
			log.Println(http.ListenAndServe(fmt.Sprintf(":%d", *pprofPort), nil))
		}()
	}

	var err error
	var server *junoKvServer.KVServer
	server, err = junoKvServer.NewKVServer(*addr)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Printf("juno-kv-server started %s", *addr)

	sc := make(chan os.Signal, 1)
	signal.Notify(
		sc,
		os.Kill,
		os.Interrupt,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go server.Start()

	<-sc

	log.Print("juno-kv-server is stoping")
	server.Stop()
	log.Print("ledis-server is stoped")
}
