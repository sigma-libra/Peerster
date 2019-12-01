package server

import (
	"encoding/hex"
	"fmt"
	"github.com/dedis/protobuf"
	"math"
	"math/rand"
	"strconv"
)

func sendAck(msg TLCMessage, gossiper *Gossiper) {
	fmt.Println("SENDING ACK origin " + msg.Origin + " ID " + strconv.FormatUint(uint64(msg.ID), 10))
	//send ack
	ack := PrivateMessage{
		Origin:      gossiper.Name,
		ID:          msg.ID,
		Text:        "",
		Destination: msg.Origin,
		HopLimit:    Hoplimit,
	}

	newPacketBytes, err := protobuf.Encode(&GossipPacket{Private: &ack})
	printerr("TLC Message ack protobuf encoding", err)

	nextHop := getNextHop(ack.Destination)
	sendPacket(newPacketBytes, nextHop, gossiper)
}

func handleRumorableMessage(msg *RumorableMessage, sender string, gossip *Gossiper) bool {

	//TODO TEST THIS
	//if msg.Origin != gossip.Name {
	initNode(msg.Origin, gossip)

	//update routing table
	routingTable.mu.Lock()
	lastID, lastIDExists := routingTable.LastMsgID[msg.Origin]

	if (!lastIDExists || lastID < msg.ID) && msg.Origin != gossip.Name {
		routingTable.Table[msg.Origin] = sender
		if !lastIDExists {
			addToMessageableNodes(msg.Origin)
		}
		routingTable.LastMsgID[msg.Origin] = msg.ID

		if (!msg.isTLC) && msg.rumorMsg.Text != "" {
			fmt.Println("DSDV " + msg.Origin + " " + sender)
		}
	}
	routingTable.mu.Unlock()

	gossip.mu.Lock()
	defer gossip.mu.Unlock()
	receivedBefore := (gossip.wantMap[msg.Origin].NextID > msg.ID) //|| (msg.Origin == gossip.Name)

	//message not received before: start mongering
	if !receivedBefore {

		//TODO check validiticy of filenames

		if !msg.isTLC {
			if msg.rumorMsg.Origin != gossip.Name {
				printMsg := "RUMOR origin " + msg.Origin + " from " + sender + " ID " + strconv.FormatUint(uint64(msg.ID), 10) + " contents " + msg.rumorMsg.Text
				fmt.Println(printMsg)
			}

			if (msg.rumorMsg.Text != "") && (msg.Origin != gossip.Name) {
				messages += msg.Origin + ": " + msg.rumorMsg.Text + "\n"
			}
		} else {

			if msg.tclMsg.Confirmed == -1 {
				fmt.Println("UNCONFIRMED GOSSIP origin " + msg.Origin + " ID " + strconv.FormatUint(uint64(msg.ID), 10) +
					" file name " + msg.tclMsg.TxBlock.Transaction.Name + " size " + strconv.FormatInt(msg.tclMsg.TxBlock.Transaction.Size, 10) +
					" metahash " + hex.EncodeToString(msg.tclMsg.TxBlock.Transaction.MetafileHash))

				if Simple_File_Share || (Round_based_TLC && gossip.roundTracker[msg.Origin] >= gossip.my_time) || AckAll {
					sendAck(*msg.tclMsg, gossip)
				}

				_, trackingOrigin := gossip.roundTracker[msg.Origin]
				if !trackingOrigin {
					gossip.roundTracker[msg.Origin] = 1
				} else {
					gossip.roundTracker[msg.Origin] += 1
				}
			}

			if msg.tclMsg.Confirmed > -1 {
				str1 := "CONFIRMED GOSSIP origin " + msg.Origin + " ID " + strconv.FormatUint(uint64(msg.tclMsg.Confirmed), 10) +
					" file name " + msg.tclMsg.TxBlock.Transaction.Name + " size " + strconv.FormatInt(msg.tclMsg.TxBlock.Transaction.Size, 10)
				str2 :=" metahash " + hex.EncodeToString(msg.tclMsg.TxBlock.Transaction.MetafileHash)
				fmt.Println(str1 + str2)
				confirmed += str1 + "\n" + str2 + "\n"
			}

		}



		//pick random peer to send to
		randomPeer := Keys[rand.Intn(len(Keys))]

		var newEncoded []byte
		var err error
		if msg.isTLC {
			newEncoded, err = protobuf.Encode(&GossipPacket{TLCMessage: msg.tclMsg})
		} else {
			newEncoded, err = protobuf.Encode(&GossipPacket{Rumor: msg.rumorMsg})
		}
		printerr("Gossiper Encode Error", err)
		sendPacket(newEncoded, randomPeer, gossip)

		//fmt.Println("MONGERING with " + randomPeer)
		addToMongering(randomPeer, msg.Origin, msg.ID)

		//if next message we want, save in vector clock
		if gossip.wantMap[msg.Origin].NextID == msg.ID {

			//check if we can go to next round
			if msg.isTLC {
				if _, tracked := gossip.roundTracker[msg.Origin]; !tracked {
					gossip.roundTracker[msg.Origin] = 0
				}
				gossip.roundTracker[msg.Origin]+=1

				if len(gossip.tlcBuffer) > 0 && gossip.tlcSentForCurrentTime {
					nbAboveRound := 0
					for _, round := range gossip.roundTracker {
						if round >= gossip.my_time {
							nbAboveRound += 1
						}
					}
					if nbAboveRound >= (N/2) + 1 || gossip.my_time == 0 {
						gossip.my_time += 1
						str := "ADVANCING TO round "+ strconv.Itoa(gossip.my_time) +" BASED ON CONFIRMED MESSAGES"
						rounds += str + "\n"
						for i, peer := range gossip.tclAcks[msg.ID] {
							id := strconv.Itoa(i+1)
							node := "origin" + id + " " + peer + " ID" + id + " " + strconv.FormatUint(uint64(msg.ID), 10) + " "
							str += node
							rounds += node + "\n"
						}
						fmt.Println(str)
							//origin1 <origin> ID1 <ID>, origin2 <origin> ID2 <ID>")
						if len(gossip.tlcBuffer) > 0 {
							nextBlock := gossip.tlcBuffer[0]
							gossip.tlcBuffer = gossip.tlcBuffer[1:]
							HandleBlock(nextBlock, *gossip)
						} else {
							gossip.tlcSentForCurrentTime = false
						}


					}
				}
			}
			gossip.wantMap[msg.Origin] = PeerStatus{
				Identifier: msg.Origin,
				NextID:     msg.ID + 1,
			}
			gossip.orderedMessages[msg.Origin] = append(gossip.orderedMessages[msg.Origin], *msg)
			fastForward(msg.Origin, gossip)
		} else {
			// too early: save for later
			gossip.earlyMessages[msg.Origin][(*msg).ID] = *msg
		}

		//start countdown to monger to someone else
		go statusCountDown(*msg, randomPeer, gossip)

	}

	//ack reception of packet
	sendPacket(makeStatusPacket(gossip), sender, gossip)
	//}
	return receivedBefore
}

func handleStatusMessage(msg *StatusPacket, sender string, gossip *Gossiper) {
	printMsg := "STATUS from " + sender

	//update our info about nodes
	for _, wanted := range msg.Want {
		initNode(wanted.Identifier, gossip)
		printMsg += " peer " + wanted.Identifier + " nextID " + strconv.FormatUint(uint64(wanted.NextID), 10)

	}
	//fmt.Println(printMsg)
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

		gossip.mu.Lock()
		if gossip.wantMap[wanted.Identifier].NextID > wanted.NextID {
			//(1) The sender has ​ other new messages that the receiver peer has not yet seen,
			// and if so repeats the rumormongering process by sending ​ one of those messages ​ to the same receiving peer​ .
			//I have more messages - find the earliest one
			sendMessage = true

			if wanted.NextID < smallestIDMissing {
				smallestOrigin = wanted.Identifier
				smallestIDMissing = wanted.NextID
			}

		} else if gossip.wantMap[wanted.Identifier].NextID < wanted.NextID {
			//(2) The sending peer does not have anything new but sees from the exchanged
			//status that the ​ receiver peer has new messages. Then the sending peer itself
			//sends a  StatusPacket containing its status vector, which causes the
			//receiver peer to send send the missing messages back ​ (one at a time); ​ (
			//I have fewer messages - send status
			sendStatus = true

		}
		gossip.mu.Unlock()

		//check if this acks any mongering
		mongerer.mu.Lock()
		mongeredIds, wasMongering := mongerer.mongeringMessages[sender][wanted.Identifier]
		ongoingMongers := make([]uint32, 0)
		if wasMongering {
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
		}
		mongerer.mongeringMessages[sender][wanted.Identifier] = ongoingMongers
		mongerer.mu.Unlock()
	}

	if sendMessage {
		gossip.mu.Lock()
		msgToSend := gossip.orderedMessages[smallestOrigin][smallestIDMissing-1]
		gossip.mu.Unlock()
		var encoded []byte
		var err error
		if msgToSend.isTLC {
			encoded, err = protobuf.Encode(&GossipPacket{TLCMessage: msgToSend.tclMsg})
		} else {
			encoded, err = protobuf.Encode(&GossipPacket{Rumor: msgToSend.rumorMsg})
		}

		printerr("Gossiper Encode Error", err)
		sendPacket(encoded, sender, gossip)
		addToMongering(sender, msgToSend.Origin, msgToSend.ID)
		//fmt.Println("MONGERING with " + sender)
		go statusCountDown(msgToSend, sender, gossip)
	} else if sendStatus {
		gossip.mu.Lock()
		sendPacket(makeStatusPacket(gossip), sender, gossip)
		gossip.mu.Unlock()
	} else {
		//fmt.Println("IN SYNC WITH " + sender)

		//(3) If neither peer has new messages, the sending peer (rumormongering) peer S
		//flips a coin (e.g., ​ rand.Int() % 2​ ), and either (heads) picks ​ a new
		//random peer to send the rumor message to​ , or (tails) ceases the
		//rumormongering process.
		if isAck {
			coin := rand.Int() % 2
			heads := coin == 1
			if heads && len(Keys) > 0 {
				originalMessage := getMessage(originAcked, idAcked, gossip)
				randomPeer := Keys[rand.Intn(len(Keys))]
				var encoded []byte
				var err error
				if originalMessage.isTLC {
					encoded, err = protobuf.Encode(&GossipPacket{TLCMessage: originalMessage.tclMsg})
				} else {
					encoded, err = protobuf.Encode(&GossipPacket{Rumor: originalMessage.rumorMsg})
				}

				printerr("Gossiper Encode Error", err)
				sendPacket(encoded, randomPeer, gossip)
				addToMongering(randomPeer, originalMessage.Origin, originalMessage.ID)
				//fmt.Println("MONGERING with " + randomPeer)
				//fmt.Println("FLIPPED COIN sending rumor to " + randomPeer)
				go statusCountDown(originalMessage, randomPeer, gossip)
			}
		}
	}
}

func initNode(name string, gossip *Gossiper) {
	gossip.mu.Lock()
	defer gossip.mu.Unlock()
	_, wantsKnownForSender := gossip.wantMap[name]
	if !wantsKnownForSender {
		gossip.wantMap[name] = PeerStatus{
			Identifier: name,
			NextID:     1,
		}
	}

	_, listExists := gossip.orderedMessages[name]
	if !listExists {
		gossip.orderedMessages[name] = make([]RumorableMessage, 0)
	}

	_, listExists = gossip.earlyMessages[name]
	if !listExists {
		gossip.earlyMessages[name] = make(map[uint32]RumorableMessage)
	}

}

func getMessage(origin string, id uint32, gossip *Gossiper) RumorableMessage {
	gossip.mu.Lock()
	defer gossip.mu.Unlock()
	isInOrdered := (gossip.wantMap[origin].NextID > id)
	if isInOrdered {
		return gossip.orderedMessages[origin][id-1]
	}

	return gossip.earlyMessages[origin][id]

}

func addToMongering(dst string, origin string, ID uint32) {
	mongerer.mu.Lock()
	_, wasMongering := mongerer.mongeringMessages[dst][origin]
	if !wasMongering {
		mongerer.mongeringMessages[dst][origin] = make([]uint32, 0)
	}
	mongerer.mongeringMessages[dst][origin] = append(mongerer.mongeringMessages[dst][origin], ID)
	mongerer.mu.Unlock()
}

func makeStatusPacket(gossip *Gossiper) []byte {
	wants := make([]PeerStatus, 0)
	for _, status := range gossip.wantMap {
		wants = append(wants, status)
	}
	wantPacket := StatusPacket{Want: wants}
	newEncoded, err := protobuf.Encode(&GossipPacket{Status: &wantPacket})
	printerr("Gossiper Encode Error", err)
	return newEncoded

}

func fastForward(origin string, gossip *Gossiper) {
	currentNext := gossip.wantMap[origin].NextID
	updated := false
	indexesDelivered := make([]uint32, 0)
	for {
		for id, savedMsg := range gossip.earlyMessages[origin] {
			if savedMsg.ID == currentNext {
				currentNext += 1
				updated = true
				indexesDelivered = append(indexesDelivered, id)
				gossip.orderedMessages[origin] = append(gossip.orderedMessages[origin], savedMsg)
			}

		}
		if !updated {
			break
		} else {
			updated = false
		}
	}
	for _, index := range indexesDelivered {
		delete(gossip.earlyMessages[origin], index)
	}
	gossip.wantMap[origin] = PeerStatus{
		origin,
		currentNext,
	}

}
