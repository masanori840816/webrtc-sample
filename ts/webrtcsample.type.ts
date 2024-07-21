export type WebSocketMessage = {
    type: "text"
    data: string|Blob|ArrayBuffer,
}
export type SdpMessage = {
    type: "video-offer" | "video-answer",
    data: string
}
export type CandidateMessage = {
    type: "new-ice-candidate",
    data: string
}
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
    if(("type" in value && "data" in value) === false) {
        return false;
    }
    return (value["type"] === "video-offer" || value["type"] === "video-answer");
}
export function isCandidateMessage(value: any): value is CandidateMessage {
    if(value == null) {
        return false;
    }
    if(("type" in value && "data" in value) === false) {
        return false;
    }
    return (value["type"] === "new-ice-candidate");
}