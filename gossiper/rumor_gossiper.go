//Author: Sabrina Kall
package gossiper

import (
	"encoding/hex"
	"fmt"
	"github.com/dedis/protobuf"
	"math/rand"
	"time"
)

func HandleRumorMessagesFrom(gossip *Gossiper) {

	if RTimer > 0 {
		SendRouteRumor(gossip)
		go FireRouteRumor(gossip)
	}

	statusPrinted = false

	for {

		pkt, sender := getAndDecodePacket(gossip)

		AddPeer(sender)
		//fmt.Println("PEERS " + FormatPeers(Keys))

		if pkt.Rumor != nil {
			msg := pkt.Rumor
			handleRumorMessage(msg, sender, gossip)
		} else if pkt.Status != nil {
			msg := pkt.Status
			handleStatusMessage(msg, sender, gossip)
		} else if pkt.Private != nil {
			msg := pkt.Private
			handlePrivateMessage(msg, gossip)
		} else if pkt.DataRequest != nil {
			msg := pkt.DataRequest
			handleRequestMessage(msg, gossip)
		} else if pkt.DataReply != nil {
			msg := pkt.DataReply
			handleReplyMessage(msg, gossip)
		}
	}
}

func exists(str *string) bool {
	return str != nil && *str != ""
}

func HandleClientRumorMessages(clientGossiper *Gossiper, name string, peerGossiper *Gossiper) {

	//ex3: uiport, dest, msg
	//ex6: uiport,dest,file, request
	//HW1: uiport, msg

	initNode(name, peerGossiper)
	for {

		clientMessage := getAndDecodeFromClient(clientGossiper)
		text := clientMessage.Text
		dest := clientMessage.Destination
		file := clientMessage.File
		request := clientMessage.Request
		msgGroups := clientMessage.Groups

		if text == "" && exists(dest) && exists(file) && request != nil { //ex6: !text, dest, file, request

			key := hex.EncodeToString(*request)

			msg := DataRequest{
				Origin:      name,
				Destination: *dest,
				HopLimit:    HOP_LIMIT - 1,
				HashValue:   *request,
			}

			newEncoded, err := protobuf.Encode(&GossipPacket{DataRequest: &msg})
			printerr("Rumor Gossiper Error", err)

			putInFileMemory(InitFileInfo(*file, *request))
			nextHop := getNextHop(msg.Destination)
			sendPacket(newEncoded, nextHop, peerGossiper)

			fmt.Println("DOWNLOADING metafile of " + *file + " from " + *dest)

			go downloadCountDown(key, *request, msg, peerGossiper)

		} else if text != "" && exists(dest) && !exists(file) && request == nil { //case ex3: text, dest, !file, !request
			//CLIENT MESSAGE <msg_text> dest <dst_name>
			fmt.Println("CLIENT MESSAGE " + text + " dest " + *dest)

			msg := PrivateMessage{
				Origin:      name,
				ID:          0,
				Text:        text,
				Destination: *dest,
				HopLimit:    HOP_LIMIT - 1,
			}

			msgGroups := [2]string{msg.Origin}
			for _, gr := range msgGroups {
				_, known := messages[gr]
				if !known {
					messages[gr] = msg.Origin + " (private): " + msg.Text + "\n"
				} else {
					messages[gr] += msg.Origin + " (private): " + msg.Text + "\n"
				}
			}

			newEncoded, err := protobuf.Encode(&GossipPacket{Private: &msg})
			printerr("Rumor Gossiper Error", err)

			nextHop := getNextHop(msg.Destination)
			sendPacket(newEncoded, nextHop, peerGossiper)

		} else if text != "" && !exists(dest) && !exists(file) && request == nil { //HW1 rumor message: //text, !dest, !file, !request

			groupstr := ""
			for _, grp := range msgGroups {
				groupstr += grp + ", "
			}
			if groupstr != "" {
				groupstr = groupstr[:len(groupstr)-2]
			}
			fmt.Println("CLIENT MESSAGE " + text + "(Groups: " + groupstr + ")")

			if len(msgGroups) == 0 || (len(msgGroups) == 1 && msgGroups[0] == "") || (len(msgGroups) == 1 && msgGroups[0] == " ") {
				fmt.Println("no group found")
				msgGroups = append(msgGroups, "no group")
			}

			peerGossiper.mu.Lock()
			msg := RumorMessage{
				Origin: name,
				ID:     getAndUpdateRumorID(),
				Text:   text,
				Groups: msgGroups,
			}

			peerGossiper.wantMap[peerGossiper.Name] = PeerStatus{
				Identifier: peerGossiper.Name,
				NextID:     msg.ID + 1,
				Groups:     peerGossiper.wantMap[peerGossiper.Name].Groups,
			}

			peerGossiper.orderedMessages[peerGossiper.Name] = append(peerGossiper.orderedMessages[peerGossiper.Name], msg)
			for _, grp := range msg.Groups {
				peerGossiper.groupMessages[grp] = append(peerGossiper.groupMessages[grp], msg)
			}

			peerGossiper.mu.Unlock()

			msgGroups := msg.Groups

			for _, gr := range msgGroups {
				_, known := messages[gr]
				wanted := Groups[gr]
				if wanted {
					if !known {
						messages[gr] = msg.Origin + ": " + msg.Text + "\n"
					} else {
						messages[gr] += msg.Origin + ": " + msg.Text + "\n"
					}
				}
			}

			newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: &msg})
			printerr("Rumor Gossiper Error", err)

			if len(KnownPeers) > 0 {
				peerGossiper.mu.Lock()
				nextPeer := get_peer_with_group(msg.Groups, *peerGossiper, msg.Text)
				peerGossiper.mu.Unlock()
				sendPacket(newEncoded, nextPeer, peerGossiper)
				addToMongering(nextPeer, msg.Origin, msg.ID)

				//fmt.Println("MONGERING with " + nextPeer)

				go statusCountDown(msg, nextPeer, peerGossiper)
			}

		} else if text == "" && !exists(dest) && exists(file) && request == nil { //hw4 - upload file: !text, !dest, file, !request
			ReadFileIntoChunks(*file)
		}

		//fmt.Println("PEERS " + FormatPeers(Keys))
	}

}

func FireAntiEntropy(gossip *Gossiper) {
	for {
		ticker := time.NewTicker(time.Duration(AntiEntropy) * time.Second)
		<-ticker.C
		if len(Keys) > 0 {
			randomPeer := Keys[rand.Intn(len(Keys))]
			gossip.mu.Lock()
			sendPacket(makeStatusPacket(gossip), randomPeer, gossip)
			gossip.mu.Unlock()
		}

	}
}

func UpdateMessages(gossip *Gossiper) {
	ticker := time.NewTicker(STATUS_COUNTDOWN_TIME * time.Second)
	<-ticker.C

	if FilterIncomingPackets {
		for grp, here := range Groups {
			if here {
				_, alreadyPrinted := messages[grp]
				if !alreadyPrinted {
					gossip.mu.Lock()
					deliveredMessages, hasMessages := gossip.groupMessages[grp]
					gossip.mu.Unlock()
					if hasMessages {
						messages[grp] = ""
						for _, oldMessage := range deliveredMessages {
							if oldMessage.Text != "" {
								messages[grp] += oldMessage.Origin + ": " + oldMessage.Text + "\n"
							}
						}
					}
				}
			}
		}
		go UpdateMessages(gossip)
	}

}

func statusCountDown(msg RumorMessage, dst string, gossip *Gossiper) {

	ticker := time.NewTicker(2 * time.Second)
	<-ticker.C

	mongerer.mu.Lock()
	mongeredIDs, dstTracked := mongerer.mongeringMessages[dst][msg.Origin]
	mongerer.mu.Unlock()

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
			printerr("Rumor Gossiper Error", err)
			randomPeer := Keys[rand.Intn(len(Keys))]
			sendPacket(encoded, randomPeer, gossip)
			addToMongering(randomPeer, msg.Origin, msg.ID)

			//fmt.Println("MONGERING with " + randomPeer)

			go statusCountDown(msg, randomPeer, gossip)

		}
	}

}
