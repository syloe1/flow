package main

import (
	"encoding/binary"
	"net"
	"os"
	"time"
)

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
	// 1. 连接UDP
	conn, err := net.Dial("udp", "127.0.0.1:3333")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 2. 打开AAC文件
	file, err := os.Open("test.aac")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// 3. RTP 头初始化
	header := RTPHeader{
		Version:        2,
		PayloadType:    97, // ✅ AAC 固定 97
		SequenceNumber: 1000,
		Timestamp:      0,
		SSRC:           0x11223344,
	}

	// 每次读取一帧AAC
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if n == 0 || err != nil {
			break
		}

		// --------------------------
		// ✅ AAC 关键：去掉 ADTS 头（7字节）
		// --------------------------
		aacBody := buf[7:n]

		// --------------------------
		// 打包 RTP
		// --------------------------
		rtpPacket := make([]byte, 12+len(aacBody))

		// 头组装
		rtpPacket[0] = (header.Version << 6) | (header.Padding << 5) | (header.Extension << 4) | header.CSRCCount
		rtpPacket[1] = (header.Marker << 7) | header.PayloadType

		binary.BigEndian.PutUint16(rtpPacket[2:4], header.SequenceNumber)
		binary.BigEndian.PutUint32(rtpPacket[4:8], header.Timestamp)
		binary.BigEndian.PutUint32(rtpPacket[8:12], header.SSRC)

		// 拷贝AAC数据
		copy(rtpPacket[12:], aacBody)

		// --------------------------
		// 发送
		// --------------------------
		_, err = conn.Write(rtpPacket)
		if err != nil {
			panic(err)
		}

		// 打印
		println("发送AAC RTP包 | 序号:", header.SequenceNumber, " AAC长度:", len(aacBody))

		// --------------------------
		// 自增
		// --------------------------
		header.SequenceNumber++
		header.Timestamp += 160 // ✅ AAC 20ms += 160

		// 20ms 一帧
		time.Sleep(20 * time.Millisecond)
	}
}
