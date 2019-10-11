package main

//./client -UIPort=10000 -msg=Hello

import (
	"flag"
	"github.com/SabrinaKall/Peerster/gossiper"

)

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

type GossipPacket struct {
	Simple *SimpleMessage
}

func main() {
	uiport := flag.String("UIPort",
		"8080", "port for the UI client (default \"8080\")")
	msg := flag.String("msg", "", "message to be sent")

	flag.Parse()

	gossiper.SendClientMessage(msg, uiport)

}

