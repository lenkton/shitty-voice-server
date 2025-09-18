package users

import "github.com/pion/webrtc/v4"

type User struct {
	remoteTrack *webrtc.TrackRemote
	localTrack  *webrtc.TrackLocalStaticRTP
}
