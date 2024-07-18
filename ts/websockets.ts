import { CandidateMessage, SdpMessage, WebsocketMessage } from "./webrtcsample.type";
import * as domains from "./appDomains";

export class WebSockets {
    private ws: WebSocket|null = null;
    private onReceived: ((data: WebsocketMessage|SdpMessage|CandidateMessage) => void) | null = null;
    public connect() {
        const wsUrl = domains.getWebSocketAddress();
        const userElm = document.getElementById("user_name") as HTMLInputElement;
        this.ws = new WebSocket(`${wsUrl}?user=${userElm.value}`);
        this.ws.onopen = () => this.sendMessage({
            type: "text",
            data: "connected",
        });
        this.ws.onmessage = data => {
            const message = JSON.parse(data.data);
            if(message == null) {
                console.warn("Failed receiving a message");
                console.log(data);                
                return;
            }
            if(this.onReceived != null) {
                this.onReceived(message);
            }
        };
        
    }
    public addEvents(onReceived: (data: WebsocketMessage|SdpMessage|CandidateMessage) => void) {
        this.onReceived = onReceived;
    }
    public sendMessage(message: WebsocketMessage|SdpMessage|CandidateMessage) {
        console.log("sendmessage " + message);
        
        if (this.ws == null) {
            return;
        }
        this.ws.send(JSON.stringify(message));
    }
    public close() {
        if(this.ws == null) {
            return;
        }
        this.ws.close();
        this.ws = null;
    }
}


