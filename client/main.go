package main

//./client -UIPort=10000 -msg=Hello

import (
	"flag"
	"github.com/SabrinaKall/Peerster/gossiper"
)

func main() {
	uiport := flag.String("UIPort",
		"8080", "port for the UI client (default \"8080\")")
	msg := flag.String("msg", "", "message to be sent")
	dest := flag.String("dest", "", "destination for the private message; ​ can be omitted")

	flag.Parse()

	gossiper.SendClientMessage(msg, uiport, dest)

}

