package main

import (
	"context"
	"flag"
	"io"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/julienschmidt/httprouter"
)

var (
	host       = flag.String("s", "127.0.0.1:8888", "listened ip and port")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	delay      = flag.Duration("delay", 0, "delay to mock business processing")
	async      = flag.Bool("a", false, "async response flag")
)

func onEcho(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	data, _ := io.ReadAll(r.Body)

	if *delay > 0 {
		time.Sleep(*delay)
	} else {
		runtime.Gosched()
	}

	if len(data) > 0 {
		w.Write(data)
	} else {
		w.Write([]byte(time.Now().Format("20060102 15:04:05")))
	}
}

func main() {
	go func() {
		if err := http.ListenAndServe("127.0.0.1:6060", nil); err != nil {
			panic(err)
		}
	}()

	router := httprouter.New()
	router.POST("/echo", onEcho)

	server := http.Server{
		Addr:    *host,
		Handler: router,
	}
	go server.ListenAndServe()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	server.Shutdown(ctx)
}
