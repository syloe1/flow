package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	// 连接目标UDP服务端
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:8888")
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 无限循环高频实时推送
	i := 0
	for {
		msg := fmt.Sprintf("实时UDP数据流 %d", i)
		_, _ = conn.Write([]byte(msg))
		i++
		time.Sleep(1 * time.Second) // 极快实时间隔
	}
}
