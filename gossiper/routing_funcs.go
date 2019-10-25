package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
	"math/rand"
	"time"
)

func FireRouteRumor(gossip *Gossiper) {
	fmt.Println(RTimer)
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
		sendPacket(makeRouteRumor(), randomPeer, gossip)
	}
}

func makeRouteRumor() []byte {
	msg := RumorMessage{
		Origin: PeerName,
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
		origins += "<span onclick='openMessageWindow()'> " + k + "</span>\n"
	}
	return origins
}
