package main

import (
	"github.com/pion/webrtc/v3"
)

func NewPeerConnection() (*webrtc.PeerConnection, error) {
	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{
					"stun:stun.l.google.com:19302",
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return peerConnection, nil
}
