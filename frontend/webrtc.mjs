import './socket.mjs';

var button = document.getElementById('start-button');
button.onclick = start;
var endButton = document.getElementById('end-button');
endButton.onclick = end;
var pc;
var localStream;
var audio = document.getElementById('audio');

export var userId = Math.floor(Math.random() * 100).toString();
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

    for (let track of localStream.getTracks()) track.enabled = false;
}
function unmute() {
    if (!localStream) return;

    for (let track of localStream.getTracks()) track.enabled = true;
}

// TODO: there is Promise.withResolvers()!
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/withResolvers
var resolveConnectedPromise;
var connectedPromise = new Promise((resolve, reject) => {
    resolveConnectedPromise = resolve;
});
function sendIceCandidate(candidate) {
    connectedPromise
        .then(() =>
            fetch(`/users/${userId}/ice`, {
                method: 'POST',
                body: JSON.stringify(candidate) }))
        .catch(console.log);
}

var dataChannel;
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
        let audioElement = document.createElement('audio');
        remoteStream.addTrack(track);
        audioElement.srcObject = remoteStream;
        document.body.appendChild(audioElement);
        audioElement.play();
    }
    pc.onicecandidate = (e) => {
        console.log('got ice candidate');
        console.log(e.candidate);
        if (e.candidate) {
            sendIceCandidate(e.candidate);
        }
    }

    // let dc = pc.createDataChannel('data');
    navigator.mediaDevices.getUserMedia({audio: true})
        .then(media => {
            localStream = media;
            pc.addTrack(media.getTracks()[0], media);
        })
        .then(()=> pc.createDataChannel('stub'))
        .then(channel => {dataChannel = channel;})
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
        .then(() => resolveConnectedPromise())
        .catch(console.log);
}

export function handleOffer(sdp) {
    pc.setRemoteDescription(sdp)
        .then(()=>pc.createAnswer())
        .then(answer => {
            return pc.setLocalDescription(answer)
                .then(() => fetch(`/users/${userId}/answer`, {
                        method: 'POST',
                        body: JSON.stringify(answer)
                    })
                );
        })
        .then(console.log)
        .catch(console.log);
    console.log('set remote offer')
}

function end() {
    console.log('hanging up');
    pc.close();
    pc = null;
    for (let track of localStream.getTracks()) track.stop();
    localStream = null;
}
