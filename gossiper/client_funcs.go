package gossiper

import (
	"github.com/dedis/protobuf"
	"net"
	"strings"
)

type Message struct {
	Text        string
	Destination *string
	File        *string
	Request     *[]byte
	Groups      []string //project
}

func SendClientMessage(msg *string, uiport *string, dest *string, fileHash *[]byte, file *string, groups *string) {

	group_list := make([]string, 0)
	if groups != nil {
		group_list = strings.Split(*groups, ",")
	}
	//To test file sending
	packet := Message{
		Text:        *msg,
		Destination: dest,
		File:        file,
		Request:     fileHash,
		Groups:      group_list,
	}
	packetToSend, err := protobuf.Encode(&packet)
	printerr("Client Encode Error", err)

	randomPort := "0"
	clientUdpAddr, err := net.ResolveUDPAddr("udp4", "localhost:"+randomPort)
	gossiperUdpAddr, err := net.ResolveUDPAddr("udp4", "localhost:"+*uiport)
	printerr("Client Resolve Addr Error", err)

	udpConn, err := net.ListenUDP("udp4", clientUdpAddr)
	printerr("Client ListenUDP Error", err)

	_, err = udpConn.WriteToUDP(packetToSend, gossiperUdpAddr)
	printerr("Client Write To UDP", err)

	err = udpConn.Close()
	printerr("Client Close connection", err)
}
