package main

import (
	"encoding/binary"
	"net"
	"sort"
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

// RTP包（头+数据）
type RTPPacket struct {
	Header   RTPHeader
	Data     []byte
	RecvTime time.Time // 接收时间（用于超时）
}

// 重排管理器
type ReorderManager struct {
	expectSeq    uint16               // 期望的下一个seq
	lastSeq      uint16               // 上一个成功输出的seq
	packetBuffer map[uint16]RTPPacket // 乱序包缓存
	maxBuffer    int                  // 最大缓存数（防内存爆炸）
	timeout      time.Duration        // 超时时间
	totalPackets int                  // 总包数
	lostPackets  int                  // 丢包数
}

func NewReorderManager(initSeq uint16) *ReorderManager {
	return &ReorderManager{
		expectSeq:    initSeq + 1,
		lastSeq:      initSeq,
		packetBuffer: make(map[uint16]RTPPacket),
		maxBuffer:    50,                    // 最多缓存50个包
		timeout:      80 * time.Millisecond, // 超时80ms
	}
}

// 比较两个seq大小（处理回绕）
func seqGreater(a, b uint16) bool {
	return int32(a-b) > 0
}

// 插入包 + 尝试输出有序包
func (r *ReorderManager) Push(packet RTPPacket) []RTPPacket {
	seq := packet.Header.SequenceNumber
	r.totalPackets++

	// 包太老，直接丢弃
	if seqGreater(r.lastSeq, seq) {
		return nil
	}

	// 缓存满了，清空
	if len(r.packetBuffer) >= r.maxBuffer {
		r.flushBuffer()
	}

	// 存入缓存
	r.packetBuffer[seq] = packet

	// 尝试输出连续包
	var output []RTPPacket
	for {
		p, ok := r.packetBuffer[r.expectSeq]
		if !ok {
			break
		}

		// 超时判断
		if time.Since(p.RecvTime) > r.timeout {
			break
		}

		// 输出有序包
		output = append(output, p)
		delete(r.packetBuffer, r.expectSeq)

		// 统计丢包
		if seqGreater(r.expectSeq, r.lastSeq+1) {
			r.lostPackets += int(r.expectSeq - r.lastSeq - 1)
		}

		r.lastSeq = r.expectSeq
		r.expectSeq++
	}

	return output
}

// 清空缓存（强制输出）
func (r *ReorderManager) flushBuffer() {
	var seqs []uint16
	for seq := range r.packetBuffer {
		seqs = append(seqs, seq)
	}

	sort.Slice(seqs, func(i, j int) bool {
		return seqGreater(seqs[i], seqs[j])
	})

	for _, seq := range seqs {
		delete(r.packetBuffer, seq)
	}

	r.expectSeq = r.lastSeq + 1
}

// 打印统计
func (r *ReorderManager) PrintStats() {
	lostRate := float64(r.lostPackets) / float64(r.totalPackets) * 100
	println("=====================================")
	println("总包数：", r.totalPackets)
	println("丢包数：", r.lostPackets)
	println("丢包率：", lostRate, "%")
	println("缓存剩余：", len(r.packetBuffer))
	println("=====================================")
}

// 解包函数
func parseRTP(packet []byte) RTPHeader {
	header := RTPHeader{}
	header.Version = (packet[0] >> 6) & 0x03
	header.Padding = (packet[0] >> 5) & 0x01
	header.Extension = (packet[0] >> 4) & 0x01
	header.CSRCCount = packet[0] & 0x0F
	header.Marker = (packet[1] >> 7) & 0x01
	header.PayloadType = packet[1] & 0x7F
	header.SequenceNumber = binary.BigEndian.Uint16(packet[2:4])
	header.Timestamp = binary.BigEndian.Uint32(packet[4:8])
	header.SSRC = binary.BigEndian.Uint32(packet[8:12])
	return header
}

func main() {
	// 监听端口
	addr, _ := net.ResolveUDPAddr("udp", ":3333")
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	println("✅ RTP 接收端启动（带重排+丢包统计），监听 :3333")

	buf := make([]byte, 1500)
	var reorder *ReorderManager
	first := true

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		// 解析RTP
		packetData := make([]byte, n)
		copy(packetData, buf[:n])
		header := parseRTP(packetData)

		// 初始化重排管理器
		if first {
			reorder = NewReorderManager(header.SequenceNumber)
			first = false
		}

		// 构造包
		pkt := RTPPacket{
			Header:   header,
			Data:     packetData[12:],
			RecvTime: time.Now(),
		}

		// 推入重排
		outputPkts := reorder.Push(pkt)

		// 输出有序包
		for _, op := range outputPkts {
			println("📦 有序输出 | seq:", op.Header.SequenceNumber, "ts:", op.Header.Timestamp)
		}

		// 每秒打印统计
		if reorder.totalPackets%50 == 0 {
			reorder.PrintStats()
		}
	}
}
