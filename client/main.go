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

const clientAddr = "127.0.0.1:10000"

func main() {
	uiport := flag.String("UIPort",
		"8080", "port for the UI client (default \"8080\")")
	msg := flag.String("msg", "", "message to be sent")

	flag.Parse()

	simplePacket := SimpleMessage{"client",clientAddr,*msg}
	packetToSend, err := protobuf.Encode(&GossipPacket{&simplePacket})
	if err != nil {
		print("Client Encode Error: " +  err.Error() + "\n")
	}

	msgTest := GossipPacket{}
	err = protobuf.Decode(packetToSend, &msgTest)
	if err != nil {
		print("Client Protobuf Decode Error: " + err.Error() + "\n")
	}


	clientUdpAddr, err := net.ResolveUDPAddr("udp4", clientAddr)
	gossiperUdpAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:" + *uiport)
	if err != nil {
		print("Client Resolve Addr Error: " + err.Error() + "\n")
	}
	udpConn, err := net.ListenUDP("udp4", clientUdpAddr)
	if err != nil {
		print("Client ListenUDP Error: " +  err.Error() + "\n")
	}
	_, err = udpConn.WriteToUDP(packetToSend, gossiperUdpAddr)
	if err != nil {
		print("Client Write To UDP: " + err.Error() + "\n")
	}


}
