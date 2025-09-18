package users

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/pion/webrtc/v4"
)

type UsersService struct {
	api  *webrtc.API
	user *User
}

func NewService(api *webrtc.API) *UsersService {
	return &UsersService{api: api, user: &User{}}
}

func (us *UsersService) HTTPHandleOffer(w http.ResponseWriter, r *http.Request) {
	pc, err := us.api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Println(err)
		return
	}
	pc.OnTrack(func(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		log.Println("INFO: wow, we have an audio track!")
		us.user.remoteTrack = tr
		// WARN: maybe it should go somewhere else...
		//       we must be sure that we have both local
		//       and remote streams at the same time
		go func() {
			for {
				rtp, _, err := us.user.remoteTrack.ReadRTP()
				if err == io.EOF {
					log.Println("INFO: end of the remote track")
					break
				}
				if err != nil {
					log.Printf("ERROR: remoteTrack.ReadRTP: %v\n", err)
					break
				}
				err = us.user.localTrack.WriteRTP(rtp)
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
	var offer webrtc.SessionDescription
	err = json.NewDecoder(r.Body).Decode(&offer)
	if err != nil {
		log.Println(err)
		return
	}
	err = pc.SetRemoteDescription(offer)
	if err != nil {
		log.Println(err)
		return
	}

	us.user.localTrack, _ = webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "shit")
	pc.AddTrack(us.user.localTrack)

	gatherPromise := webrtc.GatheringCompletePromise(pc)
	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		log.Println(err)
		return
	}
	err = pc.SetLocalDescription(answer)
	if err != nil {
		log.Println(err)
		return
	}
	<-gatherPromise
	err = json.NewEncoder(w).Encode(*pc.LocalDescription())
	if err != nil {
		log.Println(err)
		return
	}
}
