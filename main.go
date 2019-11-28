package main

//./server -UIPort=10000 -gossipAddr=127.0.0.1:5000 -name=nodeA
// -peers=127.0.0.1:5001,10.1.1.7:5002 -simple

import (
	"flag"
	"github.com/SabrinaKall/Peerster/server"
	"net/http"
	"strings"
)

func main() {

	/*
		./gossiper ​ -gossipAddr=<> -name=<> ​ -peers=<> -antiEntropy=<>
		-​ hw3ex2=true -N=<> [-stubbornTimeout=<timeout in seconds>]
		[-hopLimit=<hop limit for TLCAck>]
	*/

	uiport := flag.String("UIPort",
		"8080", "port for the UI client (default \"8080\")")
	gossipAddr := flag.String("gossipAddr",
		"127.0.0.1:5000", "ip:port for the server (default \"127.0.0.1:5000\")")
	name := flag.String("name", "", "name of the server")
	peers := flag.String("peers", "", "comma separated list of peers of the form ip:port")
	simple := flag.Bool("simple", false, "run server in simple broadcast mode")
	antiEntropy := flag.Int("antiEntropy", 10, "Use the given timeout in seconds for anti-entropy. If the flag is absent, the default anti-entropy duration is 10 seconds.")
	guiport := flag.String("GUIPort", "8080", "Port for the graphical interface")
	rtimer := flag.Int("rtimer", 0, "Timeout in seconds to send route rumors. 0 (default) means disable sending route rumors.")
	n := flag.Int("N", 1, "the total number of peers in the network, including the current one")
	stubbornTimeout := flag.Int("stubbornTimeout", 5, "resend time between messages to receive acks")
	hoplimit := flag.Int("hopLimit", int(server.HOP_LIMIT), "hop limit for private messages")

	Hw3Ex2 := flag.Bool("hw3ex2", false, "Whether this is the blockchain exercice")
	Hw3Ex3 := flag.Bool("hw3ex3", false, "Whether this is the blockchain exercice")
	Hw3Ex4 := flag.Bool("hw3ex4", false, "Whether this is the blockchain exercice")
	flag.Parse()

	server.PeerName = *name //for handler
	server.PeerUIPort = *uiport
	server.AntiEntropy = *antiEntropy
	server.RTimer = *rtimer
	server.N = *n
	server.StubbornTimeout = *stubbornTimeout
	server.Hoplimit = uint32(*hoplimit)
	server.Ex2 = *Hw3Ex2
	server.Ex3 = *Hw3Ex3
	server.Ex4 = *Hw3Ex4

	peerGossiper := *server.NewGossiper(*gossipAddr, *name)
	clientGossiper := *server.NewGossiper("localhost:"+*uiport, *name)

	server.PeerGossiper = &peerGossiper

	knownPeers := make([]string, 0)
	if *peers != "" {
		knownPeers = strings.Split(*peers, ",")
	}

	for _, peer := range knownPeers {
		server.AddPeer(peer)
	}

	server.SetNodeID(*name, *uiport, *guiport)

	if *simple {
		go server.HandleSimpleMessagesFrom(&peerGossiper, gossipAddr)
		go server.HandleSimpleClientMessagesFrom(&clientGossiper, gossipAddr, &peerGossiper)

	} else {
		go server.HandleRumorMessagesFrom(&peerGossiper)
		go server.HandleClientRumorMessages(&clientGossiper, *name, &peerGossiper)
		go server.FireAntiEntropy(&peerGossiper)

	}

	setUpWindow(*guiport)

}

func setUpWindow(guiport string) {
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/id", server.GetIdHandler)
	http.HandleFunc("/message", server.GetLatestRumorMessagesHandler)
	http.HandleFunc("/private_message", server.GetLatestMessageableNodesHandler)
	http.HandleFunc("/nodes", server.GetLatestNodesHandler)
	http.HandleFunc("/uploadFile", server.GetFileUploadHandler)
	http.HandleFunc("/download", server.GetFileDownloadHandler)
	http.HandleFunc("/matchingfiles", server.GetLatestMatchingFilesHandler)
	for {

		err := http.ListenAndServe("localhost:"+guiport, nil)
		if err == nil {
			println("Frontend err: " + err.Error())
		}

	}
}
