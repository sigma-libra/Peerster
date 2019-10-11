package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
	"math/rand"
	"strconv"
	"time"
)

func HandleRumorMessagesFrom(gossip *Gossiper, name string, isClient bool, wantUpdateChan chan PeerStatus) {

	wantMap := InitWantMap()
	earlyMessages := make(map[string][]RumorMessage)
	orderedMessages := make(map[string][]RumorMessage)
	rumorTracker := make(map[string]RumorMessage)

	waitingForReply := make(map[string]chan bool)

	updateWaiting := make(chan GossipWaiter)

	for {
		select {
		case gw := <-updateWaiting:
			waitingForReply[gw.dst] = gw.done
			rumorTracker[gw.dst] = gw.msg

		default:
		}


		pkt, sender := getAndDecodePacket(gossip)

		if pkt.Simple != nil {

		}

		isRumorPkt := isRumorPacket(pkt)

		if isRumorPkt {
			msg := pkt.Rumor

			messages += msg.Origin + ": " + msg.Text

			if isClient {
				msg.Origin = name
				printMsg := "CLIENT MESSAGE " + msg.Text
				fmt.Println(printMsg)
			} else {
				AddPeer(sender)
				printMsg := "RUMOR origin " + msg.Origin + " from " + sender + " ID " + strconv.FormatUint(uint64(msg.ID), 10) + " contents " + msg.Text
				fmt.Println(printMsg)
			}

			fmt.Println("PEERS " + FormatPeers(KnownPeers))

			receivedBefore := wantMap[msg.Origin].NextID > msg.ID

			if !receivedBefore {
				newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: msg})
				if err != nil {
					println("Gossiper Encode Error: " + err.Error())
				}

				randomPeer := KnownPeers[rand.Intn(len(KnownPeers))]
				sendPacket(newEncoded, randomPeer, gossip)
				rumorTracker[randomPeer] = *msg

				fmt.Println("MONGERING with " + randomPeer)

				ticker := time.NewTicker(10 * time.Second)
				done := make(chan bool)
				waitingForReply[randomPeer] = done

				go statusCountDown(ticker, done, *msg, gossip, updateWaiting)

				if wantMap[msg.Origin].NextID == msg.ID {
					wantMap[msg.Origin] = PeerStatus{
						Identifier: msg.Origin,
						NextID:     msg.ID + 1,
					}
					orderedMessages[msg.Origin] = append(orderedMessages[msg.Origin], *msg)
					newNext, newUndelivered, newDelivered := fastForward(wantMap[msg.Origin].NextID, earlyMessages[msg.Origin], orderedMessages[msg.Origin])
					newPeerStatus := PeerStatus{
						Identifier: msg.Origin,
						NextID:     newNext,
					}
					wantMap[msg.Origin] = newPeerStatus
					wantUpdateChan <- newPeerStatus

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

			printMsg := "STATUS from " + sender

			for _, wanted := range msg.Want {
				printMsg += " peer " + wanted.Identifier + " nextID " + string(wanted.NextID)
			}
			fmt.Println(printMsg)
			upToDate := true
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
					upToDate = false

					sendPacket(makeStatusPacket(wantMap), sender, gossip)
					break //only send one

				} else {
					//coin flip - nothing new between me and original dst
					originalMessage := rumorTracker[sender]
					coin := rand.Int() % 2
					heads := coin == 1
					if heads {
						randomPeer := KnownPeers[rand.Intn(len(KnownPeers))]
						newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: &originalMessage})
						if err != nil {
							println("Gossiper Encode Error: " + err.Error())
						}
						sendPacket(newEncoded, randomPeer, gossip)
						fmt.Println("FLIPPED COIN sending rumor to " + randomPeer)
						break
					}

				}
			}
			if upToDate {
				fmt.Println("IN SYNC WITH " + sender)
			}
		}
	}
}

func InitWantMap() map[string]PeerStatus {
	wantMap := make(map[string]PeerStatus)
	for _, peer := range KnownPeers {
		wantMap[peer] = PeerStatus{
			Identifier: peer,
			NextID:     1,
		}
	}
	return wantMap

}

func makeStatusPacket(wantMap map[string]PeerStatus) []byte {
	wants := make([]PeerStatus, 0)
	for _, status := range wantMap {
		wants = append(wants, status)
	}
	wantPacket := StatusPacket{Want: wants}
	newEncoded, err := protobuf.Encode(&GossipPacket{Status: &wantPacket})
	if err != nil {
		println("Gossiper Encode Error: " + err.Error())
	}
	return newEncoded

}

func FireAntiEntropy(wantUpdateChan chan PeerStatus, gossip *Gossiper) {
	wantMap := InitWantMap()
	for {
		ticker := time.NewTicker(time.Duration(AntiEntropy) * time.Second)
		<-ticker.C
		select {
		case wantsUpdate := <-wantUpdateChan:
			wantMap[wantsUpdate.Identifier] = wantsUpdate
		default:
		}

		randomPeer := KnownPeers[rand.Intn(len(KnownPeers))]

		sendPacket(makeStatusPacket(wantMap), randomPeer, gossip)

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

func statusCountDown(ticker *time.Ticker, messageReceived chan bool, msg RumorMessage, gossip *Gossiper, updateWaiting chan GossipWaiter) {

	encoded, err := protobuf.Encode(&GossipPacket{Rumor: &msg})
	if err != nil {
		println("Gossiper Encode Error: " + err.Error())
	}

	for {
		select {
		case <-ticker.C:
			randomPeer := KnownPeers[rand.Intn(len(KnownPeers))]
			sendPacket(encoded, randomPeer, gossip)

			ticker := time.NewTicker(10 * time.Second)
			done := make(chan bool)

			updateWaiting <- GossipWaiter{
				dst:  randomPeer,
				done: done,
				msg:  msg,
			}

			go statusCountDown(ticker, done, msg,gossip, updateWaiting)
			return

		case <-messageReceived:
			ticker.Stop()
			return
		}
	}
}
