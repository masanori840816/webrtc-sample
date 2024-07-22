package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
)

type WebRTCConnection struct {
	upgrader    websocket.Upgrader
	listLock    sync.RWMutex
	connections []ConnectionState
	trackLocals map[string]*webrtc.TrackLocalStaticRTP
}

type threadSafeWriter struct {
	*websocket.Conn
	sync.Mutex
}
type websocketMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func NewWebRTCConnection() *WebRTCConnection {
	return &WebRTCConnection{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		listLock:    sync.RWMutex{},
		connections: make([]ConnectionState, 0),
		trackLocals: map[string]*webrtc.TrackLocalStaticRTP{},
	}
}
func (webrtcConn *WebRTCConnection) websocketHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getParam(r, "user")
	if err != nil {
		log.Println(err)
		return
	}
	unsafeConn, err := webrtcConn.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	conn := &threadSafeWriter{unsafeConn, sync.Mutex{}}
	// Close the connection when the for-loop operation is finished.
	defer conn.Close()
	peerConnection, err := NewPeerConnection()
	if err != nil {
		log.Println(err)
		return
	}
	defer peerConnection.Close()
	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}
		j, _ := json.Marshal(i)
		message := websocketMessage{Type: "new-ice-candidate", Data: string(j)}
		conn.Lock()
		defer conn.Unlock()
		err := conn.WriteJSON(message)
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
		trackLocal := webrtcConn.addTrack(t)
		defer webrtcConn.removeTrack(trackLocal)

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

	webrtcConn.listLock.Lock()
	webrtcConn.connections = append(webrtcConn.connections, ConnectionState{userName: user, websocket: conn, peerConnection: peerConnection})
	webrtcConn.listLock.Unlock()

	webrtcConn.signalPeerConnections()
	message := &websocketMessage{}
	for {
		messageType, raw, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Failed reading ReadMessage Err: %s", err.Error())
			return
		} else if err := json.Unmarshal(raw, &message); err != nil {
			log.Println(err.Error())
			return
		}

		switch message.Type {
		case "video-answer":
			answer := webrtc.SessionDescription{}
			if err := json.Unmarshal([]byte(message.Data), &answer); err != nil {
				log.Println(err.Error())
				return
			}
			if err := peerConnection.SetRemoteDescription(answer); err != nil {
				log.Println(err.Error())
				return
			}
		case "new-ice-candidate":

			log.Println("candidate")
			log.Println(message.Data)
			candidate := webrtc.ICECandidateInit{}
			if err := json.Unmarshal([]byte(message.Data), &candidate); err != nil {
				log.Println(err)
				return
			}

			if err := peerConnection.AddICECandidate(candidate); err != nil {
				log.Println(err)
				return
			}
		default:
			log.Println("default")
			for _, c := range webrtcConn.connections {
				if c.userName == user {
					continue
				}
				c.websocket.WriteMessage(messageType, raw)
			}
		}
	}
}
func getParam(r *http.Request, key string) (string, error) {
	result := r.URL.Query().Get(key)
	if len(result) <= 0 {
		return "", fmt.Errorf("no value: %s", key)
	}
	return result, nil
}
func (webrtcConn *WebRTCConnection) signalPeerConnections() {
	webrtcConn.listLock.Lock()
	defer func() {
		webrtcConn.listLock.Unlock()
		webrtcConn.dispatchKeyFrame()
	}()

	attemptSync := func() (tryAgain bool) {
		for i := range webrtcConn.connections {
			if webrtcConn.connections[i].peerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
				webrtcConn.connections = append(webrtcConn.connections[:i], webrtcConn.connections[i+1:]...)
				return true
			}
			existingSenders := map[string]bool{}

			for _, sender := range webrtcConn.connections[i].peerConnection.GetSenders() {
				if sender.Track() == nil {
					continue
				}

				existingSenders[sender.Track().ID()] = true

				// If we have a RTPSender that doesn't map to a existing track remove and signal
				if _, ok := webrtcConn.trackLocals[sender.Track().ID()]; !ok {
					if err := webrtcConn.connections[i].peerConnection.RemoveTrack(sender); err != nil {
						return true
					}
				}
			}

			// Don't receive videos we are sending, make sure we don't have loopback
			for _, receiver := range webrtcConn.connections[i].peerConnection.GetReceivers() {
				if receiver.Track() == nil {
					continue
				}

				existingSenders[receiver.Track().ID()] = true
			}

			// Add all track we aren't sending yet to the PeerConnection
			for trackID := range webrtcConn.trackLocals {
				if _, ok := existingSenders[trackID]; !ok {
					if _, err := webrtcConn.connections[i].peerConnection.AddTrack(webrtcConn.trackLocals[trackID]); err != nil {
						return true
					}
				}
			}

			offer, err := webrtcConn.connections[i].peerConnection.CreateOffer(nil)
			if err != nil {
				return true
			}

			if err = webrtcConn.connections[i].peerConnection.SetLocalDescription(offer); err != nil {
				return true
			}

			offerString, err := json.Marshal(offer)
			if err != nil {
				return true
			}

			if err = webrtcConn.connections[i].websocket.WriteJSON(&websocketMessage{
				Type: "video-offer",
				Data: string(offerString),
			}); err != nil {
				return true
			}
		}

		return
	}
	log.Println("attempt")
	for syncAttempt := 0; ; syncAttempt++ {
		if syncAttempt == 25 {
			// Release the lock and attempt a sync in 3 seconds. We might be blocking a RemoveTrack or AddTrack
			go func() {
				time.Sleep(time.Second * 3)
				webrtcConn.signalPeerConnections()
			}()
			return
		}

		if !attemptSync() {
			break
		}
	}
}
func (webrtcConn *WebRTCConnection) dispatchKeyFrame() {
	webrtcConn.listLock.Lock()
	defer webrtcConn.listLock.Unlock()
	for i := range webrtcConn.connections {
		for _, receiver := range webrtcConn.connections[i].peerConnection.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}
			_ = webrtcConn.connections[i].peerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(receiver.Track().SSRC()),
				},
			})
		}
	}
}
func (webrtcConn *WebRTCConnection) addTrack(t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	webrtcConn.listLock.Lock()
	defer func() {
		webrtcConn.listLock.Unlock()
		webrtcConn.signalPeerConnections()
	}()
	log.Println(t.Kind())

	// Create a new TrackLocal with the same codec as our incoming
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil {
		log.Println("newtraklocal err")
		panic(err)
	}
	webrtcConn.trackLocals[t.ID()] = trackLocal
	log.Println("addtracl next")
	return trackLocal
}

// Remove from list of tracks and fire renegotation for all PeerConnections
func (webrtcConn *WebRTCConnection) removeTrack(t *webrtc.TrackLocalStaticRTP) {
	webrtcConn.listLock.Lock()
	defer func() {
		webrtcConn.listLock.Unlock()
		webrtcConn.signalPeerConnections()
	}()

	delete(webrtcConn.trackLocals, t.ID())
}
