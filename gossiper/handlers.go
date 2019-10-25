package gossiper

import (
	"encoding/json"
	"net/http"
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
		SendClientMessage(&newMessage, &PeerUIPort, &dst)
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
		SendClientMessage(&newMessage, &PeerUIPort, &dst)
	default:
		println(w, "Sorry, only GET and POST methods are supported.")
	}
}
