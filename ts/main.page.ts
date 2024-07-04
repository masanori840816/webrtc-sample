import { WebSockets } from "./websockets";

let websockets: WebSockets|null = null;
window.Page = {
    connect() {
        websockets = new WebSockets();
        websockets.addEvents((data) => console.log(data));
        websockets.connect();
        
    },
    send() {
        console.log("send " + (websockets != null));
        
        if(websockets != null) {
            const messageElm = document.getElementById("message") as HTMLInputElement;
            websockets.sendMessage({messageType: "text", data: messageElm.value ?? "empty"});
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
