import { CandidateMessage, SdpMessage, WebSocketMessage } from "./webrtcsample.type"

export function isWebsoMessage(value: any): value is WebSocketMessage {
    if(value == null) {
        return false;
    }
    if(("type" in value && "data" in value) === false) {
        return false;
    }
    return (value["type"] === "text");
}
export function isSdpMessage(value: any): value is SdpMessage {
    if(value == null) {
        return false;
    }
    if(("type" in value && "sdp" in value) === false) {
        return false;
    }
    return (value["type"] === "video-offer" || value["type"] === "video-answer");
}
export function isCandidateMessage(value: any): value is CandidateMessage {
    if(value == null) {
        return false;
    }
    if(("type" in value && "candidate" in value) === false) {
        return false;
    }
    return (value["type"] === "new-ice-candidate");
}