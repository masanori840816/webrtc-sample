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

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	listLock    sync.RWMutex
	connections []ConnectionState
	trackLocals map[string]*webrtc.TrackLocalStaticRTP
)

type threadSafeWriter struct {
	*websocket.Conn
	sync.Mutex
}
type websocketMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Websocket start")
	user, err := getParam(r, "user")
	if err != nil {
		log.Println(err)
		return
	}
	unsafeConn, err := upgrader.Upgrade(w, r, nil)
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

	connectionState := ConnectionState{userName: user, websocket: conn, peerConnection: peerConnection}
	defer connectionState.Close()
	listLock.Lock()
	connections = append(connections, connectionState)
	listLock.Unlock()
	signalPeerConnections()
	message := &websocketMessage{}
	for {
		messageType, raw, err := conn.ReadMessage()
		if err != nil {
			log.Println("read err")
			log.Println(err)
			return
		} else if err := json.Unmarshal(raw, &message); err != nil {
			log.Println("Faile unmarcha")
			log.Println(err)
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
			for _, c := range connections {
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
func signalPeerConnections() {
	listLock.Lock()
	defer func() {
		listLock.Unlock()
		dispatchKeyFrame()
	}()

	attemptSync := func() (tryAgain bool) {
		for i := range connections {
			if connections[i].peerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
				connections = append(connections[:i], connections[i+1:]...)
				return true // We modified the slice, start from the beginning
			}
			existingSenders := map[string]bool{}

			for _, sender := range connections[i].peerConnection.GetSenders() {
				if sender.Track() == nil {
					continue
				}

				existingSenders[sender.Track().ID()] = true

				// If we have a RTPSender that doesn't map to a existing track remove and signal
				if _, ok := trackLocals[sender.Track().ID()]; !ok {
					if err := connections[i].peerConnection.RemoveTrack(sender); err != nil {
						return true
					}
				}
			}

			// Don't receive videos we are sending, make sure we don't have loopback
			for _, receiver := range connections[i].peerConnection.GetReceivers() {
				if receiver.Track() == nil {
					continue
				}

				existingSenders[receiver.Track().ID()] = true
			}

			// Add all track we aren't sending yet to the PeerConnection
			for trackID := range trackLocals {
				if _, ok := existingSenders[trackID]; !ok {
					if _, err := connections[i].peerConnection.AddTrack(trackLocals[trackID]); err != nil {
						return true
					}
				}
			}

			offer, err := connections[i].peerConnection.CreateOffer(nil)
			if err != nil {
				return true
			}

			if err = connections[i].peerConnection.SetLocalDescription(offer); err != nil {
				return true
			}

			offerString, err := json.Marshal(offer)
			if err != nil {
				return true
			}

			if err = connections[i].websocket.WriteJSON(&websocketMessage{
				Type: "video-offer",
				Data: string(offerString),
			}); err != nil {
				return true
			}
		}

		return
	}

	for syncAttempt := 0; ; syncAttempt++ {
		if syncAttempt == 25 {
			// Release the lock and attempt a sync in 3 seconds. We might be blocking a RemoveTrack or AddTrack
			go func() {
				time.Sleep(time.Second * 3)
				signalPeerConnections()
			}()
			return
		}

		if !attemptSync() {
			break
		}
	}
}
func dispatchKeyFrame() {
	listLock.Lock()
	defer listLock.Unlock()

	for i := range connections {
		for _, receiver := range connections[i].peerConnection.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}

			_ = connections[i].peerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(receiver.Track().SSRC()),
				},
			})
		}
	}
}
func addTrack(t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	listLock.Lock()
	defer func() {
		listLock.Unlock()
		signalPeerConnections()
	}()

	// Create a new TrackLocal with the same codec as our incoming
	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil {
		panic(err)
	}

	trackLocals[t.ID()] = trackLocal
	return trackLocal
}

// Remove from list of tracks and fire renegotation for all PeerConnections
func removeTrack(t *webrtc.TrackLocalStaticRTP) {
	listLock.Lock()
	defer func() {
		listLock.Unlock()
		signalPeerConnections()
	}()

	delete(trackLocals, t.ID())
}
