package main

//./gossiper -UIPort=10000 -gossipAddr=127.0.0.1:5000 -name=nodeA
// -peers=127.0.0.1:5001,10.1.1.7:5002 -simple

import (
	"flag"
	"github.com/SabrinaKall/Peerster/gossiper"
	"net/http"
	"strings"
	"time"
)

var name *string
var gossipAddr *string
var knownPeers []string
var uiport *string

const clientAddress = "127.0.0.1"

func main() {

	uiport = flag.String("UIPort",
		"8080", "port for the UI client (default \"8080\")")
	gossipAddr = flag.String("gossipAddr",
		"127.0.0.1:5000", "ip:port for the gossiper (default \"127.0.0.1:5000\")")
	name = flag.String("name", "", "name of the gossiper")
	peers := flag.String("peers", "", "comma separated list of peers of the form ip:port")
	simple := flag.Bool("simple", false, "run gossiper in simple broadcast mode")

	flag.Parse()

	gossiper.PeerName = *name
	gossiper.PeerUIPort = *uiport

	peerGossip := gossiper.NewGossiper(*gossipAddr, *name)
	clientGossip := gossiper.NewGossiper(clientAddress+":"+*uiport, *name)

	knownPeers := make([]string, 0)
	if *peers != "" {
		knownPeers = strings.Split(*peers, ",")
	}

	chatBoxMessages := make(chan string, 1000)
	peerSharingChan := make(chan string, 1000)
	if *simple {
		go gossiper.HandleSimpleMessagesFrom(peerGossip, false, name, gossipAddr, knownPeers, peerSharingChan, chatBoxMessages)
		go gossiper.HandleSimpleMessagesFrom(clientGossip, true, name, gossipAddr, knownPeers, peerSharingChan, chatBoxMessages)

	} else {
		wantsUpdate := make(chan gossiper.PeerStatus, 1000)
		go gossiper.HandleRumorMessagesFrom(peerGossip, *name, knownPeers, false, peerSharingChan, wantsUpdate, chatBoxMessages)
		go gossiper.HandleRumorMessagesFrom(clientGossip, *name, knownPeers, true, peerSharingChan, wantsUpdate, chatBoxMessages)
		go gossiper.FireAntiEntropy(knownPeers, peerSharingChan, wantsUpdate, peerGossip)

	}

	go setUpWindow(chatBoxMessages)

	for {
		time.Sleep(2 * time.Millisecond)
	}

}

func setUpWindow(chatBoxMessages chan string) {
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/id", gossiper.GetIdHandler)
	http.HandleFunc("/message", gossiper.GetLatestRumorMessagesHandler)
	for {

		err := http.ListenAndServe("127.0.0.1:8080", nil)
		if err == nil {
			println("Frontend err: " + err.Error())
		}

	}
}


