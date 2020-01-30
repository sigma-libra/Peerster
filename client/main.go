//Author: Sabrina Kall
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/SabrinaKall/Peerster/gossiper"
	"os"
)

func main() {
	uiport := flag.String("UIPort",
		"8080", "port for the UI client (default \"8080\")")
	msg := flag.String("msg", "", "message to be sent")
	dest := flag.String("dest", "", "destination for the private message; ​ can be omitted")
	request := flag.String("request", "", "request a chunk or metafile of this hash")
	file := flag.String("file", "", "file to be indexed by the gossiper")
	clientGroups := flag.String("groups", "", "Client message groups (group1,group2,...")

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
			gossiper.SendClientMessage(msg, uiport, dest, &fileHash, file, nil)
			return
		}
	}

	if indexFileLocally || sendPrivateMessage || sendRumorMessage {
		gossiper.SendClientMessage(msg, uiport, dest, nil, file, clientGroups)
	} else {
		fmt.Println("ERROR (Bad argument combination)")
		os.Exit(1)
	}

}
