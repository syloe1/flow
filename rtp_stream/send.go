package main

//target
//发送 RTP 流
//序列号自动 + 1
//时间戳自动递增

/*
序列号 SequenceNumber
	每发送1 个 RTP 包 +1
时间戳 Timestamp
	视频固定 90000 赫兹（Hz）
	每 1 毫秒 + 90
	每 20 毫秒 + 1800（最常用的视频帧间隔）
SSRC 全程不变
Version 永远 = 2

*/
import (
	"encoding/binary"
	"net"
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
	addr := "127.0.0.1:3333"
	conn, err := net.Dial("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	header := RTPHeader{
		Version:        2,          // 固定
		PayloadType:    96,         // H264 视频
		SequenceNumber: 1000,       // 初始序列号（随便写）
		Timestamp:      0,          // 初始时间戳
		SSRC:           0x12345678, // 固定不变
	}

	payload := []byte("this is h264 video frame")

	for {
		packet := make([]byte, 12+len(payload))

		packet[0] = (header.Version << 6) | (header.Padding << 5) | (header.Extension << 4) | header.CSRCCount
		packet[1] = (header.Marker << 7) | header.PayloadType
		//SN
		binary.BigEndian.PutUint16(packet[2:4], header.SequenceNumber)
		//TImeStamp
		binary.BigEndian.PutUint32(packet[4:8], header.Timestamp)
		//SSRC
		binary.BigEndian.PutUint32(packet[8:12], header.SSRC)

		copy(packet[12:], payload)

		_, err = conn.Write(packet)
		if err != nil {
			panic(err)
		}

		header.SequenceNumber++
		header.Timestamp += 1800

		println("发送成功 | 序号:", header.SequenceNumber-1, " | 时间戳:", header.Timestamp-1800)
		time.Sleep(20 * time.Millisecond)
	}
}
