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

func HandleClientRumorMessages(gossip *Gossiper, name string, peerGossiper *Gossiper) {

	//ex3: uiport, dest, msg
	//ex6: uiport,dest,file, request
	//HW1: uiport, msg

	initNode(name, peerGossiper)
	for {

		clientMessage := getAndDecodeFromClient(gossip)
		text := clientMessage.Text
		dest := clientMessage.Destination
		file := clientMessage.File
		request := clientMessage.Request

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

			messages += msg.Origin + " (private): " + msg.Text + "\n"

			newEncoded, err := protobuf.Encode(&GossipPacket{Private: &msg})
			printerr("Rumor Gossiper Error", err)

			nextHop := getNextHop(msg.Destination)
			sendPacket(newEncoded, nextHop, peerGossiper)

		} else if text != "" && !exists(dest) && !exists(file) && request == nil { //HW1 rumor message: //text, !dest, !file, !request
			fmt.Println("CLIENT MESSAGE " + text)

			peerGossiper.mu.Lock()
			msg := RumorMessage{
				Origin: name,
				ID:     getAndUpdateRumorID(),
				Text:   text,
			}

			peerGossiper.wantMap[peerGossiper.Name] = PeerStatus{
				Identifier: peerGossiper.Name,
				NextID:     msg.ID + 1,
			}
			peerGossiper.orderedMessages[peerGossiper.Name] = append(peerGossiper.orderedMessages[peerGossiper.Name], msg)
			peerGossiper.mu.Unlock();

			if msg.Text != "" {
				messages += msg.Origin + ": " + msg.Text + "\n"
			}

			newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: &msg})
			printerr("Rumor Gossiper Error", err)

			if len(KnownPeers) > 0 {
				randomPeer := Keys[rand.Intn(len(Keys))]
				sendPacket(newEncoded, randomPeer, peerGossiper)
				addToMongering(randomPeer, msg.Origin, msg.ID)

				//fmt.Println("MONGERING with " + randomPeer)

				go statusCountDown(msg, randomPeer, peerGossiper)
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

func statusCountDown(msg RumorMessage, dst string, gossip *Gossiper) {

	ticker := time.NewTicker(STATUS_COUNTDOWN_TIME * time.Second)
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


