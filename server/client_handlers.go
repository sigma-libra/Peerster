package server

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

type IDStruct struct {
	Name    string
	Port    string
	Guiport string
}

func SetNodeID(name string, port string, guiport string) {
	NodeID = IDStruct{
		Name:    name,
		Port:    port,
		Guiport: guiport,
	}
}
func GetIdHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":

		//id := "Name: " + PeerName + "(Port: " + PeerUIPort + ")"
		idJSON, err := json.Marshal(NodeID)
		printerr("Frontend Error", err)
		// error handling, etc...
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(idJSON)
		printerr("Frontend Error", err)
	}
}

func GetLatestRumorMessagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		messagesJson, err := json.Marshal(messages)
		printerr("Frontend Error", err)
		// error handling, etc...
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(messagesJson)
		printerr("Frontend Error", err)
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			if debug {
				println(w, "ParseForm() err: %v", err)
			}
			return
		}
		newMessage := r.FormValue("newMessage")
		dst := ""
		SendClientMessage(&newMessage, &PeerUIPort, &dst, nil, nil, nil, nil)
	default:
		if debug {
			println(w, "Sorry, only GET and POST methods are supported.")
		}
	}
}


func GetLatestKeywordssHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			if debug {
				println(w, "ParseForm() err: %v", err)
			}
			return
		}
		newKeywordString := r.FormValue("keywords")
		keywords := strings.Split(newKeywordString, ",")
		dst := ""
		budget := uint64(2)
		SendClientMessage(nil, nil, &dst, nil, nil, &keywords, &budget)
	default:
		if debug {
			println(w, "Sorry, only GET and POST methods are supported.")
		}
	}
}

func GetLatestNodesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		nodesJson, err := json.Marshal(nodes)
		printerr("Frontend Error", err)
		// error handling, etc...
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(nodesJson)
		printerr("Frontend Error", err)
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			if debug {
				println(w, "ParseForm() err: %v", err)
			}
			return
		}
		newNode := r.FormValue("newNode")
		AddPeer(newNode)
	default:
		if debug {
			println(w, "Sorry, only GET and POST methods are supported.")
		}
	}
}

func GetLatestMessageableNodesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		nodesJson, err := json.Marshal(parseRoutingTable())
		printerr("Frontend Error", err)
		// error handling, etc...
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(nodesJson)
		printerr("Frontend Error", err)
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			if debug {
				println(w, "ParseForm() err: %v", err)
			}
			return
		}
		newMessage := r.FormValue("newMessage")
		dst := r.FormValue("dest")
		SendClientMessage(&newMessage, &PeerUIPort, &dst, nil, nil, nil, nil)
	default:
		if debug {
			println(w, "Sorry, only GET and POST methods are supported.")
		}
	}
}

func GetFileUploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var Buf bytes.Buffer
		// in your case file would be fileupload
		file, header, err := r.FormFile("file")
		printerr("Frontend Error", err)
		defer file.Close()
		// Copy the file data to my buffer
		io.Copy(&Buf, file)
		// do something with the contents...
		// I normally have a struct defined and unmarshal into a struct, but this will
		// work as an example
		contents := Buf.String()
		err = writeToFile(header.Filename, contents)
		printerr("Frontend Error", err)
		ReadFileIntoChunks(header.Filename)
		// I reset the buffer in case I want to use it again
		// reduces memory allocations in more intense projects
		Buf.Reset()
		// do something else
		// etc write header
	default:
		if debug {
			println(w, "Sorry, only POST method is supported.")
		}
	}

}

func GetFileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		if err := r.ParseForm(); err != nil {
			if debug {
				println(w, "ParseForm() err: %v", err)
			}
			return
		}
		dst := r.FormValue("dst")
		hash := r.FormValue("hash")
		name := r.FormValue("name")

		if hash != "" {
			fileHash, err := hex.DecodeString(hash)
			if err != nil {
				fmt.Println("​ ERROR (Unable to decode hex hash)")
				os.Exit(1)
			}
			emptyMsg := ""
			SendClientMessage(&emptyMsg, &PeerUIPort, &dst, &fileHash, &name)
		}

	default:
		if debug {
			println(w, "Sorry, only POST method is supported.")
		}
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
