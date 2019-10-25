package main

//./gossiper -UIPort=10000 -gossipAddr=127.0.0.1:5000 -name=nodeA
// -peers=127.0.0.1:5001,10.1.1.7:5002 -simple

import (
	"flag"
	"github.com/SabrinaKall/Peerster/gossiper"
	"net/http"
	"strings"
)

var name *string
var gossipAddr *string
var uiport *string
var guiport *string


func main() {

	uiport = flag.String("UIPort",
		"8080", "port for the UI client (default \"8080\")")
	gossipAddr = flag.String("gossipAddr",
		"127.0.0.1:5000", "ip:port for the gossiper (default \"127.0.0.1:5000\")")
	name = flag.String("name", "", "name of the gossiper")
	peers := flag.String("peers", "", "comma separated list of peers of the form ip:port")
	simple := flag.Bool("simple", false, "run gossiper in simple broadcast mode")
	antiEntropy := flag.Int("antiEntropy", 10, "Use the given timeout in seconds for anti-entropy. If the flag is absent, the default anti-entropy duration is 10 seconds.")
	guiport = flag.String("GUIPort", "8080", "Port for the graphical interface")
	rtimer := flag.Int("rtimer", 0, "Timeout in seconds to send route rumors. 0 (default) means disable sending route rumors.")

	flag.Parse()

	gossiper.PeerName = *name
	gossiper.PeerUIPort = *uiport
	gossiper.AntiEntropy = *antiEntropy
	gossiper.RTimer = *rtimer

	peerGossiper := *gossiper.NewGossiper(*gossipAddr, *name)
	clientGossiper := *gossiper.NewGossiper("localhost:"+*uiport, *name)

	knownPeers := make([]string, 0)
	if *peers != "" {
		knownPeers = strings.Split(*peers, ",")
	}

	for _, peer := range knownPeers {
		gossiper.AddPeer(peer)
	}

	if *simple {
		go gossiper.HandleSimpleMessagesFrom(&peerGossiper, name, gossipAddr)
		go gossiper.HandleSimpleClientMessagesFrom(&clientGossiper, name, gossipAddr, &peerGossiper)

	} else {
		go gossiper.HandleRumorMessagesFrom(&peerGossiper)
		go gossiper.HandleClientRumorMessages(&clientGossiper, *name, &peerGossiper)
		go gossiper.FireAntiEntropy(&peerGossiper)

	}

	setUpWindow()

}

func setUpWindow() {
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/id", gossiper.GetIdHandler)
	http.HandleFunc("/message", gossiper.GetLatestRumorMessagesHandler)
	http.HandleFunc("/private_message", gossiper.GetLatestMessageableNodesHandler)
	http.HandleFunc("/nodes", gossiper.GetLatestNodesHandler)
	http.HandleFunc("/uploadFile", gossiper.GetFileUploadHandler)
	for {

		err := http.ListenAndServe( "localhost:" + *guiport, nil)
		if err == nil {
			println("Frontend err: " + err.Error())
		}

	}
}


