package frontend

import (
	"encoding/json"
	"github.com/SabrinaKall/Peerster/gossiper"
	"net/http"
	"time"
)

func main() {


	for{
		time.Sleep(2*time.Millisecond)
	}

}

var correctIpAddressInformation string = "127.0.0.1:8080"


func getIdHandler(w http.ResponseWriter, r *http.Request) {

}

func getLatestRumorMessagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		msgList := gossiper.GetLatestRumorMessagesList()
		msgListJson, err := json.Marshal(msgList)
		if err != nil {
			println("frontend error: "+ err.Error())
		}
		// error handling, etc...
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(msgListJson)
	}
}
