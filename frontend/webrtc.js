var button = document.getElementById('start-button');
button.onclick = start;
var endButton = document.getElementById('end-button');
endButton.onclick = end;
var pc;
var localStream;
var audio = document.getElementById('audio');

var userId = Math.floor(Math.random() * 100);
var userIdMessage = document.getElementById('id-message');
userIdMessage.innerText = `Your id is ${userId}`;

var muteButton = document.getElementById('mute-button');
muteButton.onclick = mute;
var unmuteButton = document.getElementById('unmute-button');
unmuteButton.onclick = unmute;

var joinRoomButton = document.getElementById('join-room-button');
joinRoomButton.onclick = joinRoom;

// TODO: remove hardcoding
var roomId = 12;

var socket = new WebSocket('/socket');
socket.onopen = () => {
    console.log("opened websocket");
    socket.send("hello!");
};
socket.onmessage = (e) => {
    console.log('got a ws message');
    console.log(e.data);
};

function joinRoom(event) {
    fetch(`/rooms/${roomId}/join`, {
        method: 'POST', body: JSON.stringify({userId: userId.toString()})
    })
        .then(response => response.json())
        .then(console.log)
        .catch(console.log);
}

// TODO: disable the buttons when there is no localStream
function mute() {
    if (!localStream) return;

    for (track of localStream.getTracks()) track.enabled = false;
}
function unmute() {
    if (!localStream) return;

    for (track of localStream.getTracks()) track.enabled = true;
}

function start() {
    pc = new RTCPeerConnection();
    pc.onconnectionstatechange = (e) => {
        console.log('ice conn state: ', pc.iceConnectionState);
    }
    pc.onsignalingstatechange = () => {
        console.log('sig state changed: ',
            pc.signalingState);
    }
    // TODO: maybe it needs refactoring...
    pc.ontrack = (e) => {
        console.log('got the remote track');
        let track = (e.track);
        let remoteStream = new MediaStream();
        remoteStream.addTrack(track);
        audio.srcObject = remoteStream;
        audio.play();
    }

    // let dc = pc.createDataChannel('data');
    navigator.mediaDevices.getUserMedia({audio: true})
        .then(media => {
            localStream = media;
            pc.addTrack(media.getTracks()[0], media);
        })
        .then(() => pc.createOffer())
        .then(offer => {
            return pc.setLocalDescription(offer)
                .then(() => {
                    return fetch(`/users/${userId}/offer`, {
                        method: 'POST',
                        body: JSON.stringify(offer)
                    })
                });
        })
        .then(res => res.json())
        .then(res => pc.setRemoteDescription(res))
        .catch(console.log);
}

function end() {
    console.log('hanging up');
    pc.close();
    pc = null;
    for (track of localStream.getTracks()) track.stop();
    localStream = null;
}
