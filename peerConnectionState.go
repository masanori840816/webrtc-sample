package main

import (
	"log"

	"github.com/pion/webrtc/v3"
)

type ConnectionState struct {
	userName       string
	websocket      *threadSafeWriter
	peerConnection *webrtc.PeerConnection
}

func NewPeerConnection() (*webrtc.PeerConnection, error) {
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{
					"turn:goapp.sample.jp:3478",
				},
				Username:       "username1",
				Credential:     "password1",
				CredentialType: webrtc.ICECredentialTypePassword,
			},
		},
		ICETransportPolicy: webrtc.ICETransportPolicyRelay,
	})
	if err != nil {
		return nil, err
	}
	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := peerConnection.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			log.Print(err)
			return nil, err
		}
	}
	return peerConnection, nil
}
