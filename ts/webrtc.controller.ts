import * as dataChannel from "./dataChannels"
import { CandidateMessage, SdpMessage } from "./webrtcsample.type";

export class WebRtcController {
    private webcamStream: MediaStream | null = null;
    private peerConnection: RTCPeerConnection | null = null;
    private dataChannels: dataChannel.DataChannel[] = [];

    private sdpMessageEvent: ((data: SdpMessage) => void) | null = null;
    private candidateMessageEvent: ((data: CandidateMessage) => void) | null = null;
    private remoteTrackEvent: ((stream: MediaStream) => void)|null =null;
    private dataChannelMessageEvent: ((data: string | Uint8Array) => void) | null = null;

    public init() {
        navigator.mediaDevices.getUserMedia({ video: true, audio: true })
            .then(stream => this.webcamStream = stream)
            .catch(err => console.error(err));
    }
    public addEvents(sdpMessage: ((data: SdpMessage) => void),
        candidateMessage: ((data: CandidateMessage) => void),
        remoteTrackEvent: ((stream: MediaStream) => void),
        dataChannelMessage: ((data: string | Uint8Array) => void)) {
        this.sdpMessageEvent = sdpMessage;
        this.candidateMessageEvent = candidateMessage;
        this.remoteTrackEvent = remoteTrackEvent;
        this.dataChannelMessageEvent = dataChannelMessage;
    }
    public connect() {
        if (this.webcamStream == null) {
            console.error("Local video was null");
            return;
        }
        if (this.peerConnection == null) {
            this.createPeerConnection();
            if (this.peerConnection == null) {
                console.error("Failed getting peerconnection");
                return;
            }
        }

        for (const t of this.webcamStream.getTracks()) {
            this.peerConnection.addTrack(t);
        }

    }
    public async handleVideoOffer(sdp: RTCSessionDescription) {        
        if (this.webcamStream == null) {
            console.error("No webcam source");
            return;
        }
        if (this.peerConnection == null) {
            this.createPeerConnection();
            if (this.peerConnection == null) {
                console.error("Failed getting peerconnection");
                return;
            }
        }
        await this.peerConnection.setRemoteDescription(sdp);        
        for (const t of this.webcamStream.getTracks()) {
            this.peerConnection.addTrack(t);
        }

        const answer = await this.peerConnection.createAnswer();
        if (this.peerConnection == null) {
            return;
        }
        await this.peerConnection.setLocalDescription(answer);        
        if (this.sdpMessageEvent != null &&
            this.peerConnection.localDescription != null) {
            this.sdpMessageEvent({
                type: "video-answer",
                sdp: this.peerConnection.localDescription
            });
        }
    }
    public async handleAnswer(sdp: RTCSessionDescription) {
        if (this.peerConnection == null) {
            console.error("PeerConnection was null");
            return;
        }        
        await this.peerConnection.setRemoteDescription(sdp);
    }
    public async handleCandidate(data: RTCIceCandidateInit | null | undefined) {
        if (this.peerConnection == null ||
            data == null) {
            console.error("PeerConnection|Candidate was null");
            return;
        }        
        await this.peerConnection.addIceCandidate(data);
    }
    private createPeerConnection() {
        this.peerConnection = new RTCPeerConnection({
            iceServers: [{
                urls: `stun:stun.l.google.com:19302`,
            }]
        });

        this.peerConnection.oniceconnectionstatechange = (ev) => console.log(ev);
        this.peerConnection.onicegatheringstatechange = (ev) => console.log(ev);
        this.peerConnection.onsignalingstatechange = (ev) => console.log(ev);
        this.peerConnection.onnegotiationneeded = async (ev) => await this.handleNegotiationNeededEvent(ev);
        this.peerConnection.onconnectionstatechange = () => {
            console.log(this.peerConnection?.connectionState);
        };
        this.peerConnection.ontrack = (ev) => this.handleRemoteTrackEvent(ev);
        this.peerConnection.onicecandidate = ev => {
            if (ev.candidate == null ||
                this.candidateMessageEvent == null) {
                return;
            }
            this.candidateMessageEvent({ type: "new-ice-candidate", candidate: ev.candidate });
        };
        /*this.dataChannels.push(
            dataChannel.createTextDataChannel("sample1", 20, this.peerConnection,
                (message) => {
                    if (this.dataChannelMessageEvent != null) {
                        this.dataChannelMessageEvent(message);
                    }
                }));*/
    }
    private async handleNegotiationNeededEvent(ev: Event) {
        if (this.peerConnection == null) {
            return;
        }
        try {
            const offer = await this.peerConnection.createOffer();
            if (this.peerConnection.signalingState !== "stable") {
                console.log("-- The connection isn't stable yet; postponing...")
                return;
            }
            await this.peerConnection.setLocalDescription(offer);
            if (this.sdpMessageEvent == null) {
                console.warn("No Offer message handlers");
                return;
            }
            console.log("---> Sending the offer to the remote peer");
            if(this.peerConnection.localDescription == null) {
                console.error("Local description was null");
                return;
            }
            this.sdpMessageEvent({
                type: "video-offer",
                sdp: this.peerConnection.localDescription
            });
        } catch (err) {
            console.error(err);

        }
    }
    private handleRemoteTrackEvent(ev: RTCTrackEvent) {        
        if(this.remoteTrackEvent == null) {
            return;
        }
        if(ev.streams[0] == null) {
            const tracks = this.peerConnection?.getReceivers()?.map(r => r.track);
            if(tracks != null) {
                this.remoteTrackEvent(new MediaStream(tracks));
            }
        } else if(ev.track.kind) {
            this.remoteTrackEvent(ev.streams[0]);
        }        
    }
}