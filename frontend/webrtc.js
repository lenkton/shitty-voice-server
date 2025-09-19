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
