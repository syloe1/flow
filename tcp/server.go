package main

import (
	"log"
	"net"
)

func main() {
	addr := "127.0.0.1:3333"
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}
	defer listener.Close()
	log.Println("服务端启动，等待客户端连接...")
	for {
		// 2. Accept() 阻塞等待客户端连接
		// 客户端 Dial 成功后，这里会返回连接 = 三次握手已完成
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept() 失败: %v", err)
			continue
		}
		log.Printf("Accept() 成功: %v", conn.RemoteAddr())
		handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer func() {
		conn.Close()
		log.Printf("连接关闭: %v", conn.RemoteAddr())
	}()
}
