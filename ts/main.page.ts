import * as domains from "./appDomains";

window.Page = {
    init() {
        const wsUrl = domains.getWebSocketAddress();
        const httpUrl = domains.getBaseAddress();
        console.log(`WebS: ${wsUrl} http: ${httpUrl}`);
        
    },
    connect() {
        console.log("connect");
        
    },
    send() {

    },
    close() {

    },
    sendDataChannel() {

    }
};
