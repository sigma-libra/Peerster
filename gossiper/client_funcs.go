package gossiper

import (
	"encoding/hex"
	"fmt"
	"github.com/dedis/protobuf"
	"net"
)

type Message struct {
	Text        string
	Destination *string
	File        *string
	Request     *[]byte
}

func SendClientMessage(msg *string, uiport *string, dest *string, fileHash *[]byte, file *string) {

	fmt.Println("Hash at client sending: " + hex.EncodeToString(*fileHash))

	packet := Message{
		Text:        *msg,
		Destination: dest,
		File:        file,
		Request:     fileHash,
	}
	packetToSend, err := protobuf.Encode(&packet)
	if err != nil {
		print("Client Encode Error: " + err.Error() + "\n")
	}

	randomPort := "0"
	clientUdpAddr, err := net.ResolveUDPAddr("udp4", "localhost:"+randomPort)
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
