package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/cloudwego/netpoll"
)

var (
	bufsize = flag.Int("b", 64, "data size")
)

func main() {
	flag.Parse()

	network, address := "tcp", "127.0.0.1:8888"

	// 创建 listener
	listener, err := netpoll.CreateListener(network, address)
	if err != nil {
		panic("create netpoll listener fail")
	}

	// handle: 连接读数据和处理逻辑
	var onRequest netpoll.OnRequest = handler

	// options: EventLoop 初始化自定义配置项
	var opts = []netpoll.Option{
		netpoll.WithReadTimeout(1 * time.Second),
		netpoll.WithIdleTimeout(10 * time.Minute),
		netpoll.WithOnPrepare(nil),
	}

	// 创建 EventLoop
	eventLoop, err := netpoll.NewEventLoop(onRequest, opts...)
	if err != nil {
		panic("create netpoll event-loop fail")
	}

	// 运行 Server
	go func() {
		err = eventLoop.Serve(listener)
		if err != nil {
			panic("netpoll server exit")
		}
	}()

	for {
		time.Sleep(time.Second)
		log.Println("goroutine num:", runtime.NumCPU())
	}
}

// 读事件处理
func handler(ctx context.Context, connection netpoll.Connection) error {
	reader := connection.Reader()
	l := *bufsize
	if l <= 0 {
		l = reader.Len()
	}
	buf, err := reader.Next(l)
	if err != nil {
		return err
	}

	n, err := connection.Write(append([]byte{}, buf...))
	// or
	// n, err := connection.Write(buf)

	if err != nil {
		return err
	}
	if n != len(buf) {
		return fmt.Errorf("write failed: %v < %v", n, len(buf))
	}

	return nil
}
