package gossiper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/SabrinaKall/Peerster/helper"
	"github.com/dedis/protobuf"
	"math"
	"math/rand"
	"strconv"
	"time"
)

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
			if msg.Destination == gossip.Name {
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
			fmt.Println("TEST Datarequest")
			msg := pkt.DataRequest
			if msg.Destination == gossip.Name {

				testPrint("Request for me")

				testPrint("Hash to get: " + hex.EncodeToString(msg.HashValue))
				data, _ := getDataFor(msg.HashValue)

				reply := DataReply{
					Origin:      gossip.Name,
					Destination: msg.Origin,
					HopLimit:    HopLimit - 1,
					HashValue:   msg.HashValue,
					Data:        data,
				}
				newEncoded, err := protobuf.Encode(&GossipPacket{DataReply: &reply})
				if err != nil {
					println("Gossiper Encode Error: " + err.Error())
				}
				nextHop := routingTable.Table[reply.Destination]
				sendPacket(newEncoded, nextHop, gossip)

			} else {

				testPrint("Request for " + msg.Destination)
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
				testPrint("Reply for me")

				fileInfo, meta, here, err := findFileWithHash(msg.HashValue)
				if err != nil {
					fmt.Println(err)
				}

				if here && !fileInfo.downloadComplete && len(msg.Data) > 0 {
					shaCheck := sha256.Sum256(msg.Data)

					if helper.Equal(msg.HashValue, fileInfo.hashCurrentlyBeingFetched) && helper.Equal(shaCheck[:], msg.HashValue) {

						if !fileInfo.metafileFetched && meta {
							fmt.Println("DOWNLOADING metafile of " + fileInfo.filename + " from " + msg.Origin)
							fileInfo.metafile = msg.Data
							fileInfo.metafileFetched = true
							fileInfo.chunkIndexBeingFetched = 0

							fileInfo.nbChunks = int(math.Ceil(float64(float64(len(msg.Data)) / float64(32))))

						} else {
							fmt.Println("DOWNLOADING " + fileInfo.filename + " chunk " + strconv.Itoa(fileInfo.chunkIndexBeingFetched+1) + "  from " + msg.Origin)
							key := hex.EncodeToString(fileInfo.hashCurrentlyBeingFetched)
							fileInfo.orderedHashes[key] = fileInfo.chunkIndexBeingFetched
							fileInfo.orderedChunks[fileInfo.chunkIndexBeingFetched] = msg.Data

							fileInfo.filesize += len(msg.Data)
							fileInfo.chunkIndexBeingFetched += 1

						}

						fmt.Println("Index being fetched: " + strconv.Itoa(fileInfo.chunkIndexBeingFetched))
						fmt.Println("Nb Chunks: " + strconv.Itoa(fileInfo.nbChunks))

						//TODO figure out order of index incrementation
						if fileInfo.chunkIndexBeingFetched >= fileInfo.nbChunks {
							downloadFile(*fileInfo)
						} else {
							nextChunkHash := fileInfo.metafile[32*(fileInfo.chunkIndexBeingFetched) : 32*(fileInfo.chunkIndexBeingFetched+1)]
							testPrint("Next hash: " + hex.EncodeToString(nextChunkHash))
							fileInfo.hashCurrentlyBeingFetched = nextChunkHash
							testPrint("Saved next hash: " + hex.EncodeToString(fileInfo.hashCurrentlyBeingFetched))

							newMsg := DataRequest{
								Origin:      gossip.Name,
								Destination: msg.Origin,
								HopLimit:    HopLimit - 1,
								HashValue:   nextChunkHash,
							}
							testPrint("Hash for new chunk: " + hex.EncodeToString(nextChunkHash))

							newEncoded, err := protobuf.Encode(&GossipPacket{DataRequest: &newMsg})
							if err != nil {
								println("Gossiper Encode Error: " + err.Error())
							}

							Files[fileInfo.metahash] = *fileInfo

							nextHop := routingTable.Table[newMsg.Destination]
							sendPacket(newEncoded, nextHop, gossip)

							go downloadCountDown(hex.EncodeToString(msg.HashValue), nextChunkHash, newMsg, gossip)
						}

					}

				} else {

					testPrint("Reply for " + msg.Destination)
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

			key := hex.EncodeToString(*request)

			fmt.Println("Hash at client handler: " + key)

			Files[key] = FileInfo{
				filename:                  *file,
				filesize:                  0,
				metafile:                  nil,
				orderedHashes:             make(map[string]int),
				orderedChunks:             make(map[int][]byte),
				metahash:                  key,
				downloadComplete:          false,
				metafileFetched:           false,
				chunkIndexBeingFetched:    0,
				hashCurrentlyBeingFetched: *request,
				nbChunks:                  0,
			}

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
			sendPacket(makeStatusPacket(gossip), randomPeer, gossip)
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

func downloadCountDown(key string, hash []byte, msg DataRequest, peerGossiper *Gossiper) {

	ticker := time.NewTicker(5 * time.Second)
	<-ticker.C

	fileInfo, _, _, _ := findFileWithHash(hash)
	if helper.Equal(fileInfo.hashCurrentlyBeingFetched, hash) {

		newEncoded, err := protobuf.Encode(&GossipPacket{DataRequest: &msg})
		if err != nil {
			println("Gossiper Encode Error: " + err.Error())
		}

		nextHop := routingTable.Table[msg.Destination]
		sendPacket(newEncoded, nextHop, peerGossiper)
		go downloadCountDown(key, hash, msg, peerGossiper)
	}

}

func testPrint(msg string) {
	fmt.Println("TEST " + msg)
}
