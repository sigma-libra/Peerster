package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
	"math/rand"
	"strconv"
	"time"
)

var wantMap = make(map[string]PeerStatus)
var earlyMessages = make(map[string][]RumorMessage)
var orderedMessages = make(map[string][]RumorMessage)

var rumorTracker = make(map[string]RumorMessage)

var waitingForReply = make(map[string]bool)

func HandleRumorMessagesFrom(gossip *Gossiper) {

	for {

		pkt, sender := getAndDecodePacket(gossip)

		AddPeer(sender)
		fmt.Println("PEERS " + FormatPeers(Keys))

		isRumorPkt := isRumorPacket(pkt)

		if isRumorPkt {
			msg := pkt.Rumor

			messages += msg.Origin + ": " + msg.Text + "\n"

			printMsg := "RUMOR origin " + msg.Origin + " from " + sender + " ID " + strconv.FormatUint(uint64(msg.ID), 10) + " contents " + msg.Text
			fmt.Println(printMsg)

			_, wantsKnownForSender := wantMap[msg.Origin]
			if !wantsKnownForSender {
				wantMap[msg.Origin] = PeerStatus{
					Identifier: msg.Origin,
					NextID:     1,
				}
			}

			receivedBefore := wantMap[msg.Origin].NextID > msg.ID

			if !receivedBefore {
				newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: msg})
				if err != nil {
					println("Gossiper Encode Error: " + err.Error())
				}

				if len(KnownPeers) > 0 {
					randomPeer := Keys[rand.Intn(len(Keys))]
					sendPacket(newEncoded, randomPeer, gossip)
					rumorTracker[randomPeer] = *msg

					fmt.Println("MONGERING with " + randomPeer)

					waitingForReply[randomPeer] = true

					go statusCountDown(*msg, randomPeer, gossip)
				}

				if wantMap[msg.Origin].NextID == msg.ID {
					wantMap[msg.Origin] = PeerStatus{
						Identifier: msg.Origin,
						NextID:     msg.ID + 1,
					}
					_, listExists := orderedMessages[msg.Origin]
					if !listExists {
						orderedMessages[msg.Origin] = make([]RumorMessage, 0)
					}

					_, listExists = earlyMessages[msg.Origin]
					if !listExists {
						earlyMessages[msg.Origin] = make([]RumorMessage, 0)
					}

					orderedMessages[msg.Origin] = append(orderedMessages[msg.Origin], *msg)
					fastForward(msg.Origin)
				} else {
					_, listExists := earlyMessages[msg.Origin]
					if !listExists {
						earlyMessages[msg.Origin] = make([]RumorMessage, 0)
					}
					earlyMessages[msg.Origin] = append(earlyMessages[msg.Origin], *msg)
				}

			}

		} else {
			// packet is status
			msg := pkt.Status
			waitingForReply[sender] = false

			for _, wanted := range msg.Want {
				_, wantsKnownForWanted := wantMap[wanted.Identifier]
				if !wantsKnownForWanted {
					wantMap[wanted.Identifier] = PeerStatus{
						Identifier: wanted.Identifier,
						NextID:     1,
					}
				}

				_, listExists := orderedMessages[wanted.Identifier]
				if !listExists {
					orderedMessages[wanted.Identifier] = make([]RumorMessage, 0)
				}

				_, listExists = earlyMessages[wanted.Identifier]
				if !listExists {
					earlyMessages[wanted.Identifier] = make([]RumorMessage, 0)
				}
			}

			printMsg := "STATUS from " + sender

			for _, wanted := range msg.Want {
				printMsg += " peer " + wanted.Identifier + " nextID " + strconv.FormatUint(uint64(wanted.NextID), 10)
			}
			fmt.Println(printMsg)
			upToDate := true
			for _, wanted := range msg.Want {

				if wantMap[wanted.Identifier].NextID > wanted.NextID {
					//I have more messages
					msgToSend := orderedMessages[wanted.Identifier][wanted.NextID-1]
					rumorEncoded, err := protobuf.Encode(&GossipPacket{Rumor: &msgToSend})
					if err != nil {
						println("Gossiper Encode Error: " + err.Error())
					}
					sendPacket(rumorEncoded, sender, gossip)
					//break //only send one

				} else if wantMap[wanted.Identifier].NextID < wanted.NextID {
					//I have fewer messages
					upToDate = false

					sendPacket(makeStatusPacket(), sender, gossip)
					//break //only send one

				} else {
					//coin flip - nothing new between me and original dst
					originalMessage, _ := rumorTracker[sender]

					coin := rand.Int() % 2
					heads := (coin == 1)
					if heads && len(KnownPeers) > 0 {
						randomPeer := Keys[rand.Intn(len(Keys))]
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

func HandleClientRumorMessages(gossip *Gossiper, name string, peerGossiper *Gossiper) {

	localID := 1

	for {

		pkt := getAndDecodeFromClient(gossip)
		text := pkt.Text

		fmt.Println("CLIENT MESSAGE " + text)

		fmt.Println("PEERS " + FormatPeers(Keys))

		msg := RumorMessage{
			Origin: name,
			ID:     uint32(localID),
			Text:   text,
		}

		localID += 1

		messages += msg.Origin + ": " + msg.Text + "\n"

		printMsg := "RUMOR origin " + msg.Origin + " from " + peerGossiper.address.String() + " ID " + strconv.FormatUint(uint64(msg.ID), 10) + " contents " + msg.Text
		fmt.Println(printMsg)

		newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: &msg})
		if err != nil {
			println("Gossiper Encode Error: " + err.Error())
		}

		if len(KnownPeers) > 0 {
			randomPeer := Keys[rand.Intn(len(Keys))]
			sendPacket(newEncoded, randomPeer, peerGossiper)
			rumorTracker[randomPeer] = msg

			fmt.Println("MONGERING with " + randomPeer)

			go statusCountDown(msg, randomPeer, peerGossiper)
		}
	}

}

func makeStatusPacket() []byte {
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

func FireAntiEntropy(gossip *Gossiper) {
	for {
		ticker := time.NewTicker(time.Duration(AntiEntropy) * time.Second)
		<-ticker.C
		if len(Keys) > 0 {
			randomPeer := Keys[rand.Intn(len(Keys))]
			sendPacket(makeStatusPacket(), randomPeer, gossip)
		}

	}
}

func fastForward(origin string) {
	currentNext := wantMap[origin].NextID
	updated := false
	indexesDelivered := make([]int, 0)
	for {
		for i, savedMsg := range earlyMessages[origin] {
			if savedMsg.ID == currentNext {
				currentNext += 1
				updated = true
				indexesDelivered = append(indexesDelivered, i)
				orderedMessages[origin] = append(orderedMessages[origin], savedMsg)
			}

		}
		if !updated {
			break
		} else {
			updated = false
		}
	}
	for _, index := range indexesDelivered {
		earlyMessages[origin] = RemoveIndex(earlyMessages[origin], index)
	}
	wantMap[origin] = PeerStatus{
		origin,
		currentNext,
	}

}

func RemoveIndex(s []RumorMessage, index int) []RumorMessage {
	return append(s[:index], s[index+1:]...)
}

func statusCountDown(msg RumorMessage, dst string, gossip *Gossiper) {

	ticker := time.NewTicker(10 * time.Second)
	<-ticker.C
	if waitingForReply[dst] == true && len(Keys) > 0 {
		encoded, err := protobuf.Encode(&GossipPacket{Rumor: &msg})
		if err != nil {
			println("Gossiper Encode Error: " + err.Error())
		}
		randomPeer := Keys[rand.Intn(len(Keys))]
		sendPacket(encoded, randomPeer, gossip)

		fmt.Println("MONGERING with " + randomPeer)

		waitingForReply[randomPeer] = true

		go statusCountDown(msg, randomPeer, gossip)

	}
	return
}
