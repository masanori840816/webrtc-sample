import * as dataChannel from "./dataChannels"
import { CandidateMessage, SdpMessage } from "./webrtcsample.type";

export class WebRtcController {
    private webcamStream: MediaStream | null = null;
    private peerConnection: RTCPeerConnection | null = null;
    private dataChannels: dataChannel.DataChannel[] = [];

    private sdpMessageEvent: ((data: SdpMessage) => void) | null = null;
    private candidateMessageEvent: ((data: CandidateMessage) => void) | null = null;
    private dataChannelMessageEvent: ((data: string | Uint8Array) => void) | null = null;

    public init() {
        navigator.mediaDevices.getUserMedia({ video: true, audio: true })
            .then(stream => this.webcamStream = stream)
            .catch(err => console.error(err));
    }
    public addEvents(sdpMessage: ((data: SdpMessage) => void),
            candidateMessage: ((data: CandidateMessage) => void),
            dataChannelMessage: ((data: string | Uint8Array) => void)) {
        this.sdpMessageEvent = sdpMessage;
        this.candidateMessageEvent = candidateMessage;
        this.dataChannelMessageEvent = dataChannelMessage;
    }
    public connect() {
        if (this.webcamStream == null) {
            console.error("Local video was null");
            return;
        }
        this.peerConnection = new RTCPeerConnection({
            iceServers: [{
                urls: `stun:stun.l.google.com:19302`,  // A STUN server              
            }]
        });

        this.peerConnection.onconnectionstatechange = () => {
            /*if(this.peerConnection?.connectionState === "connected") {
                this.localAudioContext.resume();
            } else {
                this.localAudioContext.suspend();
            }*/
           console.log(this.peerConnection?.connectionState);
           
        };
        this.peerConnection.ontrack = (ev) => {};
        for(const t of this.webcamStream.getTracks()) {
            this.peerConnection.addTrack(t);
        }
        this.peerConnection.onicecandidate = ev => {
            if (ev.candidate == null ||
                this.candidateMessageEvent == null) {
                return;
            }
            this.candidateMessageEvent({ type: "new-ice-candidate", candidate: ev.candidate});
        };
        this.dataChannels.push(
            dataChannel.createTextDataChannel("sample3", 20, this.peerConnection,
                (message) => {
                    if (this.dataChannelMessageEvent != null) {
                        this.dataChannelMessageEvent(message);
                    }
                }));
    }
}