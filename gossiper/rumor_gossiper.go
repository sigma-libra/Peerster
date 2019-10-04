package gossiper

import (
	"github.com/dedis/protobuf"
	"math/rand"
	"time"
)

func HandleRumorMessagesFrom(gossip *Gossiper, name string, gossipAddr string, knownPeers []string, isClient bool) {

	wantMap := make(map[string]PeerStatus)
	orderedMessages := make(map[string][]RumorMessage)
	earlyMessages := make(map[string][]RumorMessage)

	rumorTracker := make(map[string]RumorMessage)

	waitingForReply := make(map[string]chan bool)

	updateWaiting := make(chan GossipWaiter)

	for {
		select {
		case gw := <-updateWaiting:
			waitingForReply[gw.dst] = gw.done
			rumorTracker[gw.dst] = gw.msg

		default:

			pkt, sender := getAndDecodePacket(gossip)

			isRumorPkt := isRumorPacket(pkt)

			if isRumorPkt {
				msg := pkt.Rumor

				if isClient {
					msg.Origin = name
				}

				receivedBefore := wantMap[msg.Origin].NextID > msg.ID

				if !receivedBefore {
					newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: msg})
					if err != nil {
						println("Gossiper Encode Error: " + err.Error())
					}

					randomPeer := knownPeers[rand.Intn(len(knownPeers))]
					sendPacket(newEncoded, randomPeer, gossip)
					rumorTracker[randomPeer] = *msg

					ticker := time.NewTicker(10 * time.Second)
					done := make(chan bool)
					waitingForReply[randomPeer] = done

					go statusCountDown(ticker, done, *msg, knownPeers, gossip, updateWaiting)

					if wantMap[msg.Origin].NextID == msg.ID {
						wantMap[msg.Origin] = PeerStatus{
							Identifier: msg.Origin,
							NextID:     msg.ID + 1,
						}
						orderedMessages[msg.Origin] = append(orderedMessages[msg.Origin], *msg)
						newNext, newUndelivered, newDelivered := fastForward(wantMap[msg.Origin].NextID, earlyMessages[msg.Origin], orderedMessages[msg.Origin])
						wantMap[msg.Origin] = PeerStatus{
							Identifier: msg.Origin,
							NextID:     newNext,
						}
						orderedMessages[msg.Origin] = newDelivered
						earlyMessages[msg.Origin] = newUndelivered
					} else {
						earlyMessages[msg.Origin] = append(earlyMessages[msg.Origin], *msg)

					}

				}

			} else {
				msg := pkt.Status
				if done, ok := waitingForReply[sender]; ok {
					done <- true
				}
				for _, wanted := range msg.Want {
					if wantMap[wanted.Identifier].NextID > wanted.NextID {
						//I have more messages
						msgToSend := orderedMessages[wanted.Identifier][wanted.NextID]
						rumorEncoded, err := protobuf.Encode(&GossipPacket{Rumor: &msgToSend})
						if err != nil {
							println("Gossiper Encode Error: " + err.Error())
						}
						sendPacket(rumorEncoded, sender, gossip)
						break //only send one

					} else if wantMap[wanted.Identifier].NextID < wanted.NextID {
						//I have fewer messages
						wants := make([]PeerStatus, 0)
						for _, ps := range wantMap {
							wants = append(wants, ps)
						}
						wantPacket := StatusPacket{Want: wants}
						statusEncoded, err := protobuf.Encode(&GossipPacket{Status: &wantPacket})
						if err != nil {
							println("Gossiper Encode Error: " + err.Error())
						}
						sendPacket(statusEncoded, sender, gossip)
						break //only send one

					} else {
						//coin flip - nothing new between me and original dst
						originalMessage := rumorTracker[sender]
						coin := rand.Int() % 2
						heads := coin == 1
						if heads {
							randomPeer := knownPeers[rand.Intn(len(knownPeers))]
							newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: &originalMessage})
							if err != nil {
								println("Gossiper Encode Error: " + err.Error())
							}
							sendPacket(newEncoded, randomPeer, gossip)
							break
						}

					}
				}
			}
		}
	}
}

func fastForward(currentNext uint32, undelivered []RumorMessage, ordered []RumorMessage) (uint32, []RumorMessage, []RumorMessage) {
	updated := false
	indexesDelivered := make([]int, 0)
	for {
		for i, savedMsg := range undelivered {
			if savedMsg.ID == currentNext {
				currentNext += 1
				updated = true
				indexesDelivered = append(indexesDelivered, i)
				ordered = append(ordered, savedMsg)
			}

		}
		if !updated {
			break
		} else {
			updated = false
		}
	}
	for _, index := range indexesDelivered {
		undelivered = RemoveIndex(undelivered, index)
	}
	return currentNext, undelivered, ordered

}

func RemoveIndex(s []RumorMessage, index int) []RumorMessage {
	return append(s[:index], s[index+1:]...)
}

func statusCountDown(ticker *time.Ticker, messageReceived chan bool, msg RumorMessage,
	knownPeers []string, gossip *Gossiper, updateWaiting chan GossipWaiter) {

	encoded, err := protobuf.Encode(&GossipPacket{Rumor: &msg})
	if err != nil {
		println("Gossiper Encode Error: " + err.Error())
	}

	for {
		select {
		case <-ticker.C:
			randomPeer := knownPeers[rand.Intn(len(knownPeers))]
			sendPacket(encoded, randomPeer, gossip)

			ticker := time.NewTicker(10 * time.Second)
			done := make(chan bool)

			updateWaiting <- GossipWaiter{
				dst:  randomPeer,
				done: done,
				msg:  msg,
			}

			go statusCountDown(ticker, done, msg, knownPeers, gossip, updateWaiting)
			return

		case <-messageReceived:
			ticker.Stop()
			return
		}
	}
}
