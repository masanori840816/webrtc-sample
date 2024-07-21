import { hasAnyTexts } from "./hasAnyTexts";

export function getWebSocketAddress() {
    let port = location.port;
    if(hasAnyTexts(port)) {
        port = ":" + port;
    }

    if(location.protocol === "https:") {
        return `wss://${location.hostname}${port}/webrtc/websocket`;
    }
    return "ws://localhost:8083/websocket";
    //return `ws://${location.hostname}${port}/webrtc/websocket`;
}
export function getBaseAddress() {
    let port = location.port;
    if(hasAnyTexts(port)) {
        port = ":" + port;
    }
    return `${location.protocol}//${location.hostname}${port}/webrtc`;
}