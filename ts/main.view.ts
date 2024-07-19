export class MainView {
    private localVideoElm: HTMLVideoElement;
    private remoteVideoElm: HTMLVideoElement;
    public constructor() {
        this.localVideoElm = document.getElementById("local_video") as HTMLVideoElement;
        this.remoteVideoElm = document.getElementById("remote_video") as HTMLVideoElement;
    }
    public setLocalVideo() {
        navigator.mediaDevices.getUserMedia({ video: true, audio: true })
            .then(stream => this.localVideoElm.srcObject = stream)
            .catch(err => console.error(err));
    }
    public setRemoteVideo(stream: MediaStream) {
        this.remoteVideoElm.srcObject = stream;
    }
}