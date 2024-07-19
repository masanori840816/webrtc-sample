import { WebSockets } from "./websockets";
import * as webSocketMessage from "./webrtcsample.type";
import { WebRtcController } from "./webrtc.controller";
import { MainView } from "./main.view";

const view = new MainView();
let websockets: WebSockets|null = null;
let webrtc: WebRtcController|null = null;
window.Page = {
    connect() {
        websockets = new WebSockets();
        webrtc = new WebRtcController();
        websockets.addEvents((data) => {
            if(webSocketMessage.isWebsoMessage(data)) {
                console.log("websockmessage");
                
                console.log(data);
            } else if(webSocketMessage.isSdpMessage(data)) {
                if(data.type === "video-offer") {
                    webrtc?.handleVideoOffer(data.sdp);
                }else if(data.type === "video-answer") {
                    webrtc?.handleAnswer(data.sdp);
                }
            } else if(webSocketMessage.isCandidateMessage(data)) {
                webrtc?.handleCandidate(data.candidate);
            }
        });
        view.setLocalVideo();
        webrtc.init();
        webrtc.addEvents((data) => websockets?.sendMessage(data),
        (data) => websockets?.sendMessage(data),
        (stream) => view.setRemoteVideo(stream),
        (data) => console.log(data));
        websockets.connect();
    },
    sendOffer() {
        webrtc?.connect();
    },
    send() {
        console.log("send " + (websockets != null));
        
        if(websockets != null) {
            const messageElm = document.getElementById("message") as HTMLInputElement;
            websockets.sendMessage({type: "text", data: messageElm.value ?? "empty"});
        }
    },
    close() {
        if(websockets != null) {
            websockets.close();
        }

    },
    sendDataChannel() {

    }
};
