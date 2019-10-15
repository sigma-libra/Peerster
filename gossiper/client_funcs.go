package gossiper

import (
	"github.com/dedis/protobuf"
	"math/rand"
	"net"
	"strconv"
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

	randomPort := "90" + strconv.Itoa(rand.Intn(50) + 10)
	clientUdpAddr, err := net.ResolveUDPAddr("udp4", "localhost:" + randomPort)
	gossiperUdpAddr, err := net.ResolveUDPAddr("udp4", "localhost:"+*uiport)
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