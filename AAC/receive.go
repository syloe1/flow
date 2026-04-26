package main

import (
	"encoding/binary"
	"net"
)

// RTP 头结构（和发送端一致）
type RTPHeader struct {
	Version        uint8
	Padding        uint8
	Extension      uint8
	CSRCCount      uint8
	Marker         uint8
	PayloadType    uint8
	SequenceNumber uint16
	Timestamp      uint32
	SSRC           uint32
}

func main() {
	addr, _ := net.ResolveUDPAddr("udp", ":3333")
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	println("✅ RTP 接收端已启动，监听 :3333 端口...")
	println("等待接收数据包...\n")

	buf := make([]byte, 1500)

	n, remoteAddr, err := conn.ReadFromUDP(buf)
	if err != nil {
		panic(err)
	}
	println("📦 收到 RTP 包，总长度：", n)
	println("来自：", remoteAddr)

	packet := buf[:n]
	header := RTPHeader{}

	// 第 1 字节
	header.Version = (packet[0] >> 6) & 0x03
	header.Padding = (packet[0] >> 5) & 0x01
	header.Extension = (packet[0] >> 4) & 0x01
	header.CSRCCount = packet[0] & 0x0F

	// 第 2 字节
	header.Marker = (packet[1] >> 7) & 0x01
	header.PayloadType = packet[1] & 0x7F

	// 3-4 字节：序列号（大端序解析）
	header.SequenceNumber = binary.BigEndian.Uint16(packet[2:4])
	// 5-8 字节：时间戳
	header.Timestamp = binary.BigEndian.Uint32(packet[4:8])
	// 9-12 字节：SSRC
	header.SSRC = binary.BigEndian.Uint32(packet[8:12])

	// 载荷数据（12字节后）
	payload := packet[12:]
	println("\n==== RTP 头部解析完成 ====")
	println("版本号(V)：", header.Version)
	println("填充(P)：", header.Padding)
	println("扩展(X)：", header.Extension)
	println("CSRC计数(CC)：", header.CSRCCount)
	println("标记(M)：", header.Marker)
	println("载荷类型(PT)：", header.PayloadType, " → 96=H264")
	println("序列号：", header.SequenceNumber)
	println("时间戳：", header.Timestamp)
	println("SSRC：", header.SSRC)
	println("载荷长度：", len(payload))
	println("载荷内容：", string(payload))
}
