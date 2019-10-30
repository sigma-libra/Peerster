package gossiper

import (
	"github.com/dedis/protobuf"
	"math/rand"
	"time"
)

func FireRouteRumor(gossip *Gossiper) {
	if RTimer > 0 {
		for {
			ticker := time.NewTicker(time.Duration(RTimer) * time.Second)
			<-ticker.C
			SendRouteRumor(gossip)
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
	for k, _:= range routingTable.Table {
		origins += "<span onclick='openMessageWindow((this.textContent || this.innerText))'>" + k +"</span>\n"
	}
	return origins
}
