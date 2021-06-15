package main

import (
	"fmt"
	"log"
	"runtime"
	"syscall"
	"time"

	"github.com/lesismal/nbio"
	"github.com/lesismal/nbio/taskpool"
)

var pool = taskpool.NewFixedPool(8, 1024)

func handleEvent(c *nbio.Conn, event int) {
	switch event {
	case nbio.PollerEventRead:
		pool.GoByIndex(c.Hash(), func() {
			buf := make([]byte, 1024*4)
			for {
				n, err := c.Read(buf)
				if err == syscall.EINTR {
					continue
				}
				if err == syscall.EAGAIN {
					return
				}
				if err != nil || n == 0 {
					c.CloseWithError(err)
					return
				}
				if n > 0 {
					c.Write(append([]byte{}, buf[:n]...))
				}
				if n < len(buf) {
					return
				}
			}
		})
	case nbio.PollerEventWrite:
		c.Flush()
	case nbio.PollerEventError:
		c.Close()
	default:
	}
}

func main() {
	g := nbio.NewGopher(nbio.Config{
		Network: "tcp",
		NPoller: runtime.NumCPU() * 1,
		Addrs:   []string{"localhost:8888"},
	})

	g.OnEvent(handleEvent)

	err := g.Start()
	if err != nil {
		fmt.Printf("nbio.Start failed: %v\n", err)
		return
	}

	for {
		time.Sleep(time.Second)
		log.Println("goroutine num:", runtime.NumCPU())
	}
}
