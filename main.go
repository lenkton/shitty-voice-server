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

func handleOffer(w http.ResponseWriter, r *http.Request) {
	pc, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Println(err)
		return
	}
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
	log.Fatal(http.ListenAndServe(":8080", nil))
}
