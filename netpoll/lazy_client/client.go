package main

import (
	"log"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		log.Fatal("dial failed:", err)
	}

	// 用于测试，设定单个完整协议包 2 字节，server端使用 connection.Reader.Next(2) 进行读取
	// server 端超时时间设定为 5s

	// 第一组，在超时时间内发送，server端能读到完整包，但 connection.Reader.Next(2) 阻塞 2s
	conn.Write([]byte("a"))
	time.Sleep(time.Second * 2)
	conn.Write([]byte("a"))

	time.Sleep(time.Second * 1)

	// 第二组，超过超时时间、分开多次发送完整包，server端 connection.Reader.Next(2) 阻塞 5s（超时时间为5s）后报错，但是连接没有断开
	// 30s 内、client 发送完剩余数据前，server 端多次触发 OnRequest，并且每次 OnRequest 中的 connection.Reader.Next(2) 阻塞 5s（超时时间为5s）后报错，但是连接没有断开
	// 30s 后 client 发送完整包剩余数据，server 端 connection.Reader.Next(2) 读到完整包
	conn.Write([]byte("b"))
	time.Sleep(time.Second * 30)
	conn.Write([]byte("b"))

	time.Sleep(time.Second * 1)

	// 第三组，只发送半包，client 端不再有行动
	// server 端多次触发 OnRequest，并且每次 OnRequest 中的 connection.Reader.Next(2) 阻塞 5s（超时时间为5s）后报错，但是连接没有断开
	// 实际场景中，server 端可能收不到 tcp FIN1，比如 client 设备断电，server 端无法及时迅速地释放该连接，如果大量连接进行攻击，存在服务不可用风险
	conn.Write([]byte("c"))

	<-make(chan int)
}
