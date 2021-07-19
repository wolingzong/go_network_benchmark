package main

import (
	"log"
	"net"
)

func handle(conn net.Conn) {
	buf := make([]byte, 1024*32)
	for {
		nread, err := conn.Read(buf)
		if err != nil {
			return
		}
		nwrite, err := conn.Write(append([]byte{}, buf[:nread]...))
		if err != nil {
			return
		}
		if nwrite != nread {
			return
		}
	}
}
func main() {
	addr := "localhost:8888"
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	log.Println("Running on:", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept failed:", err)
			return
		}
		go handle(conn)
	}
}
