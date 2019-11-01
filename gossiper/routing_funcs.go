package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
	"math/rand"
	"strconv"
	"time"
)

func FireRouteRumor(gossip *Gossiper) {
	if RTimer > 0 {
		for {
			SendRouteRumor(gossip)
			ticker := time.NewTicker(time.Duration(RTimer) * time.Second)
			<-ticker.C
		}
	}
}

func SendRouteRumor(gossip *Gossiper) {
	if RTimer > 0 && len(Keys) > 0 {
		randomPeer := Keys[rand.Intn(len(Keys))]
		sendPacket(makeRouteRumor(gossip), randomPeer, gossip)
	}
}

func makeRouteRumor(gossip *Gossiper) []byte {
	msg := RumorMessage{
		Origin: gossip.Name,
		ID:     getAndUpdateRumorID(),
		Text:   "",
	}

	newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: &msg})
	if err != nil {
		println("Route Rumor Error: " + err.Error())
	}
	return newEncoded
}

func parseRoutingTable() string {
	origins := ""
	for k, _ := range routingTable.Table {
		origins += "<span onclick='openMessageWindow((this.textContent || this.innerText))'>" + k + "</span>\n"
	}
	return origins
}

func handlePrivateMessage(msg *PrivateMessage, gossip *Gossiper) {

	if msg.Destination == gossip.Name {
		fmt.Println("PRIVATE origin " + msg.Origin + " hop-limit " + strconv.FormatInt(int64(msg.HopLimit), 10) + " contents " + msg.Text)
	} else {
		if msg.HopLimit > 0 {
			msg.HopLimit -= 1
			newEncoded, err := protobuf.Encode(&GossipPacket{Private: msg})
			if err != nil {
				println("Gossiper Encode Error: " + err.Error())
			}
			nextHop := routingTable.Table[msg.Destination]
			sendPacket(newEncoded, nextHop, gossip)
		}
	}
}
