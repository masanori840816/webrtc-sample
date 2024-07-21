declare global {
    interface Window {
        Page: MainPageApi,
    };
}
export interface MainPageApi {
    init: () => void,
    connect: () => void,
    sendOffer: () => void,
    send: () => void,
    close: () => void,
    sendDataChannel: () => void,
}