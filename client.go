package main

import (
	"flag"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lesismal/nbio"
)

var (
	addr = "localhost:8888"

	clientNum = flag.Int("c", 20000, "client num")
	testTime  = flag.Int("t", 30, "total test time")
	bufsize   = flag.Int("b", 64, "data size")
)

func main() {
	var (
		wg         sync.WaitGroup
		qps        int64
		totalRead  int64
		totalWrite int64
	)
	g := nbio.NewGopher(nbio.Config{})
	defer g.Stop()

	g.OnData(func(c *nbio.Conn, data []byte) {
		atomic.AddInt64(&qps, 1)
		atomic.AddInt64(&totalRead, int64(len(data)))
		atomic.AddInt64(&totalWrite, int64(len(data)))
		c.Write(append([]byte{}, data...))
	})

	err := g.Start()
	if err != nil {
		log.Printf("Start failed: %v\n", err)
	}

	log.Printf("loat test for %v connections, buffer size: %v", *clientNum, *bufsize)
	t := time.Now()
	for i := 0; i < *clientNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data := make([]byte, *bufsize)
			c, err := nbio.Dial("tcp", addr)
			if err != nil {
				log.Printf("Dial failed: %v", err)
			}
			g.AddConn(c)
			c.Write([]byte(data))
			atomic.AddInt64(&totalWrite, int64(len(data)))
		}()
	}
	wg.Wait()
	t2 := time.Since(t)
	log.Printf("%v clients connected, used: %v s, %v ns/op", *clientNum, t2.Seconds(), t2.Nanoseconds()/int64(*clientNum))

	//warm up
	log.Println("warm up for 5s...")
	time.Sleep(time.Second * 5)
	log.Println("warm up over, start io statistics")

	t = time.Now()
	atomic.SwapInt64(&totalRead, 0)
	atomic.SwapInt64(&totalWrite, 0)

	time.Sleep(time.Second * time.Duration(*testTime))
	log.Printf("%vs, total read: %.2f M, total write: %.2f M", *testTime, float64(atomic.LoadInt64(&totalRead))/1024/1024, float64(atomic.LoadInt64(&totalWrite))/1024/1024)

	log.Println(g.State().String())

}
