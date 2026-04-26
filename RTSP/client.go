package main

import (
	"log"
	"os"
	"time"

	"github.com/bluenviron/gortsplib/v5"
	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/bluenviron/gortsplib/v5/pkg/format"
	"github.com/pion/rtp"
)

func main() {
	rtspURL := "rtsp://127.0.0.1:18556/test"
	if len(os.Args) > 1 {
		rtspURL = os.Args[1]
	}

	u, err := base.ParseURL(rtspURL)
	if err != nil {
		log.Fatalf("parse RTSP URL failed: %v", err)
	}

	protocol := gortsplib.ProtocolTCP
	client := gortsplib.Client{
		Scheme:   u.Scheme,
		Host:     u.Host,
		Protocol: &protocol,
	}

	if err := client.Start(); err != nil {
		log.Fatalf("connect RTSP server failed: %v", err)
	}
	defer client.Close()

	desc, _, err := client.Describe(u)
	if err != nil {
		log.Fatalf("DESCRIBE failed: %v", err)
	}

	var h264Format *format.H264
	media := desc.FindFormat(&h264Format)
	if media == nil {
		log.Fatal("no H264 media found")
	}

	if _, err := client.Setup(desc.BaseURL, media, 0, 0); err != nil {
		log.Fatalf("SETUP failed: %v", err)
	}

	packetCount := 0
	client.OnPacketRTP(media, h264Format, func(pkt *rtp.Packet) {
		packetCount++
		log.Printf(
			"received RTP: seq=%d ts=%d marker=%v payload=%dB",
			pkt.SequenceNumber,
			pkt.Timestamp,
			pkt.Marker,
			len(pkt.Payload),
		)
	})

	if _, err := client.Play(nil); err != nil {
		log.Fatalf("PLAY failed: %v", err)
	}

	log.Printf("connected to %s, waiting for RTP packets", rtspURL)
	for {
		time.Sleep(3 * time.Second)
		log.Printf("received %d RTP packets so far", packetCount)
	}
}
