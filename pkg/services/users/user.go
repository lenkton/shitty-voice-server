package users

import (
	"echo-webrtc-test/pkg/socket"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/pion/webrtc/v4"
)

// TODO: hide the fields and introduce some DTO
type User struct {
	ID          string `json:"id"`
	remoteTrack *webrtc.TrackRemote
	localTrack  *webrtc.TrackLocalStaticRTP
	pc          *webrtc.PeerConnection
}

func NewUser(id string) *User {
	return &User{ID: id}
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

	pc.OnTrack(u.onTrack)
	pc.OnNegotiationNeeded(func() {
		log.Println("ERROR: NEGOTIATION NEEDED!!!!!")
		offer, err := pc.CreateOffer(nil)
		if err != nil {
			log.Printf("ERROR: creating offer: %v\n", err)
			return
		}

		err = pc.SetLocalDescription(offer)
		if err != nil {
			log.Printf("ERROR: setting local description: %v\n", err)
			return
		}

		err = socket.SendMessage(u.ID, map[string]any{
			"type": "offer",
			"sdp":  offer,
		})
		if err != nil {
			log.Printf("ERROR: sending offer: %v\n", err)
		}

		log.Printf("INFO: Renegotiation offer sent to %v\n", u.ID)
	})
	return nil
}

// TODO: maybe we should not return a pointer here
func (u *User) HandleOffer(offer webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	err := u.pc.SetRemoteDescription(offer)
	if err != nil {
		return nil, fmt.Errorf("pc.SetRemoteDescription: %v", err)
	}

	u.localTrack, _ = webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", u.ID)
	// u.pc.AddTrack(u.localTrack)
	// u.pc.AddStream()

	gatherPromise := webrtc.GatheringCompletePromise(u.pc)
	answer, err := u.pc.CreateAnswer(nil)
	if err != nil {
		return nil, fmt.Errorf("pc.CreateAnswer: %v", err)
	}
	err = u.pc.SetLocalDescription(answer)
	if err != nil {
		return nil, fmt.Errorf("pc.SetLocalDescription: %v", err)
	}
	<-gatherPromise
	// TODO: check, why do we have a pointer here,
	//       while passing the object all around in other places
	resAnswer := u.pc.LocalDescription()
	return resAnswer, nil
}

func (u *User) onTrack(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
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
		err := u.pc.Close()
		if err != nil {
			log.Printf("ERROR: pc.Close: %v\n", err)
		}
		// TODO: do we need to stop the local track?
	}()
}
