package users

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/pion/webrtc/v4"
)

type User struct {
	remoteTrack *webrtc.TrackRemote
	localTrack  *webrtc.TrackLocalStaticRTP
	pc          *webrtc.PeerConnection
}

var ErrPCAlreadyCreated = errors.New("peer connection already created")

func (u *User) CreatePeerConnection(api *webrtc.API) error {
	if u.pc != nil {
		return ErrPCAlreadyCreated
	}
	pc, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return fmt.Errorf("NewPeerConnection: %v", err)
	}
	u.pc = pc

	pc.OnTrack(func(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		log.Println("INFO: wow, we have an audio track!")
		u.remoteTrack = tr
		// WARN: maybe it should go somewhere else...
		//       we must be sure that we have both local
		//       and remote streams at the same time
		go func() {
			for {
				rtp, _, err := u.remoteTrack.ReadRTP()
				if err == io.EOF {
					log.Println("INFO: end of the remote track")
					break
				}
				if err != nil {
					log.Printf("ERROR: remoteTrack.ReadRTP: %v\n", err)
					break
				}
				err = u.localTrack.WriteRTP(rtp)
				if err != nil {
					log.Printf("ERROR: localTrack.WriteRTP: %v\n", err)
				}
			}
			err = pc.Close()
			if err != nil {
				log.Printf("ERROR: pc.Close: %v\n", err)
			}
			// TODO: do we need to stop the local track?
		}()
	})
	return nil
}
