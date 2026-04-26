package main

import (
	"bytes"
	"encoding/binary"
	"net"
	"os"
	"time"
)

// RTP 头
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
	file, err := os.Open("test.h264")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	header := RTPHeader{
		Version:        2,
		PayloadType:    96, // H264
		SequenceNumber: 1000,
		Timestamp:      0,
		SSRC:           0x12345678,
	}

	buf := make([]byte, 1024*1024)
	n, _ := file.Read(buf)
	h264Data := buf[:n]

	//按照00 00 00 01
	nalus := bytes.Split(h264Data, []byte{0x00, 0x00, 0x01, 0x01})
	for _, nalu := range nalus {
		if len(nalu) < 2 {
			continue
		}
		rtpBuf := make([]byte, 12+len(nalu))

		// 组装RTP头
		rtpBuf[0] = (header.Version << 6) | (header.Padding << 5) | (header.Extension << 4) | header.CSRCCount
		rtpBuf[1] = (header.Marker << 7) | header.PayloadType

		binary.BigEndian.PutUint16(rtpBuf[2:4], header.SequenceNumber)
		binary.BigEndian.PutUint32(rtpBuf[4:8], header.Timestamp)
		binary.BigEndian.PutUint32(rtpBuf[8:12], header.SSRC)

		copy(rtpBuf[12:], nalu)

		_, err := conn.Write(rtpBuf)
		if err != nil {
			panic(err)
		}
		println("发送h264 RTP包 | 序号", header.SequenceNumber)

		header.SequenceNumber++
		header.Timestamp += 1800 //20ms

		time.Sleep(20 * time.Millisecond)
	}
}
