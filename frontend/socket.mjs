import { handleOffer, userId } from './webrtc.mjs';

var socket = new WebSocket('/socket');

socket.onopen = () => {
    console.log("opened websocket");
    socket.send(JSON.stringify({type: 'login', userId: userId}));
};

socket.onmessage = (e) => {
    console.log('got a ws message');
    console.log(e.data);
    let parsed = JSON.parse(e.data);
    switch (parsed.type) {
        case "offer":
            handleOfferSocket(parsed.sdp);
            break;
        default:
            console.log('unknown message type: ', parsed['type']);
            break;
    }
};

function handleOfferSocket(sdp) {
    handleOffer(sdp);
}
