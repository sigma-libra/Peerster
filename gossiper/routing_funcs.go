package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
	"math/rand"
	"strconv"
	"time"
)

func getNextHop(dest string) string {
	routingTable.mu.RLock()
	defer routingTable.mu.RUnlock()
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
	gossip.mu.Lock()
	msg := RumorMessage{
		Origin: gossip.Name,
		ID:     getAndUpdateRumorID(),
		Text:   "",
	}

	groupArray := make([]string, 0)
	for gr, here := range groups {
		if here {
			groupArray = append(groupArray, gr)
		}
	}

	gossip.wantMap[gossip.Name] = PeerStatus{
		Identifier: gossip.Name,
		NextID:     msg.ID + 1,
		Groups:     groupArray,
	}
	gossip.orderedMessages[gossip.Name] = append(gossip.orderedMessages[gossip.Name], msg)
	for _, grp := range msg.Groups {
		gossip.groupMessages[grp] = append(gossip.groupMessages[grp], msg)
	}
	gossip.mu.Unlock()
	newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: &msg})
	printerr("Routing Error", err)
	return newEncoded
}

func parseRoutingTable() string {
	origins := ""
	routingTable.mu.RLock()
	defer routingTable.mu.RUnlock()
	for k, _ := range routingTable.Table {
		origins += "<span onclick='openMessageWindow((this.textContent || this.innerText))'>" + k + "</span>\n"
	}
	return origins
}

func handlePrivateMessage(msg *PrivateMessage, gossip *Gossiper) {

	if msg.Destination == gossip.Name {
		fmt.Println("PRIVATE origin " + msg.Origin + " hop-limit " + strconv.FormatInt(int64(msg.HopLimit), 10) + " contents " + msg.Text)
		msgGroups := [2]string{"all", msg.Origin}
		for _, gr := range msgGroups {
			_, known := messages[gr]
			if !known {
				messages[gr] = msg.Origin + " (private): " + msg.Text + "\n"
			} else {
				messages[gr] += msg.Origin + " (private): " + msg.Text + "\n"
			}
		}
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
