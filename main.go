package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/pion/webrtc/v4"
)

var api *webrtc.API

var remoteTrack *webrtc.TrackRemote
var localTrack *webrtc.TrackLocalStaticRTP

func handleOffer(w http.ResponseWriter, r *http.Request) {
	pc, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Println(err)
		return
	}
	pc.OnTrack(func(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		log.Println("INFO: wow, we have an audio track!")
		remoteTrack = tr
		// WARN: maybe it should go somewhere else...
		//       we must be sure that we have both local
		//       and remote streams at the same time
		go func() {
			for {
				rtp, _, err := remoteTrack.ReadRTP()
				if err != nil {
					log.Printf("ERROR: remoteTrack.ReadRTP: %v\n", err)
					break
				}
				err = localTrack.WriteRTP(rtp)
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

	localTrack, _ = webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus}, "audio", "shit")
	pc.AddTrack(localTrack)

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

func main() {
	settingEngine := webrtc.SettingEngine{}

	// Enable support only for TCP ICE candidates.
	settingEngine.SetNetworkTypes([]webrtc.NetworkType{
		webrtc.NetworkTypeTCP4,
		webrtc.NetworkTypeTCP6,
	})

	tcpListener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.IP{0, 0, 0, 0},
		Port: 8443,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Listening for ICE TCP at %s\n", tcpListener.Addr())

	tcpMux := webrtc.NewICETCPMux(nil, tcpListener, 8)
	settingEngine.SetICETCPMux(tcpMux)

	api = webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine))

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("POST /offer", handleOffer)
	log.Println("INFO: the app is listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
