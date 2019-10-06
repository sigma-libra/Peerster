package frontend

import (
	"encoding/json"
	"github.com/SabrinaKall/Peerster/gossiper"
	"net/http"
)

func setUpWindow(){
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/messages", getLatestRumorMessagesHandler)
	for {
		err := http.ListenAndServe(correctIpAddressInformation, nil)
		// error handling, etc..
	}
}

// Somewhere, maybe even in a different package you create.
func getLatestRumorMessagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		msgList := gossiper.GetLatestRumorMessagesList()
		msgListJson, err := json.Marshal(msgList)
		// error handling, etc...
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(msgListJson)
	}
}
