package main

import (
	"context"
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	host       = flag.String("s", "127.0.0.1:8888", "listened ip and port")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	delay      = flag.Duration("delay", 0, "delay to mock business processing")
	async      = flag.Bool("a", false, "async response flag")
)

func onEcho(ctx *fasthttp.RequestCtx) {
	data := ctx.PostBody()

	if *delay > 0 {
		time.Sleep(*delay)
	} else {
		runtime.Gosched()
	}

	if len(data) > 0 {
		ctx.Write(data)
	} else {
		ctx.Write([]byte(time.Now().Format("20060102 15:04:05")))
	}
}

func main() {
	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil {
			panic(err)
		}
	}()

	go fasthttp.ListenAndServe("localhost:8888", onEcho)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	_, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
}
