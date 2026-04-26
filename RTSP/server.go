package main

import (
	"log"
	"time"

	"github.com/bluenviron/gortsplib/v5"
	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/bluenviron/gortsplib/v5/pkg/description"
	"github.com/bluenviron/gortsplib/v5/pkg/format"
	"github.com/pion/rtp"
)

type serverHandler struct {
	stream *gortsplib.ServerStream
	media  *description.Media
}

func (h *serverHandler) OnDescribe(ctx *gortsplib.ServerHandlerOnDescribeCtx) (*base.Response, *gortsplib.ServerStream, error) {
	if ctx.Path != "/test" && ctx.Path != "test" {
		return &base.Response{StatusCode: base.StatusNotFound}, nil, nil
	}
	return &base.Response{StatusCode: base.StatusOK}, h.stream, nil
}

func (h *serverHandler) OnSetup(ctx *gortsplib.ServerHandlerOnSetupCtx) (*base.Response, *gortsplib.ServerStream, error) {
	if ctx.Path != "/test" && ctx.Path != "test" {
		return &base.Response{StatusCode: base.StatusNotFound}, nil, nil
	}
	return &base.Response{StatusCode: base.StatusOK}, h.stream, nil
}

func (h *serverHandler) OnPlay(*gortsplib.ServerHandlerOnPlayCtx) (*base.Response, error) {
	return &base.Response{StatusCode: base.StatusOK}, nil
}

func main() {
	media := &description.Media{
		Type: description.MediaTypeVideo,
		Formats: []format.Format{
			&format.H264{
				PayloadTyp: 96,
				SPS: []byte{
					0x67, 0x42, 0x00, 0x1e, 0x8d, 0x68, 0x50, 0x1e,
					0xd0, 0x0f, 0x12, 0x26, 0xa0,
				},
				PPS: []byte{0x68, 0xce, 0x38, 0x80},
				PacketizationMode: 1,
			},
		},
	}

	server := &gortsplib.Server{
		Handler:     &serverHandler{media: media},
		RTSPAddress: ":18556",
	}

	stream := &gortsplib.ServerStream{
		Server: server,
		Desc: &description.Session{
			Medias: []*description.Media{media},
		},
	}
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	if err := stream.Initialize(); err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	handler := server.Handler.(*serverHandler)
	handler.stream = stream

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		var seq uint16
		var ts uint32

		for range ticker.C {
			pkt := &rtp.Packet{
				Header: rtp.Header{
					Version:        2,
					PayloadType:    96,
					SequenceNumber: seq,
					Timestamp:      ts,
					SSRC:           0x12345678,
					Marker:         true,
				},
				Payload: []byte{0x05, 0xff, 0xff, 0xff},
			}

			err := stream.WritePacketRTP(media, pkt)
			if err != nil {
				log.Printf("send RTP failed: %v", err)
			} else {
				log.Printf("sent RTP: seq=%d ts=%d", seq, ts)
			}

			seq++
			ts += 45000
		}
	}()

	log.Println("RTSP server started: rtsp://127.0.0.1:18556/test (TCP transport)")
	log.Fatal(server.Wait())
}
