declare global {
    interface Window {
        Page: MainPageApi,
    };
}
export interface MainPageApi {
    connect: () => void,
    sendOffer: () => void,
    send: () => void,
    close: () => void,
    sendDataChannel: () => void,
}