package gossiper

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/SabrinaKall/Peerster/helper"
	"github.com/dedis/protobuf"
	"io/ioutil"
	"math"
	"strconv"
)

func ReadFileIntoChunks(filename string) {

	fmt.Println("reading file " + filename + " into chunks")

	data, err := ioutil.ReadFile(FILE_FOLDER + filename)
	checkErr(err)

	nbChunks := int(math.Ceil(float64(len(data)) / float64(CHUNK_SIZE)))

	if nbChunks == 0 {
		fmt.Println("TEST Rounding up")
		nbChunks = 1
	}

	fmt.Println("Nb chunks: " + strconv.Itoa(nbChunks))

	fileInfo := FileInfo{
		filename:                  filename,
		filesize:                  len(data),
		metafile:                  make([]byte, SHA_SIZE*nbChunks),
		orderedHashes:             make(map[string]int),
		orderedChunks:             make(map[int][]byte),
		metahash:                  "",
		downloadComplete:          true,
		downloadInterrupted:       false,
		metafileFetched:           true,
		chunkIndexBeingFetched:    nbChunks,
		hashCurrentlyBeingFetched: nil,
		nbChunks:                  nbChunks,
	}

	for i := 0; i < nbChunks; i++ {

		var buf []byte
		if i == nbChunks-1 {
			buf = data[i*CHUNK_SIZE:]
		} else {
			buf = data[i*CHUNK_SIZE : ((i + 1) * CHUNK_SIZE)]
		}
		chunkSha := sha256.Sum256(buf)
		for j := 0; j < len(chunkSha); j++ {
			fileInfo.metafile[(i*SHA_SIZE)+j] = chunkSha[j]
		}
		chunkShaString := hex.EncodeToString(chunkSha[:])

		fileInfo.orderedHashes[chunkShaString] = i
		fileInfo.orderedChunks[i] = buf
	}

	metahash := sha256.Sum256(fileInfo.metafile)
	fileInfo.metahash = hex.EncodeToString(metahash[:])

	Files[fileInfo.metahash] = fileInfo

	fmt.Println("Metahash: " + fileInfo.metahash)
}

func handleRequestMessage(msg *DataRequest, gossip *Gossiper) {
	if msg.Destination == gossip.Name {

		testPrint("Request for me")

		testPrint("Hash to get: " + hex.EncodeToString(msg.HashValue))
		data, _ := getDataFor(msg.HashValue)

		reply := DataReply{
			Origin:      gossip.Name,
			Destination: msg.Origin,
			HopLimit:    HOP_LIMIT,
			HashValue:   msg.HashValue,
			Data:        data,
		}
		newEncoded, err := protobuf.Encode(&GossipPacket{DataReply: &reply})
		if err != nil {
			println("Gossiper Encode Error: " + err.Error())
		}

		nextHop := RoutingTable.Table[reply.Destination]
		sendPacket(newEncoded, nextHop, gossip)

	} else {

		testPrint("Request for " + msg.Destination)
		if msg.HopLimit > 0 {
			msg.HopLimit -= 1
			newEncoded, err := protobuf.Encode(&GossipPacket{DataRequest: msg})
			if err != nil {
				println("Gossiper Encode Error: " + err.Error())
			}
			nextHop := RoutingTable.Table[msg.Destination]
			sendPacket(newEncoded, nextHop, gossip)
		}
	}

}

func handleReplyMessage(msg *DataReply, gossip *Gossiper) {
	if msg.Destination == gossip.Name {
		//save chunk/metafile and send for next
		testPrint("Reply for me")

		fileInfo, meta, here, err := findFileWithHash(msg.HashValue)
		if err != nil {
			fmt.Println(err)
		}

		if here && !fileInfo.downloadComplete && !fileInfo.downloadInterrupted {

			if len(msg.Data) == 0 {
				fileInfo.downloadInterrupted = true
			} else {
				shaCheck := sha256.Sum256(msg.Data)

				if helper.Equal(msg.HashValue, fileInfo.hashCurrentlyBeingFetched) && helper.Equal(shaCheck[:], msg.HashValue) {

					if !fileInfo.metafileFetched && meta {
						fileInfo.metafile = msg.Data
						fileInfo.metafileFetched = true
						fileInfo.chunkIndexBeingFetched = 0

						fileInfo.nbChunks = int(math.Ceil(float64(float64(len(msg.Data)) / float64(32))))

					} else {
						key := hex.EncodeToString(fileInfo.hashCurrentlyBeingFetched)
						fileInfo.orderedHashes[key] = fileInfo.chunkIndexBeingFetched
						fileInfo.orderedChunks[fileInfo.chunkIndexBeingFetched] = msg.Data

						fileInfo.filesize += len(msg.Data)
						fileInfo.chunkIndexBeingFetched += 1

					}

					if fileInfo.chunkIndexBeingFetched >= fileInfo.nbChunks {
						downloadFile(*fileInfo)
						fileInfo.downloadComplete = true
					} else {
						fmt.Println("DOWNLOADING " + fileInfo.filename + " chunk " + strconv.Itoa(fileInfo.chunkIndexBeingFetched+1) + " from " + msg.Origin)
						nextChunkHash := fileInfo.metafile[SHA_SIZE*(fileInfo.chunkIndexBeingFetched) : SHA_SIZE*(fileInfo.chunkIndexBeingFetched+1)]
						fileInfo.hashCurrentlyBeingFetched = nextChunkHash

						newMsg := DataRequest{
							Origin:      gossip.Name,
							Destination: msg.Origin,
							HopLimit:    HOP_LIMIT,
							HashValue:   nextChunkHash,
						}
						testPrint("Hash for new chunk: " + hex.EncodeToString(nextChunkHash))
						testPrint("Origin: " + newMsg.Origin)

						newEncoded, err := protobuf.Encode(&GossipPacket{DataRequest: &newMsg})
						if err != nil {
							println("Gossiper Encode Error: " + err.Error())
						}

						nextHop := RoutingTable.Table[newMsg.Destination]
						sendPacket(newEncoded, nextHop, gossip)

						go downloadCountDown(hex.EncodeToString(msg.HashValue), nextChunkHash, newMsg, gossip)
					}

					Files[fileInfo.metahash] = *fileInfo
				}
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
			nextHop := RoutingTable.Table[msg.Destination]
			sendPacket(newEncoded, nextHop, gossip)
		}
	}

}

func getDataFor(hash []byte) ([]byte, bool) {

	key := hex.EncodeToString(hash)

	metafileToSend, isMeta := Files[key]

	if isMeta {
		fmt.Println("TEST found meta")
		return metafileToSend.metafile, true

	} else {
		//chunk requested
		for _, fileInfo := range Files {
			index, isHere := fileInfo.orderedHashes[key]
			if isHere {
				fmt.Println("TEST found chunk")
				return fileInfo.orderedChunks[index], true
			}
		}
	}
	fmt.Println("TEST did not find hash")
	return []byte{}, false

}

func findFileWithHash(hash []byte) (*FileInfo, bool, bool, error) {
	hashString := hex.EncodeToString(hash)

	metafileToSend, isMeta := Files[hashString]

	if isMeta {
		return &metafileToSend, isMeta, true, nil
	}

	for _, fileInfo := range Files {
		if helper.Equal(hash, fileInfo.hashCurrentlyBeingFetched) {
			return &fileInfo, isMeta, true, nil
		}
	}
	return nil, isMeta, false, errors.New("No such hash")
}

func downloadFile(fileInfo FileInfo) {

	data := make([]byte, 0)

	for i := 0; i < len(fileInfo.orderedChunks); i++ {
		chunk := fileInfo.orderedChunks[i]
		data = append(data, chunk...)
	}

	err := ioutil.WriteFile(DOWNLOAD_FOLDER+fileInfo.filename, data, 0644)
	checkErr(err)

	fmt.Println("RECONSTRUCTED file " + fileInfo.filename)

}

func checkErr(err error) {
	if err != nil {
		println("File func error: " + err.Error())
	}
}
