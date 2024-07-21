import { WebSockets } from "./websockets";
import * as webSocketMessage from "./webrtcsample.type";
import { WebRtcController } from "./webrtc.controller";
import { MainView } from "./main.view";

const view = new MainView();
let websockets: WebSockets|null = null;
let webrtc: WebRtcController|null = null;
window.Page = {
    init() {
        view.setLocalVideo();
        webrtc = new WebRtcController();
        webrtc.init();
    },
    connect() {
        websockets = new WebSockets();
        websockets.addEvents((data) => {
            if(webSocketMessage.isWebsoMessage(data)) {
                console.log("websockmessage");
                
                console.log(data);
            } else if(webSocketMessage.isSdpMessage(data)) {
                if(data.type === "video-offer") {
                    webrtc?.handleVideoOffer(JSON.parse(data.data));
                }else if(data.type === "video-answer") {
                    webrtc?.handleAnswer(JSON.parse(data.data));
                }
            } else if(webSocketMessage.isCandidateMessage(data)) {
                webrtc?.handleCandidate(JSON.parse(data.data));
            }
        });
        webrtc?.addEvents((data) => websockets?.sendMessage(data),
        (data) => websockets?.sendMessage(data),
        (stream) => view.setRemoteVideo(stream),
        (data) => console.log(data));
        websockets.connect();
    },
    sendOffer() {
        webrtc?.connect();
    },
    send() {        
        if(websockets != null) {
            const messageElm = document.getElementById("message") as HTMLInputElement;
            websockets.sendMessage({type: "text", data: messageElm.value ?? "empty"});
        }
    },
    close() {
        webrtc?.close();
        websockets?.close();
    },
    sendDataChannel() {

    }
};
