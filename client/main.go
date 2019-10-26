package main

//./client -UIPort=10000 -msg=Hello

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

	flag.Parse()

	sendPrivateMessage := *dest != "" && *msg != ""             //ex3: uiport, dest, msg
	indexFileLocally := *file != ""                             //ex4: uiport, file -> index message locally
	requestFile := *dest != "" && *file != "" && *request != "" //ex6: uiport,dest,file, request

	fileHash := make([]byte, 1024)
	if *request != "" {
		_, err := hex.Decode(fileHash, []byte(*request))
		if err != nil {
			fmt.Println("​ ERROR (Unable to decode hex hash)")
			os.Exit(1)
		}
	}

	if indexFileLocally {
		gossiper.ReadFileIntoChunks(*file)
	} else if sendPrivateMessage || requestFile {
		gossiper.SendClientMessage(msg, uiport, dest, &fileHash, file)
	} else {
		fmt.Println("ERROR (Bad argument combination)")
		os.Exit(1)
	}

}
