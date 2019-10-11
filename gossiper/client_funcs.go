package gossiper

import (
	"github.com/dedis/protobuf"
	"net"
)

type Message struct {
	Text string
}


func SendClientMessage(msg *string, uiport *string) {
	packet := Message{*msg}
	packetToSend, err := protobuf.Encode(&packet)
	if err != nil {
		print("Client Encode Error: " + err.Error() + "\n")
	}

	clientUdpAddr, err := net.ResolveUDPAddr("udp4", clientAddr)
	gossiperUdpAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:"+*uiport)
	if err != nil {
		println("Client Resolve Addr Error: " + err.Error())
	}
	udpConn, err := net.ListenUDP("udp4", clientUdpAddr)
	if err != nil {
		print("Client ListenUDP Error: " + err.Error() + "\n")
	}
	_, err = udpConn.WriteToUDP(packetToSend, gossiperUdpAddr)
	if err != nil {
		println("Client Write To UDP: " + err.Error())
	}

	udpConn.Close()
}