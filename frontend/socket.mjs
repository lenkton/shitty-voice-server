import { userId } from './webrtc.mjs';

var socket = new WebSocket('/socket');

socket.onopen = () => {
    console.log("opened websocket");
    socket.send(JSON.stringify({type: 'login', userId: userId}));
};

socket.onmessage = (e) => {
    console.log('got a ws message');
    console.log(e.data);
};
