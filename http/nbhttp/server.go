package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/lesismal/nbio/nbhttp"
)

var (
	host       = flag.String("s", "127.0.0.1:8888", "listened ip and port")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	delay      = flag.Duration("delay", 0, "delay to mock business processing")
	async      = flag.Bool("a", false, "async response flag")
)

func onEcho(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	data := r.Body.(*nbhttp.BodyReader).RawBody()

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

	svr := nbhttp.NewServer(nbhttp.Config{
		Network: "tcp",
		Addrs:   []string{*host},
	}, router, nil) // pool.Go)

	err := svr.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	svr.Shutdown(ctx)
}
