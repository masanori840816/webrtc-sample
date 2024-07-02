declare global {
    interface Window {
        Page: MainPageApi,
    };
}
export interface MainPageApi {
    init: (baseUrl: string) => void,
    connect: () => void,
    send: () => void,
    close: () => void,
    sendDataChannel: () => void,
}