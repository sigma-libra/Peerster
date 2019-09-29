package client

//./client -UIPort=10000 -msg=Hello

import (
	"flag"
	"net"
	"github.com/SabrinaKall/Peerster/gossiper"
	"github.com/dedis/protobuf"
)

func main() {
	uiport := flag.String("UIPort",
		"8080", "port for the UI client (default \"8080\")")
	msg := flag.String("msg", "", "message to be sent")

	flag.Parse()

	simplePacket := gossiper.SimpleMessage{"","",*msg}
	packetToSend, _ := protobuf.Encode(gossiper.GossipPacket{&simplePacket})

	udpAddr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:" + *uiport)
	udpConn, _ := net.ListenUDP("udp4", udpAddr)
	udpConn.WriteToUDP(packetToSend, udpAddr)


}
