package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
	"math"
	"math/rand"
	"strconv"
)

func handleRumorMessage(msg *RumorMessage, sender string, gossip *Gossiper) {

	//TODO TEST THIS
	//if msg.Origin != gossip.Name {
	initNode(msg.Origin, gossip)

	printMsg := "RUMOR origin " + msg.Origin + " from " + sender + " ID " + strconv.FormatUint(uint64(msg.ID), 10) + " contents " + msg.Text
	fmt.Println(printMsg)

	//update routing table
	routingTable.mu.Lock()
	lastID, lastIDExists := routingTable.LastMsgID[msg.Origin]

	if !lastIDExists || lastID < msg.ID {
		routingTable.Table[msg.Origin] = sender
		routingTable.LastMsgID[msg.Origin] = msg.ID

		if msg.Text != "" {
			fmt.Println("DSDV " + msg.Origin + " " + sender)
		}
	}
	routingTable.mu.Unlock()

	gossip.mu.Lock()
	defer gossip.mu.Unlock()
	receivedBefore := (gossip.wantMap[msg.Origin].NextID > msg.ID) || (msg.Origin == gossip.Name)

	//message not received before: start mongering
	if !receivedBefore {

		if msg.Text != "" && (msg.Origin != gossip.Name) {
			messages += msg.Origin + ": " + msg.Text + "\n"
		}

		//pick random peer to send to
		randomPeer := Keys[rand.Intn(len(Keys))]

		newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: msg})
		printerr("Gossiper Encode Error", err)
		sendPacket(newEncoded, randomPeer, gossip)

		//fmt.Println("MONGERING with " + randomPeer)
		addToMongering(randomPeer, msg.Origin, msg.ID)

		//if next message we want, save in vector clock
		if gossip.wantMap[msg.Origin].NextID == msg.ID {
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

		//ack reception of packet
		sendPacket(makeStatusPacket(gossip), sender, gossip)
		//start countdown to monger to someone else
		go statusCountDown(*msg, randomPeer, gossip)

	}
	//}
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
		msgToSend := gossip.orderedMessages[smallestOrigin][smallestIDMissing-1]
		rumorEncoded, err := protobuf.Encode(&GossipPacket{Rumor: &msgToSend})
		printerr("Gossiper Encode Error", err)

		sendPacket(rumorEncoded, sender, gossip)
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
			if heads && len(KnownPeers) > 0 {
				originalMessage := getMessage(originAcked, idAcked, gossip)
				randomPeer := Keys[rand.Intn(len(Keys))]
				newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: &originalMessage})
				printerr("Gossiper Encode Error", err)
				sendPacket(newEncoded, randomPeer, gossip)
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
		gossip.orderedMessages[name] = make([]RumorMessage, 0)
	}

	_, listExists = gossip.earlyMessages[name]
	if !listExists {
		gossip.earlyMessages[name] = make(map[uint32]RumorMessage)
	}

}

func getMessage(origin string, id uint32, gossip *Gossiper) RumorMessage {
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
