package gossiper

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func GetIdHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		id := "Name: " + PeerName + "(Port: " + PeerUIPort + ")"
		idJSON, err := json.Marshal(id)
		if err != nil {
			println("frontend error: " + err.Error())
		}
		// error handling, etc...
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(idJSON)
		if err != nil {
			println("Frontend Error - Get id handler: " + err.Error())
		}
	}
}

func GetLatestRumorMessagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		messagesJson, err := json.Marshal(messages)
		if err != nil {
			println("frontend error: " + err.Error())
		}
		// error handling, etc...
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(messagesJson)
		if err != nil {
			println("Frontend Error - Get message handler: " + err.Error())
		}
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			println(w, "ParseForm() err: %v", err)
			return
		}
		//fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
		newMessage := r.FormValue("newMessage")
		dst := ""
		SendClientMessage(&newMessage, &PeerUIPort, &dst, nil, nil)
	default:
		println(w, "Sorry, only GET and POST methods are supported.")
	}
}

func GetLatestNodesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		nodesJson, err := json.Marshal(nodes)
		if err != nil {
			println("frontend error: " + err.Error())
		}
		// error handling, etc...
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(nodesJson)
		if err != nil {
			println("Frontend Error - Get nodes handler: " + err.Error())
		}
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			println(w, "ParseForm() err: %v", err)
			return
		}
		//fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
		newNode := r.FormValue("newNode")
		AddPeer(newNode)
	default:
		println(w, "Sorry, only GET and POST methods are supported.")
	}
}

func GetLatestMessageableNodesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		nodesJson, err := json.Marshal(parseRoutingTable())
		if err != nil {
			println("frontend error: " + err.Error())
		}
		// error handling, etc...
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(nodesJson)
		if err != nil {
			println("Frontend Error - Get nodes handler: " + err.Error())
		}
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			println(w, "ParseForm() err: %v", err)
			return
		}
		//fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
		newMessage := r.FormValue("newMessage")
		dst := r.FormValue("dest")
		SendClientMessage(&newMessage, &PeerUIPort, &dst, nil, nil)
	default:
		println(w, "Sorry, only GET and POST methods are supported.")
	}
}

func GetFileUploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var Buf bytes.Buffer
		// in your case file would be fileupload
		file, header, err := r.FormFile("file")
		if err != nil {
			println("Write to file err: " + err.Error())
		}
		defer file.Close()
		name := strings.Split(header.Filename, ".")
		fmt.Printf("File name %s\n", name[0])
		// Copy the file data to my buffer
		io.Copy(&Buf, file)
		// do something with the contents...
		// I normally have a struct defined and unmarshal into a struct, but this will
		// work as an example
		contents := Buf.String()
		fmt.Println(contents)
		err = writeToFile(header.Filename, contents)
		if err != nil {
			println("Write to file err: " + err.Error())
		}
		ReadFileIntoChunks(header.Filename)
		// I reset the buffer in case I want to use it again
		// reduces memory allocations in more intense projects
		Buf.Reset()
		// do something else
		// etc write header
	default:
		println(w, "Sorry, only POST method is supported.")
	}

}

func GetFileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		if err := r.ParseForm(); err != nil {
			println(w, "ParseForm() err: %v", err)
			return
		}
		//fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
		dst := r.FormValue("dst")
		hash := r.FormValue("hash")
		name := r.FormValue("name")

		if hash != "" {
			fileHash, err := hex.DecodeString(hash)
			if err != nil {
				fmt.Println("​ ERROR (Unable to decode hex hash)")
				os.Exit(1)
			}
			SendClientMessage(nil, &PeerUIPort, &dst, &fileHash, &name)
		}

	default:
		println(w, "Sorry, only POST method is supported.")
	}

}

func writeToFile(filename string, data string) error {
	file, err := os.Create("./_SharedFiles/" + filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, data)
	if err != nil {
		return err
	}
	return file.Sync()
}
