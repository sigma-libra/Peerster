package server

import (
	"encoding/hex"
	"fmt"
	"github.com/dedis/protobuf"
	"math/rand"
	"time"
)

func HandleRumorMessagesFrom(gossiper *Gossiper) {

	if RTimer > 0 {
		SendRouteRumor(gossiper)
		go FireRouteRumor(gossiper)
	}

	for {

		pkt, sender, arrivalTime := getAndDecodePacket(gossiper)

		AddPeer(sender)
		//fmt.Println("PEERS " + FormatPeers(Keys))

		if pkt.Rumor != nil {
			msg := pkt.Rumor

			rumorWrapper := RumorableMessage{
				Origin:   msg.Origin,
				ID:       msg.ID,
				isTLC:    false,
				rumorMsg: msg,
				tclMsg:   nil,
			}

			handleRumorableMessage(&rumorWrapper, sender, gossiper)

		} else if pkt.Status != nil {
			msg := pkt.Status
			handleStatusMessage(msg, sender, gossiper)
		} else if pkt.Private != nil {
			msg := pkt.Private
			handlePrivateMessage(msg, gossiper)
		} else if pkt.DataRequest != nil {
			msg := pkt.DataRequest
			handleRequestMessage(msg, gossiper)
		} else if pkt.DataReply != nil {
			msg := pkt.DataReply
			handleReplyMessage(msg, gossiper)
		} else if pkt.SearchRequest != nil {
			msg := pkt.SearchRequest
			handleSearchRequest(*msg, gossiper, arrivalTime)
		} else if pkt.SearchReply != nil {
			msg := pkt.SearchReply
			handleSearchReply(msg, gossiper)
		} else if pkt.TLCMessage != nil {
			msg := pkt.TLCMessage
			rumorWrapper := RumorableMessage{
				Origin:   msg.Origin,
				ID:       msg.ID,
				isTLC:    true,
				rumorMsg: nil,
				tclMsg:   msg,
			}
			handleRumorableMessage(&rumorWrapper, sender, gossiper)
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
		keywords := clientMessage.Keywords
		budget := clientMessage.Budget

		if text == "" && exists(file) && request != nil { //ex6: !text, dest, file, request -> get file from dest with hash

			var fileOwner string
			if !exists(dest) {
				SearchReplyTracker.Mu.Lock()
				senders := SearchReplyTracker.Messages[*file]
				for sender, reply := range senders {
					if reply.ChunkCount == uint64(len(reply.ChunkMap)) {
						fileOwner = sender
						break
					}
				}
				SearchReplyTracker.Mu.Unlock()

			} else {
				fileOwner = *dest
			}

			key := hex.EncodeToString(*request)

			msg := DataRequest{
				Origin:      name,
				Destination: fileOwner,
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

		} else if text != "" && exists(dest) && !exists(file) && request == nil { //case ex3: text, dest, !file, !request -> send private message
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

		} else if text != "" && !exists(dest) && !exists(file) && request == nil { //HW1 rumor message: //text, !dest, !file, !request -> send public message
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
			wrapper := RumorableMessage{
				Origin:   msg.Origin,
				ID:       msg.ID,
				isTLC:    false,
				rumorMsg: &msg,
				tclMsg:   nil,
			}

			peerGossiper.orderedMessages[peerGossiper.Name] = append(peerGossiper.orderedMessages[peerGossiper.Name], wrapper)
			peerGossiper.mu.Unlock()
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

				go statusCountDown(wrapper, randomPeer, peerGossiper)
			}

		} else if text == "" && !exists(dest) && exists(file) && request == nil { //hw4 - upload file: !text, !dest, file, !request -> upload file
			ReadFileIntoChunks(*file)
		} else if keywords != nil && len(*keywords) > 0 { //keyword

		var increasingBudget bool
			if *budget == 0 {
				*budget = 2
				increasingBudget = true
			} else {
				increasingBudget = false
			}

			msg := SearchRequest{
				Origin:   name,
				Budget:   *budget,
				Keywords: *keywords,
			}
			go SendRepeatedSearchRequests(msg, peerGossiper, increasingBudget)
		}
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

func statusCountDown(msg RumorableMessage, dst string, gossip *Gossiper) {

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
			var encoded []byte
			var err error
			if msg.isTLC {
				encoded, err = protobuf.Encode(&GossipPacket{TLCMessage: msg.tclMsg})
			} else {
				encoded, err = protobuf.Encode(&GossipPacket{Rumor: msg.rumorMsg})
			}
			printerr("Rumor Gossiper Error", err)
			randomPeer := Keys[rand.Intn(len(Keys))]
			sendPacket(encoded, randomPeer, gossip)
			addToMongering(randomPeer, msg.Origin, msg.ID)

			//fmt.Println("MONGERING with " + randomPeer)

			go statusCountDown(msg, randomPeer, gossip)

		}
	}

}
