export type WebsocketMessage = {
    messageType: "text"
    data: string|Blob|ArrayBuffer,
}
export type SdpMessage = {
    type: "video-offer" | "video-answer",
    sdp: RTCSessionDescription|null
}
export type CandidateMessage = {
    type: "new-ice-candidate",
    candidate: RTCIceCandidateInit|null
}