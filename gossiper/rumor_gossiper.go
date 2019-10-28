package gossiper

import (
	"crypto/sha256"
	"fmt"
	"github.com/SabrinaKall/Peerster/helper"
	"github.com/dedis/protobuf"
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
		//fmt.Println("PEERS " + FormatPeers(Keys))

		if pkt.Rumor != nil {
			msg := pkt.Rumor

			handleRumorMessage(msg, sender, gossip)

		} else if pkt.Status != nil {
			// packet is status
			msg := pkt.Status

			handleStatusMessage(msg, sender, gossip)

		} else if pkt.Private != nil {
			msg := pkt.Private
			if msg.Destination == PeerName {
				fmt.Println("PRIVATE origin " + msg.Origin + " hop-limit " + strconv.FormatInt(int64(msg.HopLimit), 10) + " contents " + msg.Text)
			} else {
				if msg.HopLimit > 0 {
					msg.HopLimit -= 1
					newEncoded, err := protobuf.Encode(&GossipPacket{Private: msg})
					if err != nil {
						println("Gossiper Encode Error: " + err.Error())
					}
					nextHop := routingTable.Table[msg.Destination]
					sendPacket(newEncoded, nextHop, gossip)
				}
			}

		} else if pkt.DataRequest != nil {
			msg := pkt.DataRequest
			if msg.Destination == gossip.Name {

				data := getDataFor(msg.HashValue)

				reply := DataReply{
					Origin:      PeerName,
					Destination: msg.Origin,
					HopLimit:    HopLimit - 1,
					HashValue:   msg.HashValue,
					Data:        data,
				}
				newEncoded, err := protobuf.Encode(&GossipPacket{DataReply: &reply})
				if err != nil {
					println("Gossiper Encode Error: " + err.Error())
				}
				nextHop := routingTable.Table[msg.Destination]
				sendPacket(newEncoded, nextHop, gossip)

			} else {
				if msg.HopLimit > 0 {
					msg.HopLimit -= 1
					newEncoded, err := protobuf.Encode(&GossipPacket{DataRequest: msg})
					if err != nil {
						println("Gossiper Encode Error: " + err.Error())
					}
					nextHop := routingTable.Table[msg.Destination]
					sendPacket(newEncoded, nextHop, gossip)
				}
			}

		} else if pkt.DataReply != nil {
			msg := pkt.DataReply
			if msg.Destination == gossip.Name {
				//save chunk/metafile and send for next
				metaHash := [32]byte{}
				copy(metaHash[:], msg.HashValue[:])
				inProgress := DownloadsInProgress[metaHash]
				fileInfo := Files[metaHash]
				shaCheck := sha256.Sum256(msg.Data)
				if helper.Equal(msg.HashValue, inProgress.hashCurrentlyBeingFetched) && helper.Equal(shaCheck[:], msg.HashValue) {

					if !inProgress.metafileFetched {
						fileInfo.metafile = msg.Data
						inProgress.metafile = msg.Data

						inProgress.metafileFetched = true
						inProgress.nbChunks = len(inProgress.metafile) / 32
					} else {
						inProgress.chunks[inProgress.chunkIndexBeingFetched] = msg.Data

						key := [32]byte{}
						for i := 0; i < 32; i++ {
							key[i] = inProgress.hashCurrentlyBeingFetched[i]
						}
						fileInfo.chunks[key] = msg.Data
						fileInfo.filesize += len(msg.Data)

					}

					if inProgress.chunkIndexBeingFetched < inProgress.nbChunks {

						fmt.Println("DOWNLOADING " + inProgress.filename + " chunk " +strconv.Itoa(inProgress.chunkIndexBeingFetched + 1)+"  from " + msg.Origin)
						nextChunkHash := inProgress.metafile[32*inProgress.chunkIndexBeingFetched : 32*(inProgress.chunkIndexBeingFetched+1)]

						newMsg := DataRequest{
							Origin:      PeerName,
							Destination: msg.Origin,
							HopLimit:    HopLimit - 1,
							HashValue:   nextChunkHash,
						}

						newEncoded, err := protobuf.Encode(&GossipPacket{DataRequest: &newMsg})
						if err != nil {
							println("Gossiper Encode Error: " + err.Error())
						}
						nextHop := routingTable.Table[newMsg.Destination]
						sendPacket(newEncoded, nextHop, gossip)

						inProgress.chunkIndexBeingFetched += 1
						inProgress.hashCurrentlyBeingFetched = nextChunkHash

						DownloadsInProgress[metaHash] = inProgress
						Files[metaHash] = fileInfo


						if inProgress.chunkIndexBeingFetched == inProgress.nbChunks  {
							downloadFile(inProgress)
							delete(DownloadsInProgress, metaHash)
						} else {
							go downloadCountDown(metaHash, nextChunkHash, newMsg, gossip)
						}


					}

				}

			} else {
				if msg.HopLimit > 0 {
					msg.HopLimit -= 1
					newEncoded, err := protobuf.Encode(&GossipPacket{DataReply: msg})
					if err != nil {
						println("Gossiper Encode Error: " + err.Error())
					}
					nextHop := routingTable.Table[msg.Destination]
					sendPacket(newEncoded, nextHop, gossip)
				}
			}

		}

	}
}

func HandleClientRumorMessages(gossip *Gossiper, name string, peerGossiper *Gossiper) {

	//ex3: uiport, dest, msg
	//ex6: uiport,dest,file, request
	//HW1: uiport, msg

	for {

		clientMessage := getAndDecodeFromClient(gossip)
		text := clientMessage.Text
		dest := clientMessage.Destination
		file := clientMessage.File
		request := clientMessage.Request

		if dest != nil && *dest != "" && file != nil && *file != "" && request != nil { //ex6: uiport, dest,file, request

			key := [32]byte{}
			for i := 0; i < 32; i++ {
				key[i] = (*request)[i]
			}

			DownloadsInProgress[key] = DownloadInProgress{
				filename:                  *file,
				metafile:                  []byte{},
				chunks:                    make([][]byte, 0),
				metafileFetched:           false,
				chunkIndexBeingFetched:    0,
				hashCurrentlyBeingFetched: *request,
			}

			Files[key] = FileInfo{
				filename: *file,
				filesize: 0,
				metafile: []byte{},
				chunks:   make(map[[32]byte][]byte),
				metahash: key,
			}

			fmt.Println("DOWNLOADING metafile of "+ *file +" from " + *dest)

			msg := DataRequest{
				Origin:      name,
				Destination: *dest,
				HopLimit:    HopLimit - 1,
				HashValue:   *request,
			}

			newEncoded, err := protobuf.Encode(&GossipPacket{DataRequest: &msg})
			if err != nil {
				println("Gossiper Encode Error: " + err.Error())
			}

			nextHop := routingTable.Table[msg.Destination]
			sendPacket(newEncoded, nextHop, peerGossiper)

			go downloadCountDown(key, *request, msg, peerGossiper)

		} else if text != "" && dest != nil && *dest != "" { //case ex3: uiport, dest, msg
			fmt.Println("CLIENT MESSAGE " + text)

			msg := PrivateMessage{
				Origin:      name,
				ID:          0,
				Text:        text,
				Destination: *dest,
				HopLimit:    HopLimit - 1,
			}

			messages += msg.Origin + " (private): " + msg.Text + "\n"

			newEncoded, err := protobuf.Encode(&GossipPacket{Private: &msg})
			if err != nil {
				println("Gossiper Encode Error: " + err.Error())
			}

			nextHop := routingTable.Table[msg.Destination]
			sendPacket(newEncoded, nextHop, peerGossiper)

		} else if text != "" { //HW1 rumor message
			fmt.Println("CLIENT MESSAGE " + text)

			msg := RumorMessage{
				Origin: name,
				ID:     getAndUpdateRumorID(),
				Text:   text,
			}

			if msg.Text != "" {
				messages += msg.Origin + ": " + msg.Text + "\n"
			}

			newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: &msg})
			if err != nil {
				println("Gossiper Encode Error: " + err.Error())
			}

			if len(KnownPeers) > 0 {
				randomPeer := Keys[rand.Intn(len(Keys))]
				sendPacket(newEncoded, randomPeer, peerGossiper)
				addToMongering(randomPeer, msg.Origin, msg.ID)

				//fmt.Println("MONGERING with " + randomPeer)

				go statusCountDown(msg, randomPeer, peerGossiper)
			}

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

			//fmt.Println("MONGERING with " + randomPeer)

			go statusCountDown(msg, randomPeer, gossip)

		}
	}

}

func downloadCountDown(key [32]byte, hash []byte, msg DataRequest, peerGossiper *Gossiper) {

	ticker := time.NewTicker(5 * time.Second)
	<-ticker.C

	download := DownloadsInProgress[key]
	if helper.Equal(download.hashCurrentlyBeingFetched, hash) {

		newEncoded, err := protobuf.Encode(&GossipPacket{DataRequest: &msg})
		if err != nil {
			println("Gossiper Encode Error: " + err.Error())
		}

		nextHop := routingTable.Table[msg.Destination]
		sendPacket(newEncoded, nextHop, peerGossiper)
		go downloadCountDown(key, hash, msg, peerGossiper)
	}

}
