package main

import (
	// "fmt"
	"bytes"
	"crypto/rand"
	"errors"
	"flag"
	"io"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/juju/ratelimit"
	"github.com/smallnest/rpcx/log"

	benchmark "github.com/rpcxio/rpcx-benchmark"
)

var total = flag.Int("n", 100000, "total requests for all clients")
var host = flag.String("s", "http://127.0.0.1:8888/echo", "server ip and port")
var pool = flag.Int("pool", 1000, "shared rpcx clients")
var rate = flag.Int("r", 10000, "throughputs")
var payload = make([]byte, 512)

func post(client *http.Client) error {
	req, err := http.NewRequest("POST", *host, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	// fmt.Println("post:", err)
	if err != nil {
		return err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if !bytes.Equal(payload, body) {
		return errors.New("not equal")
	}
	return nil
}

func main() {
	flag.Parse()

	tb := ratelimit.NewBucket(time.Second/time.Duration(*rate), int64(*rate))

	// 并发goroutine数.模拟客户端
	n := *pool
	// 每个客户端需要发送的请求数
	m := *total / n
	log.Infof("concurrency: %d\nrequests per client: %d\n\n", n, m)

	// 准备好参数
	rand.Read(payload)

	// 参数的大小
	log.Infof("message size: %d bytes\n\n", len(payload))

	// 等待所有测试完成
	var wg sync.WaitGroup
	wg.Add(n * m)

	// 创建客户端连接池
	var poolClients = make([]http.Client, *pool)
	// var pbCodec = &codec.ProtoBuffer{}
	for i := 0; i < *pool; i++ {
		client := &poolClients[i]
		for j := 0; j < 5; j++ {
			post(client)
		}
	}

	// 栅栏，控制客户端同时开始测试
	var startWg sync.WaitGroup
	startWg.Add(n + 1) // +1 是因为有一个goroutine用来记录开始时间

	// 总请求数
	var trans uint64
	// 返回正常的总请求数
	var transOK uint64

	// 每个goroutine的耗时记录
	d := make([][]int64, n, n)

	// 创建客户端 goroutine 并进行测试
	var startTime = time.Now().UnixNano()
	go func() {
		startWg.Done()
		startWg.Wait()
		startTime = time.Now().UnixNano()
	}()

	var clientIndex uint64
	for i := 0; i < *pool; i++ {
		dt := make([]int64, 0, m)
		d = append(d, dt)

		go func(i int) {
			startWg.Done()
			startWg.Wait()

			for j := 0; j < m; j++ {
				// 限流，这里不把限流的时间计算到等待耗时中
				tb.Wait(1)

				t := time.Now().UnixNano()
				ci := atomic.AddUint64(&clientIndex, 1)
				ci = ci % uint64(*pool)
				client := &poolClients[int(ci)]

				err := post(client)
				t = time.Now().UnixNano() - t // 等待时间+服务时间，等待时间是客户端调度的等待时间以及服务端读取请求、调度的时间，服务时间是请求被服务处理的实际时间

				d[i] = append(d[i], t)

				if err == nil {
					atomic.AddUint64(&transOK, 1)
				}

				atomic.AddUint64(&trans, 1)
				wg.Done()
			}
		}(i)
	}

	// 等待测试完成
	wg.Wait()

	// 统计
	benchmark.Stats(startTime, *total, d, trans, transOK)
}
