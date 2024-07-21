package main

import (
	"encoding/json"
	"log"

	"github.com/pion/webrtc/v3"
)

type ConnectionState struct {
	userName       string
	websocket      *threadSafeWriter
	peerConnection *webrtc.PeerConnection
}

func (conn *ConnectionState) Close() {
	if conn.peerConnection.ConnectionState() != webrtc.PeerConnectionStateClosed {
		conn.peerConnection.Close()
	}
}
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
	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := peerConnection.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			return nil, err
		}
	}
	return peerConnection, nil
}
func NewConnectionState(name string, websocket *threadSafeWriter, peerConnection *webrtc.PeerConnection) *ConnectionState {
	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}
		j, _ := json.Marshal(i)
		message := websocketMessage{Type: "new-ice-candidate", Data: string(j)}
		websocket.Lock()
		defer websocket.Unlock()
		err := websocket.WriteJSON(message)
		if err != nil {
			return
		}
	})
	peerConnection.OnConnectionStateChange(func(p webrtc.PeerConnectionState) {
		log.Printf("State: %s", p.String())
		switch p {
		case webrtc.PeerConnectionStateConnected:
			for _, rcv := range peerConnection.GetReceivers() {
				track := rcv.Track()
				if track == nil {
					continue
				}
				log.Printf("RECV ID: %s MID: %s MSID: %s Kind: %s", track.ID(), track.RID(), track.Msid(), track.Kind())
			}
		case webrtc.PeerConnectionStateFailed:
			return
		case webrtc.PeerConnectionStateClosed:
			return
		}
	})
	peerConnection.OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		log.Println("Ontrack")
		trackLocal := addTrack(t)
		defer removeTrack(trackLocal)

		buf := make([]byte, 1500)
		for {
			i, _, err := t.Read(buf)
			if err != nil {
				return
			}

			if _, err = trackLocal.Write(buf[:i]); err != nil {
				return
			}
		}
	})

	return &ConnectionState{userName: name, websocket: websocket, peerConnection: peerConnection}
}
