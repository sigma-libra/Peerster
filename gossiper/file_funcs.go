package gossiper

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/SabrinaKall/Peerster/helper"
	"github.com/dedis/protobuf"
	chunker2 "github.com/restic/chunker"
	"io"
	"io/ioutil"
	"math"
	"strconv"
)

func ReadFileIntoChunks(filename string) {

	fmt.Println("reading file " + filename + " into chunks")

	data, err := ioutil.ReadFile(FILE_FOLDER + filename)
	checkErr(err)

	nbChunks := int(math.Ceil(float64(len(data) / CHUNK_SIZE)))

	if nbChunks == 0 {
		nbChunks = 1
	}

	fmt.Println("Nb chunks: " + strconv.Itoa(nbChunks))

	fileInfo := FileInfo{
		filename:                  filename,
		filesize:                  len(data),
		metafile:                  make([]byte, 32*nbChunks),
		orderedHashes:             make(map[string]int),
		orderedChunks:             make(map[int][]byte),
		metahash:                  "",
		downloadComplete:          true,
		metafileFetched:           true,
		chunkIndexBeingFetched:    nbChunks,
		hashCurrentlyBeingFetched: nil,
		nbChunks:                  nbChunks,
	}

	// create a chunker
	chunker := chunker2.New(bytes.NewReader(data), chunker2.Pol(0x3DA3358B4DC173))

	// reuse this buffer
	buf := make([]byte, CHUNK_SIZE) //8kB

	for i := 0; i < nbChunks; i++ {
		chunk, err := chunker.Next(buf)
		if err != nil {
			fmt.Println(err)
		}
		chunkSha := sha256.Sum256(chunk.Data)
		for j := 0; j < len(chunkSha); j++ {
			fileInfo.metafile[i+j] = chunkSha[j]
		}
		chunkShaString := hex.EncodeToString(chunkSha[:])

		fileInfo.orderedHashes[chunkShaString] = i
		fileInfo.orderedChunks[i] = chunk.Data
		if err == io.EOF {
			break
		}
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
			HopLimit:    HOP_LIMIT - 1,
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

}

func handleReplyMessage(msg *DataReply, gossip *Gossiper) {
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
						HopLimit:    HOP_LIMIT - 1,
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
		return &metafileToSend, true, true, nil
	}

	for _, fileInfo := range Files {
		if helper.Equal(hash, fileInfo.hashCurrentlyBeingFetched) {
			return &fileInfo, false, true, nil
		}
	}
	return nil, false, false, errors.New("No such hash")
}

func downloadFile(fileInfo FileInfo) error {

	data := make([]byte, 0)

	for i := 0; i < len(fileInfo.orderedChunks); i++ {
		chunk := fileInfo.orderedChunks[i]
		data = append(data, chunk...)
	}

	err := ioutil.WriteFile(DOWNLOAD_FOLDER+fileInfo.filename, data, 0644)
	checkErr(err)

	fmt.Println("RECONSTRUCTED file " + fileInfo.filename)
	fileInfo.downloadComplete = true
	return nil

}

func checkErr(err error) {
	if err != nil {
		println("File func error: " + err.Error())
	}
}
