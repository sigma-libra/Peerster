package main

//./client -UIPort=10000 -msg=Hello

import (
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/SabrinaKall/Peerster/server"
	"os"
	"strings"
)

func main() {
	uiport := flag.String("UIPort",
		"8080", "port for the UI client (default \"8080\")")
	msg := flag.String("msg", "", "message to be sent")
	dest := flag.String("dest", "", "destination for the private message; ​ can be omitted")
	request := flag.String("request", "", "request a chunk or metafile of this hash")
	file := flag.String("file", "", "file to be indexed by the server")
	keywords := flag.String("keywords", "", "comma-separated string of keywords for filenames")
	budget := flag.Int("budget", 2, "initial budget to distribute keywords; can be omitted")

	flag.Parse()

	sendPrivateMessage := *dest != "" && *msg != "" //ex3: uiport, dest, msg
	sendRumorMessage := *msg != ""
	indexFileLocally := *file != "" && *request == ""           //ex4: uiport, file -> index message locally
	requestFile := *dest != "" && *file != "" && *request != "" //ex6: uiport,dest,file, request
	search := *keywords != ""

	keys := make([]string, 0)
	if *keywords != "" {
		keys = strings.Split(*keywords, ",")
	}
	budget64 := uint64(*budget)

	if *request != "" {
		fileHash, err := hex.DecodeString(*request)
		if err != nil {
			fmt.Println("​ ERROR (Unable to decode hex hash)")
			os.Exit(1)
		}
		if requestFile {
			server.SendClientMessage(msg, uiport, dest, &fileHash, file, &keys, &budget64)
			return
		}
	}

	if indexFileLocally || sendPrivateMessage || sendRumorMessage || search {
		server.SendClientMessage(msg, uiport, dest, nil, file, &keys, &budget64)
	} else {
		fmt.Println("ERROR (Bad argument combination)")
		os.Exit(1)
	}

}
