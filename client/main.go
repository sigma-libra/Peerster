package main

//./client -UIPort=10000 -msg=Hello

import (
	"flag"
	"github.com/dedis/protobuf"
	"net"
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

	simplePacket := SimpleMessage{"","",*msg}
	packetToSend, err := protobuf.Encode(&GossipPacket{&simplePacket})
	if err != nil {
		print("Client Encode Error: " +  err.Error() + "\n")
	}

	udpAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:" + *uiport)
	if err != nil {
		print("Client Resolve Addr Error: " + err.Error() + "\n")
	}
	udpConn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		print("Client ListenUDP Error: " +  err.Error() + "\n")
	}
	_, err = udpConn.WriteToUDP(packetToSend, udpAddr)
	if err != nil {
		print("Client Write To UDP: " + err.Error() + "\n")
	}


}
