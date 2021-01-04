package main

import (
	"context"
	"encoding/json"

	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"time"

	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/ivfreader"
)

type udpConn struct {
	conn *net.UDPConn
	port int
}

func saveToDisk(i media.Writer, track *webrtc.TrackRemote) {
	defer func() {
		if err := i.Close(); err != nil {
			panic(err)
		}
	}()

	for {
		rtpPacket, _, err := track.ReadRTP()
		if err != nil {
			panic(err)
		}
		if err := i.WriteRTP(rtpPacket); err != nil {
			panic(err)
		}
	}
}

func main() {
	fmt.Println("START")
	http.HandleFunc("/connectsender", func(res http.ResponseWriter, req *http.Request) {
		fmt.Println("CONNECT REQUEST")
		body, _ := ioutil.ReadAll(req.Body)

		ans := make(chan string)

		go func() {
			offer := webrtc.SessionDescription{}
			json.Unmarshal(body, &offer)

			// Create a MediaEngine object to configure the supported codec
			m := webrtc.MediaEngine{}

			// Setup the codecs you want to use.
			// We'll use a VP8 and Opus but you can also define your own
			if err := m.RegisterCodec(webrtc.RTPCodecParameters{
				RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: "video/VP8", ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
				PayloadType:        96,
			}, webrtc.RTPCodecTypeVideo); err != nil {
				panic(err)
			}

			// Create the API object with the MediaEngine
			api := webrtc.NewAPI(webrtc.WithMediaEngine(&m))

			// Prepare the configuration
			config := webrtc.Configuration{
				ICEServers: []webrtc.ICEServer{
					{
						URLs: []string{"stun:stun.l.google.com:19302"},
					},
				},
			}

			// Create a new RTCPeerConnection
			peerConnection, err := api.NewPeerConnection(config)
			if err != nil {
				panic(err)
			}

			iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())

			// Create Track that we send video back to browser on
			videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/vp8"}, "video", "pion")
			if err != nil {
				panic(err)
			}

			// Add this newly created track to the PeerConnection
			rtpSender, err := peerConnection.AddTrack(videoTrack)
			if err != nil {
				panic(err)
			}

			// Read incoming RTCP packets
			// Before these packets are retuned they are processed by interceptors. For things
			// like NACK this needs to be called.
			go func() {
				rtcpBuf := make([]byte, 1500)
				for {
					if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
						return
					}
				}
			}()

			go func() {
				// Wait for connection established
				<-iceConnectedCtx.Done()
				
				reader, header, readerErr := ivfreader.NewWith(os.Stdin)
				if readerErr != nil {
					panic(readerErr)
				}

				// Send our video file frame at a time. Pace our sending so we send it at the same speed it should be played back as.
				// This isn't required since the video is timestamped, but we will such much higher loss if we send all at once.
				sleepTime := time.Millisecond * time.Duration((float32(header.TimebaseNumerator)/float32(header.TimebaseDenominator))*1000)
				for {
					frame, _, readerErr := reader.ParseNextFrame()
					if readerErr == io.EOF {
						fmt.Printf("All video frames parsed and sent")
						os.Exit(0)
					}

					if readerErr != nil {
						panic(readerErr)
					}

					time.Sleep(sleepTime)
					if writeErr := videoTrack.WriteSample(media.Sample{Data: frame, Duration: time.Second}); writeErr != nil {
						panic(writeErr)
					}
				}
			}()

			// Set the handler for ICE connection state
			// This will notify you when the peer has connected/disconnected
			peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
				fmt.Printf("Connection State has changed %s \n", connectionState.String())

				if connectionState == webrtc.ICEConnectionStateConnected {
					iceConnectedCtxCancel()
					fmt.Println("Ctrl+C the remote client to stop the demo")
				} else if connectionState == webrtc.ICEConnectionStateFailed ||
					connectionState == webrtc.ICEConnectionStateDisconnected {
					fmt.Println("Done")
					os.Exit(0)
				}
			})

			// Set the remote SessionDescription
			if err = peerConnection.SetRemoteDescription(offer); err != nil {
				panic(err)
			}

			// Create answer
			answer, err := peerConnection.CreateAnswer(nil)
			if err != nil {
				panic(err)
			}

			// Create channel that is blocked until ICE Gathering is complete
			gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

			// Sets the LocalDescription, and starts our UDP listeners
			if err = peerConnection.SetLocalDescription(answer); err != nil {
				panic(err)
			}

			// Block until ICE Gathering is complete, disabling trickle ICE
			// we do this because we only can exchange one signaling message
			// in a production application you should exchange ICE Candidates via OnICECandidate
			<-gatherComplete

			b, err := json.Marshal(*peerConnection.LocalDescription())
			if err != nil {
				panic(err)
			}
			ans <- string(b)

			// Block forever
			select {}
		}()

		msg := <-ans
		fmt.Fprintf(res, msg)
	})

	fs := http.FileServer(http.Dir("./www"))
	http.Handle("/www/", http.StripPrefix("/www/", fs))

	err := http.ListenAndServe("0.0.0.0:8080", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("END")

}
