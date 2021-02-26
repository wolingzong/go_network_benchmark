package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/panjf2000/gnet"
)

type echoServer struct {
	*gnet.EventServer
}

func (es *echoServer) OnInitComplete(srv gnet.Server) (action gnet.Action) {
	log.Printf("Echo server is listening on %s (multi-cores: %t, loops: %d)\n",
		srv.Addr.String(), srv.Multicore, srv.NumEventLoop)
	return
}

func (es *echoServer) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	// Echo synchronously.
	out = append([]byte{}, frame...)
	return

	/*
		// Echo asynchronously.
		data := append([]byte{}, frame...)
		go func() {
			time.Sleep(time.Second)
			c.AsyncWrite(data)
		}()
		return
	*/
}

func main() {
	echo := new(echoServer)

	go func() {
		for {
			time.Sleep(time.Second)
			log.Println("goroutine num:", runtime.NumCPU())
		}
	}()

	log.Fatal(gnet.Serve(echo, fmt.Sprintf("tcp://:%d", 8888), gnet.WithMulticore(true), gnet.WithReusePort(false)))
}
