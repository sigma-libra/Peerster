package gossiper

import (
	"encoding/json"
	"fmt"
	"github.com/dedis/protobuf"
	"net"
	"net/http"
)


func FormatPeers(peerSlice []string) string {
	peers := ""
	for _, peer := range peerSlice {
		peers = peers + peer + ","
	}
	peers = peers[:len(peers)-1]
	return peers
}

func getAndDecodePacket(gossip *Gossiper) (GossipPacket, string) {
	packetBytes := make([]byte, 1024)
	_, sender, err := gossip.conn.ReadFromUDP(packetBytes)
	if err != nil {
		print("Gossiper funcs Read Error: " + err.Error() + "\n")
	}

	pkt := GossipPacket{}
	protobuf.Decode(packetBytes, &pkt)
	return pkt, sender.String()
}

func sendPacket(pkt []byte, dst string, gossip *Gossiper) {
	udpAddr, err := net.ResolveUDPAddr("udp4", dst)
	if err != nil {
		println("Gossiper Funcs Resolve UDP Address Error: " + err.Error())
	}
	_, err = gossip.conn.WriteToUDP(pkt, udpAddr)
	if err != nil {
		println("Gossiper Funcs Write to UDP Error: " + err.Error())
	}
}

func isRumorPacket(pkt GossipPacket) bool {
	pktHasSimpleMsg := (pkt.Simple != nil) && (pkt.Rumor == nil) && (pkt.Status == nil)
	pktHasRumorMsg := (pkt.Simple == nil) && (pkt.Rumor != nil) && (pkt.Status == nil)
	pktHasStatusMsg := (pkt.Simple == nil) && (pkt.Rumor == nil) && (pkt.Status != nil)

	if !pktHasSimpleMsg && !pktHasRumorMsg && !pktHasStatusMsg {
		panic("Packet has no message")
	}

	return pktHasRumorMsg
}

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
		println("Caught post!")
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
		newMessage := r.FormValue("newMessage")
		SendClientMessage(&newMessage, &PeerUIPort)
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
	}
}
