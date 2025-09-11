var button = document.getElementById('start-button');
button.onclick = start;
var pc;
var track;
var audio = document.getElementById('audio');

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
        track = (e.track);
        let remoteStream = new MediaStream();
        remoteStream.addTrack(track);
        audio.srcObject = remoteStream;
        audio.play();
    }

    // let dc = pc.createDataChannel('data');
    navigator.mediaDevices.getUserMedia({audio: true})
        .then(media => {
            pc.addTrack(media.getTracks()[0], media)
        })
        .then(() => 
    pc.createOffer())
        .then(offer => {
            return pc.setLocalDescription(offer)
                .then(() => {
                    return fetch('/offer', {
                        method: 'POST',
                        body: JSON.stringify(offer)
                    })
                });
        })
        .then(res => res.json())
        .then(res => pc.setRemoteDescription(res))
        .catch(console.log);
}