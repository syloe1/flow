package main

import (
	"encoding/binary"
	"net"
)

type RTPHeader struct {
	Version        uint8  //2bit 版本
	Padding        uint8  //1bit 填充
	Extension      uint8  //1bit 扩展字段
	CSRCCount      uint8  //4bit CSRC字段数量
	Marker         uint8  //1bit 标记位,视频 I 帧、音频帧结束
	PayloadType    uint8  //7bit 有效载荷类型 0 PCMU 8 PCMA 96  H264 97 AAC
	SequenceNumber uint16 //16bit 序列号 自增
	Timestamp      uint32 //32bit 时间戳
	SSRC           uint32 //32bit 源端 SSRC 同步源
}

func main() {
	conn, err := net.Dial("udp", "127.0.0.1:3333")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	header := RTPHeader{
		Version:        2,
		PayloadType:    96,   // H264 视频
		SequenceNumber: 1000, //序列号
		Timestamp:      12345678,
		SSRC:           0x12345678,
	}

	payload := []byte("this is h264 video data from GO")

	packet := make([]byte, 12+len(payload))

	//V P EX CC
	packet[0] = (header.Version << 6) | (header.Padding << 5) | (header.Extension << 4) | header.CSRCCount
	//M PT
	packet[1] = (header.Marker << 7) | header.PayloadType
	//SN
	binary.BigEndian.PutUint16(packet[2:4], header.SequenceNumber)
	//TImeStamp
	binary.BigEndian.PutUint32(packet[4:8], header.Timestamp)
	//SSRC
	binary.BigEndian.PutUint32(packet[8:12], header.SSRC)

	//拷贝载荷
	copy(packet[12:], payload)

	_, err = conn.Write(packet)
	if err != nil {
		panic(err)
	}
	println("✅ RTP 包发送成功！")
	println("目标:udp://127.0.0.1:3333")
	println("序列号:", header.SequenceNumber)
	println("时间戳:", header.Timestamp)
	println("SSRC:", header.SSRC)
	println("载荷长度:", len(payload))
}
