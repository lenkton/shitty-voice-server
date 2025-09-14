package main

import (
	"echo-webrtc-test/pkg/services/users"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/pion/webrtc/v4"
)

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

	api := webrtc.NewAPI(webrtc.WithSettingEngine(settingEngine))
	usersService := users.NewService(api)

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	// TODO: move http-handling into some separate package
	http.HandleFunc("POST /offer", usersService.HTTPHandleOffer)
	log.Println("INFO: the app is listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
