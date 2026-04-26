package main

import (
	"fmt"
	"net"
)

func main() {
	// 监听UDP端口 :8888
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 8888,
	})
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buf := make([]byte, 4096)

	fmt.Println("UDP服务端启动 :8888")

	// 循环实时接收
	for {
		// 阻塞读取实时数据
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		fmt.Printf("收到 %s: %s\n", addr, string(buf[:n]))
	}
}
