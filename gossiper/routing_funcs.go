package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
	"math/rand"
	"strconv"
	"time"
)

func getNextHop(dest string) string {
	routingTable.mu.Lock()
	defer routingTable.mu.Unlock()
	return routingTable.Table[dest]
}

func FireRouteRumor(gossip *Gossiper) {
	if RTimer > 0 {
		ticker := time.NewTicker(time.Duration(RTimer) * time.Second)
		<-ticker.C
		SendRouteRumor(gossip)
		go FireRouteRumor(gossip)

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
	printerr("Routing Error", err)
	return newEncoded
}

func parseRoutingTable() string {
	origins := ""
	routingTable.mu.Lock()
	defer routingTable.mu.Unlock()
	for k, _ := range routingTable.Table {
		origins += "<span onclick='openMessageWindow((this.textContent || this.innerText))'>" + k + "</span>\n"
	}
	return origins
}

func handlePrivateMessage(msg *PrivateMessage, gossip *Gossiper) {

	if msg.Destination == gossip.Name {
		fmt.Println("PRIVATE origin " + msg.Origin + " hop-limit " + strconv.FormatInt(int64(msg.HopLimit), 10) + " contents " + msg.Text)
		messages += msg.Origin + " (private): " + msg.Text + "\n"
	} else {
		if msg.HopLimit > 0 {
			msg.HopLimit -= 1
			newEncoded, err := protobuf.Encode(&GossipPacket{Private: msg})
			printerr("Routing Error", err)
			nextHop := getNextHop(msg.Destination)
			sendPacket(newEncoded, nextHop, gossip)
		}
	}
}
