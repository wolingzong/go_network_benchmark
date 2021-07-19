package main

import (
	"flag"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	addr = "localhost:8888"

	clientNum       = flag.Int("c", 1024*5, "client num")
	stickyClientNum = flag.Int("s", 1024*5, "sticky client num")
	lazySend        = flag.Bool("l", true, "send data after all clients connected")
	testTime        = flag.Int("t", 30, "total test time")
	bufsize         = flag.Int("b", 1024, "data size")

	totalRead  int64
	totalWrite int64
)

func main() {
	flag.Parse()

	log.Printf("loat test for %v normal connections and %v sticky connections, buffer size: %v", *clientNum, *stickyClientNum, *bufsize)

	t := time.Now()
	wg := sync.WaitGroup{}
	mux := sync.Mutex{}
	conns := []net.Conn{}
	lazyConns := []net.Conn{}
	for i := 0; i < *clientNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				c, err := net.Dial("tcp", addr)
				if err != nil {
					time.Sleep(time.Millisecond * 20)
					continue
				}
				if *lazySend {
					mux.Lock()
					conns = append(conns, c)
					mux.Unlock()
				} else {
					go handle(c)
				}
				break
			}
		}()
	}
	for i := 0; i < *stickyClientNum; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				c, err := net.Dial("tcp", addr)
				if err != nil {
					time.Sleep(time.Millisecond * 20)
					continue
				}
				if *lazySend {
					mux.Lock()
					lazyConns = append(lazyConns, c)
					mux.Unlock()
				} else {
					go handleLazy(c)
				}
				break
			}
		}()
	}
	wg.Wait()
	t2 := time.Since(t)
	log.Printf("%v clients connected, used: %v s, %v ns/op", *clientNum, t2.Seconds(), t2.Nanoseconds()/int64(*clientNum))

	if *lazySend {
		for _, c := range conns {
			go handle(c)
		}
		for _, c := range lazyConns {
			go handleLazy(c)
		}
	}

	//warm up
	log.Println("warm up for 5s...")
	time.Sleep(time.Second * 5)
	log.Println("warm up over, start io statistics")

	t = time.Now()
	atomic.SwapInt64(&totalRead, 0)
	atomic.SwapInt64(&totalWrite, 0)

	time.Sleep(time.Second * time.Duration(*testTime))
	log.Printf("%vs, total read: %.2f M, total write: %.2f M", *testTime, float64(atomic.LoadInt64(&totalRead))/1024/1024, float64(atomic.LoadInt64(&totalWrite))/1024/1024)
}

func handle(conn net.Conn) {
	buf := make([]byte, *bufsize)
	for {
		nwrite, err := conn.Write(buf)
		if err != nil {
			return
		}
		atomic.AddInt64(&totalWrite, int64(nwrite))

		nread, err := io.ReadFull(conn, buf)
		if err != nil {
			return
		}
		atomic.AddInt64(&totalRead, int64(nread))
		if nwrite != nread {
			return
		}
	}
}

func handleLazy(conn net.Conn) {
	buf := make([]byte, *bufsize)
	for {
		nwrite := 0
		for i := 0; i < len(buf)-1; i++ {
			n, err := conn.Write(buf[i : i+1])
			if err != nil {
				return
			}
			nwrite += n
			time.Sleep(time.Second * 5)
		}
		n, err := conn.Write(buf[len(buf)-1:])
		if err != nil {
			return
		}
		nwrite += n
		atomic.AddInt64(&totalWrite, int64(nwrite))

		nread, err := io.ReadFull(conn, buf)
		if err != nil {
			return
		}
		atomic.AddInt64(&totalRead, int64(nread))
		if nwrite != nread {
			return
		}
	}
}
