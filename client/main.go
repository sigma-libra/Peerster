package main

//./client -UIPort=10000 -msg=Hello

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/dedis/protobuf"
	"net"
	"os"
)


type Message struct {
	Text        string
	Destination *string
	File        *string
	Request     *[]byte
}

func main() {
	uiport := flag.String("UIPort",
		"8080", "port for the UI client (default \"8080\")")
	msg := flag.String("msg", "", "message to be sent")
	dest := flag.String("dest", "", "destination for the private message; ​ can be omitted")
	request := flag.String("request", "", "request a chunk or metafile of this hash")
	file := flag.String("file", "", "file to be indexed by the gossiper")

	flag.Parse()

	sendPrivateMessage := *dest != "" && *msg != "" //ex3: uiport, dest, msg
	sendRumorMessage := *msg != ""
	indexFileLocally := *file != "" && *request == ""           //ex4: uiport, file -> index message locally
	requestFile := *dest != "" && *file != "" && *request != "" //ex6: uiport,dest,file, request

	if *request != "" {
		fileHash, err := hex.DecodeString(*request)
		if err != nil {
			fmt.Println("​ ERROR (Unable to decode hex hash)")
			os.Exit(1)
		}
		if requestFile {
			SendClientMessage(msg, uiport, dest, &fileHash, file)
			return
		}
	}

	if indexFileLocally || sendPrivateMessage || sendRumorMessage {
		SendClientMessage(msg, uiport, dest, nil, file)
	} else {
		fmt.Println("ERROR (Bad argument combination)")
		os.Exit(1)
	}

}

func SendClientMessage(msg *string, uiport *string, dest *string, fileHash *[]byte, file *string) {

	//To test file sending

	//fmt.Println("Hash at client sending: " + hex.EncodeToString(*fileHash))
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
