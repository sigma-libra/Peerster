package server

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

	gossip.wantMap[gossip.Name] = PeerStatus{
		Identifier: gossip.Name,
		NextID:     msg.ID + 1,
	}
	gossip.orderedMessages[gossip.Name] = append(gossip.orderedMessages[gossip.Name], msg)
	gossip.mu.Unlock()
	newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: &msg})
	printerr("Routing Error", err)
	return newEncoded
}

func handlePrivateMessage(msg *PrivateMessage, gossip *Gossiper) {

	if msg.Destination == gossip.Name {
		if msg.ID == 0 {
			fmt.Println("PRIVATE origin " + msg.Origin + " hop-limit " + strconv.FormatInt(int64(msg.HopLimit), 10) + " contents " + msg.Text)
			messages += msg.Origin + " (private): " + msg.Text + "\n"
		} else {
			gossip.mu.Lock()
			gossip.tclAcks[msg.ID] = append(gossip.tclAcks[msg.ID], msg.Origin)
			if len(gossip.tclAcks[msg.ID]) < ((N/2)+1) {
				gossip.mu.Unlock()
			} else {
				oldTLCMessage := gossip.orderedMessages[gossip.Name][msg.ID-1]
				newTLCMessage := TLCMessage{
					Origin:      oldTLCMessage.Origin,
					ID:          getAndUpdateRumorID(),
					Confirmed:   int(oldTLCMessage.tclMsg.ID),
					TxBlock:     BlockPublish{},
					VectorClock: nil, //TODO correctly initialize vector clock
					Fitness:     0,
				}

				newWrap := RumorableMessage{
					Origin:   newTLCMessage.Origin,
					ID:       newTLCMessage.ID,
					isTLC:    true,
					rumorMsg: nil,
					tclMsg:   newTLCMessage,
				}
				gossip.orderedMessages[gossip.Name] = append(gossip.orderedMessages[gossip.Name], newWrap)

				gossip.mu.Unlock()

				logMsg := "RE-BROADCAST ID "+ strconv.Itoa(newTLCMessage.Confirmed)+
					"WITNESSES ";

				for _, name:= range gossip.tclAcks[msg.ID] {
					logMsg += name + ","
				}

				fmt.Println(logMsg[:len(logMsg) - 1])

				newEncoded, err := protobuf.Encode(&GossipPacket{TLCMessage: &newTLCMessage})
				printerr("Rumor Gossiper Error", err)

				if len(KnownPeers) > 0 {
					randomPeer := Keys[rand.Intn(len(Keys))]
					sendPacket(newEncoded, randomPeer, gossip)
					addToMongering(randomPeer, msg.Origin, msg.ID)

					go statusCountDown(newWrap, randomPeer, gossip)
				}
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
