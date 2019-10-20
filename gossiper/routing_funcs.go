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
	if RTimer > 0 {
		if len(Keys) > 0 {
			randomPeer := Keys[rand.Intn(len(Keys))]
			sendPacket(makeRouteRumor(), randomPeer, gossip)
		}
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
