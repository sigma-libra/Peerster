package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
	"math"
	"math/rand"
	"strconv"
	"time"
)

var wantMap = make(map[string]PeerStatus)
var earlyMessages = make(map[string]map[uint32]RumorMessage)
var orderedMessages = make(map[string][]RumorMessage)

// map (ip we monger to) -> (origin of mongered message) -> (ids of mongered messages from origin)
var mongeringMessages = make(map[string]map[string][]uint32)

var routingTable = InitRoutingTable()

func HandleRumorMessagesFrom(gossip *Gossiper) {

	SendRouteRumor(gossip)
	go FireRouteRumor(gossip)

	for {

		pkt, sender := getAndDecodePacket(gossip)

		AddPeer(sender)
		fmt.Println("PEERS " + FormatPeers(Keys))

		if pkt.Rumor != nil {
			msg := pkt.Rumor

			initNode(msg.Origin)

			//update routing table
			prevSender, prevExists := routingTable.Table[msg.Origin]
			routingTable.Table[msg.Origin] = sender

			if (!prevExists || (prevExists && prevSender != sender)) && msg.Text == "" {
				fmt.Println("DSDV " + msg.Origin + " " + sender)
			}

			if msg.Text != "" {
				printMsg := "RUMOR origin " + msg.Origin + " from " + sender + " ID " + strconv.FormatUint(uint64(msg.ID), 10) + " contents " + msg.Text
				fmt.Println(printMsg)
			}

			receivedBefore := (wantMap[msg.Origin].NextID > msg.ID) || (msg.Origin == PeerName)

			//message not received before: start mongering
			if !receivedBefore {

				messages += msg.Origin + ": " + msg.Text + "\n"

				//pick random peer to send to
				randomPeer := Keys[rand.Intn(len(Keys))]

				newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: msg})
				if err != nil {
					println("Gossiper Encode Error: " + err.Error())
				}
				sendPacket(newEncoded, randomPeer, gossip)

				fmt.Println("MONGERING with " + randomPeer)
				addToMongering(randomPeer, msg.Origin, msg.ID)

				//start countdown to monger to someone else
				go statusCountDown(*msg, randomPeer, gossip)

				//if next message we want, save in vector clock
				if wantMap[msg.Origin].NextID == msg.ID {
					wantMap[msg.Origin] = PeerStatus{
						Identifier: msg.Origin,
						NextID:     msg.ID + 1,
					}
					orderedMessages[msg.Origin] = append(orderedMessages[msg.Origin], *msg)
					fastForward(msg.Origin)
				} else {
					// too early: save for later
					earlyMessages[msg.Origin][(*msg).ID] = *msg
				}

				//ack reception of packet
				sendPacket(makeStatusPacket(), sender, gossip)

			}

		} else if pkt.Status != nil {
			// packet is status
			msg := pkt.Status

			printMsg := "STATUS from " + sender

			//update our info about nodes
			for _, wanted := range msg.Want {
				initNode(wanted.Identifier)
				printMsg += " peer " + wanted.Identifier + " nextID " + strconv.FormatUint(uint64(wanted.NextID), 10)

			}
			fmt.Println(printMsg)
			//check if this is an ack

			sendMessage := false
			smallestOrigin := ""
			smallestIDMissing := uint32(math.MaxUint32)

			sendStatus := false

			//
			isAck := false
			originAcked := ""
			idAcked := uint32(math.MaxUint32)

			for _, wanted := range msg.Want {

				if wantMap[wanted.Identifier].NextID > wanted.NextID {
					//(1) The sender has ​ other new messages that the receiver peer has not yet seen,
					// and if so repeats the rumormongering process by sending ​ one of those messages ​ to the same receiving peer​ .
					//I have more messages - find the earliest one
					sendMessage = true

					if wanted.NextID < smallestIDMissing {
						smallestOrigin = wanted.Identifier
						smallestIDMissing = wanted.NextID
					}

				} else if wantMap[wanted.Identifier].NextID < wanted.NextID {
					//(2) The sending peer does not have anything new but sees from the exchanged
					//status that the ​ receiver peer has new messages. Then the sending peer itself
					//sends a  StatusPacket containing its status vector, which causes the
					//receiver peer to send send the missing messages back ​ (one at a time); ​ (
					//I have fewer messages - send status
					sendStatus = true

				}

				//check if this acks any mongering
				mongeredIds, wasMongering := mongeringMessages[sender][wanted.Identifier]
				if wasMongering {
					ongoingMongers := make([]uint32, 0)
					for _, mongeredId := range mongeredIds {
						if mongeredId < wanted.NextID {
							isAck = true
							if idAcked > mongeredId {
								originAcked = wanted.Identifier
								idAcked = mongeredId
							}
						} else {
							ongoingMongers = append(ongoingMongers, mongeredId)
						}

					}
					mongeringMessages[sender][wanted.Identifier] = ongoingMongers
				}
			}

			if sendMessage {
				msgToSend := orderedMessages[smallestOrigin][smallestIDMissing-1]
				rumorEncoded, err := protobuf.Encode(&GossipPacket{Rumor: &msgToSend})
				if err != nil {
					println("Gossiper Encode Error: " + err.Error())
				}

				sendPacket(rumorEncoded, sender, gossip)
				addToMongering(sender, msgToSend.Origin, msgToSend.ID)
				fmt.Println("MONGERING with " + sender)
				go statusCountDown(msgToSend, sender, gossip)

			} else if sendStatus {
				sendPacket(makeStatusPacket(), sender, gossip)
			} else {
				fmt.Println("IN SYNC WITH " + sender)

				//(3) If neither peer has new messages, the sending peer (rumormongering) peer S
				//flips a coin (e.g., ​ rand.Int() % 2​ ), and either (heads) picks ​ a new
				//random peer to send the rumor message to​ , or (tails) ceases the
				//rumormongering process.
				if isAck {
					coin := rand.Int() % 2
					heads := coin == 1
					if heads && len(KnownPeers) > 0 {
						originalMessage := getMessage(originAcked, idAcked)
						randomPeer := Keys[rand.Intn(len(Keys))]
						newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: &originalMessage})
						if err != nil {
							println("Gossiper Encode Error: " + err.Error())
						}
						sendPacket(newEncoded, randomPeer, gossip)
						addToMongering(randomPeer, originalMessage.Origin, originalMessage.ID)
						//fmt.Println("MONGERING with " + randomPeer)
						fmt.Println("FLIPPED COIN sending rumor to " + randomPeer)
						go statusCountDown(originalMessage, randomPeer, gossip)
					}
				}
			}
		}

	}
}

func HandleClientRumorMessages(gossip *Gossiper, name string, peerGossiper *Gossiper) {

	for {

		pkt := getAndDecodeFromClient(gossip)
		text := pkt.Text

		fmt.Println("CLIENT MESSAGE " + text)

		fmt.Println("PEERS " + FormatPeers(Keys))

		msg := RumorMessage{
			Origin: name,
			ID:     getAndUpdateRumorID(),
			Text:   text,
		}

		messages += msg.Origin + ": " + msg.Text + "\n"

		newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: &msg})
		if err != nil {
			println("Gossiper Encode Error: " + err.Error())
		}

		if len(KnownPeers) > 0 {
			randomPeer := Keys[rand.Intn(len(Keys))]
			sendPacket(newEncoded, randomPeer, peerGossiper)
			addToMongering(randomPeer, msg.Origin, msg.ID)

			fmt.Println("MONGERING with " + randomPeer)

			go statusCountDown(msg, randomPeer, peerGossiper)
		}
	}

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

func statusCountDown(msg RumorMessage, dst string, gossip *Gossiper) {

	ticker := time.NewTicker(10 * time.Second)
	<-ticker.C

	mongeredIDs, dstTracked := mongeringMessages[dst][msg.Origin]

	if dstTracked {
		stillMongering := false
		for _, mongeredID := range mongeredIDs {
			if mongeredID == msg.ID {
				stillMongering = true
				break
			}
		}

		if stillMongering && len(Keys) > 0 {
			encoded, err := protobuf.Encode(&GossipPacket{Rumor: &msg})
			if err != nil {
				println("Gossiper Encode Error: " + err.Error())
			}
			randomPeer := Keys[rand.Intn(len(Keys))]
			sendPacket(encoded, randomPeer, gossip)
			addToMongering(randomPeer, msg.Origin, msg.ID)

			fmt.Println("MONGERING with " + randomPeer)

			go statusCountDown(msg, randomPeer, gossip)

		}
	}

}
