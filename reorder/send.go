package main

import (
	"encoding/binary"
	"net"
	"time"
)

// RTP 固定头 12字节
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
	// 1. 连接UDP 3333端口（和接收端一致）
	conn, err := net.Dial("udp", "127.0.0.1:3333")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 2. 初始化 RTP 头
	header := RTPHeader{
		Version:        2,
		PayloadType:    96,   // H264
		SequenceNumber: 1000, // 初始序号
		Timestamp:      0,
		SSRC:           0x12345678,
	}

	// 模拟视频数据
	payload := []byte("this is h264 video frame")

	// 3. 无限循环发送
	for {
		// --------------------------
		// 打包 RTP
		// --------------------------
		packet := make([]byte, 12+len(payload))

		// 第1字节
		packet[0] = (header.Version << 6) |
			(header.Padding << 5) |
			(header.Extension << 4) |
			header.CSRCCount

		// 第2字节
		packet[1] = (header.Marker << 7) | header.PayloadType

		// 序列号（大端）
		binary.BigEndian.PutUint16(packet[2:4], header.SequenceNumber)
		// 时间戳（大端）
		binary.BigEndian.PutUint32(packet[4:8], header.Timestamp)
		// SSRC
		binary.BigEndian.PutUint32(packet[8:12], header.SSRC)

		// 拷贝数据
		copy(packet[12:], payload)

		// --------------------------
		// 发送
		// --------------------------
		_, err = conn.Write(packet)
		if err != nil {
			panic(err)
		}

		// 打印
		println("发送 | 序号:", header.SequenceNumber, " 时间戳:", header.Timestamp)

		// --------------------------
		// 自增
		// --------------------------
		header.SequenceNumber++
		header.Timestamp += 1800

		// 20ms 发一帧
		time.Sleep(20 * time.Millisecond)
	}
}
