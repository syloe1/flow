package main

import (
	"log"
	"net"
)

func main() {
	log.Println("尝试连接服务端....")
	conn, err := net.Dial("tcp", "127.0.0.1:3333")
	if err != nil {
		log.Fatalf("连接服务端失败: %v", err)
	}
	log.Printf("连接服务端成功: %s <-> %s", conn.LocalAddr(), conn.RemoteAddr())
	defer func() {
		conn.Close()
		log.Printf("连接关闭: %v", conn.RemoteAddr())
	}()
}
