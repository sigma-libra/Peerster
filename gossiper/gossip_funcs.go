package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
	"math"
	"math/rand"
	"strconv"
)

func handleRumorMessage(msg *RumorMessage, sender string, gossip *Gossiper) {

	initNode(msg.Origin)

	//update routing table
	if msg.Origin != PeerName {
		lastID, originKnown := routingTable.LastMsgID[msg.Origin]
		prevSender, prevExists := routingTable.Table[msg.Origin]

		if !originKnown || lastID < msg.ID  {
			routingTable.Table[msg.Origin] = sender
			routingTable.LastMsgID[msg.Origin] = msg.ID

			if !prevExists || (prevExists&& prevSender != sender) {
				fmt.Println("DSDV " + msg.Origin + " " + sender)
			}
		}
	}

	if msg.Text != "" {
		printMsg := "RUMOR origin " + msg.Origin + " from " + sender + " ID " + strconv.FormatUint(uint64(msg.ID), 10) + " contents " + msg.Text
		fmt.Println(printMsg)
	}

	receivedBefore := (wantMap[msg.Origin].NextID > msg.ID) || (msg.Origin == PeerName)

	//message not received before: start mongering
	if !receivedBefore {

		if msg.Text != "" {
			messages += msg.Origin + ": " + msg.Text + "\n"
		}

		//pick random peer to send to
		randomPeer := Keys[rand.Intn(len(Keys))]

		newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: msg})
		if err != nil {
			println("Gossiper Encode Error: " + err.Error())
		}
		sendPacket(newEncoded, randomPeer, gossip)

		//fmt.Println("MONGERING with " + randomPeer)
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
}

func handleStatusMessage(msg *StatusPacket, sender string, gossip *Gossiper) {
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

func initNode(name string) {

	_, wantsKnownForSender := wantMap[name]
	if !wantsKnownForSender {
		wantMap[name] = PeerStatus{
			Identifier: name,
			NextID:     1,
		}
	}

	_, listExists := orderedMessages[name]
	if !listExists {
		orderedMessages[name] = make([]RumorMessage, 0)
	}

	_, listExists = earlyMessages[name]
	if !listExists {
		earlyMessages[name] = make(map[uint32]RumorMessage)
	}

}

func getMessage(origin string, id uint32) RumorMessage {
	isInOrdered := (wantMap[origin].NextID > id)
	if isInOrdered {
		return orderedMessages[origin][id-1]
	}

	return earlyMessages[origin][id]

}

func addToMongering(dst string, origin string, ID uint32) {
	_, wasMongering := mongeringMessages[dst][origin]
	if !wasMongering {
		mongeringMessages[dst][origin] = make([]uint32, 0)
	}
	mongeringMessages[dst][origin] = append(mongeringMessages[dst][origin], ID)
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

func fastForward(origin string) {
	currentNext := wantMap[origin].NextID
	updated := false
	indexesDelivered := make([]uint32, 0)
	for {
		for id, savedMsg := range earlyMessages[origin] {
			if savedMsg.ID == currentNext {
				currentNext += 1
				updated = true
				indexesDelivered = append(indexesDelivered, id)
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
		delete(earlyMessages[origin], index)
	}
	wantMap[origin] = PeerStatus{
		origin,
		currentNext,
	}

}
