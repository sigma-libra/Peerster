package server

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/dedis/protobuf"
	"io/ioutil"
	"math"
	"strconv"
	"time"
)

func ReadFileIntoChunks(filename string) {

	data, err := ioutil.ReadFile(FILE_FOLDER + filename)
	printerr("File func error", err)

	nbChunks := int(math.Ceil(float64(len(data)) / float64(CHUNK_SIZE)))

	if nbChunks == 0 {
		nbChunks = 1
	}

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
		if debug {
			println("Chunk" + strconv.Itoa(i) + ": " + chunkShaString)
		}
	}

	metahash := sha256.Sum256(fileInfo.metafile)
	fileInfo.metahash = hex.EncodeToString(metahash[:])

	putInFileMemory(fileInfo)

	if debug {
		println("Metahash: " + fileInfo.metahash)
	}
}

func putInFileMemory(info FileInfo) {
	fileMemory.mu.Lock()
	fileMemory.Files[info.metahash] = info
	fileMemory.mu.Unlock()
}

func getFromFileMemory(metaHashKey string) (FileInfo, bool) {
	fileMemory.mu.Lock()
	defer fileMemory.mu.Unlock()
	fileInfo, exists := fileMemory.Files[metaHashKey]
	return fileInfo, exists
}

func searchChunksFor(hash string) ([]byte, bool) {
	fileMemory.mu.Lock()
	defer fileMemory.mu.Unlock()
	for _, fileInfo := range fileMemory.Files {
		index, isHere := fileInfo.orderedHashes[hash]
		if isHere {
			chunk := fileInfo.orderedChunks[index]
			return chunk, true
		}
	}
	return []byte{}, false
}

func handleRequestMessage(msg *DataRequest, gossip *Gossiper) {

	if debug {
		println("Got request from " + msg.Origin)
	}

	if msg.Destination == gossip.Name {

		data, _ := getDataFor(msg.HashValue)

		reply := DataReply{
			Origin:      gossip.Name,
			Destination: msg.Origin,
			HopLimit:    HOP_LIMIT - 1,
			HashValue:   msg.HashValue,
			Data:        data,
		}
		newEncoded, err := protobuf.Encode(&GossipPacket{DataReply: &reply})

		printerr("Gossiper Encode Error", err)

		nextHop := getNextHop(reply.Destination)

		sendPacket(newEncoded, nextHop, gossip)

	} else {

		if msg.HopLimit > 0 {
			msg.HopLimit -= 1
			newEncoded, err := protobuf.Encode(&GossipPacket{DataRequest: msg})
			printerr("Gossiper Encode Error", err)
			nextHop := getNextHop(msg.Destination)
			sendPacket(newEncoded, nextHop, gossip)
			if debug {
				println("Passed on")
			}
		}
	}

}

func handleReplyMessage(msg *DataReply, gossip *Gossiper) {

	// not mine
	if msg.Destination != gossip.Name {
		if msg.HopLimit > 0 {
			msg.HopLimit -= 1
			newEncoded, err := protobuf.Encode(&GossipPacket{DataReply: msg})
			printerr("Gossiper Encode Error", err)
			nextHop := getNextHop(msg.Destination)
			sendPacket(newEncoded, nextHop, gossip)
			return
		}
	}

	//mine

	fileInfo, here := checkHashBeingFetched(msg.HashValue) //someone is trying to get this?

	if here && !fileInfo.downloadComplete {

		if len(msg.Data) == 0 { //next chunk not found: download interrupted - stop asking for it
			fileInfo.downloadInterrupted = true
			putInFileMemory(*fileInfo)
			return
		}

		shaCheck := sha256.Sum256(msg.Data)

		if Equal(shaCheck[:], msg.HashValue) { //check integrity of data

			meta := fileInfo.metahash == hex.EncodeToString(msg.HashValue) //is meta?

			if !fileInfo.metafileFetched && meta { //we want this meta
				fileInfo.metafile = msg.Data
				fileInfo.metafileFetched = true
				fileInfo.chunkIndexBeingFetched = 0

				fileInfo.nbChunks = int(math.Ceil(float64(float64(len(msg.Data)) / float64(32))))

				if fileInfo.nbChunks == 0 { //prevent rounding down in division
					fileInfo.nbChunks = 1
				}

			} else {
				//check not overwriting old reply
				key := hex.EncodeToString(fileInfo.hashCurrentlyBeingFetched)
				fileInfo.orderedHashes[key] = fileInfo.chunkIndexBeingFetched
				fileInfo.orderedChunks[fileInfo.chunkIndexBeingFetched] = msg.Data

				fileInfo.filesize += len(msg.Data)
				fileInfo.chunkIndexBeingFetched += 1

			}

			if fileInfo.chunkIndexBeingFetched >= fileInfo.nbChunks {
				downloadFile(*fileInfo)
				fileInfo.downloadComplete = true
				putInFileMemory(*fileInfo)
			} else {
				fmt.Println("DOWNLOADING " + fileInfo.filename + " chunk " + strconv.Itoa(fileInfo.chunkIndexBeingFetched+1) + " from " + msg.Origin)
				nextChunkHash := fileInfo.metafile[SHA_SIZE*(fileInfo.chunkIndexBeingFetched) : SHA_SIZE*(fileInfo.chunkIndexBeingFetched+1)]
				fileInfo.hashCurrentlyBeingFetched = nextChunkHash

				newMsg := DataRequest{
					Origin:      gossip.Name,
					Destination: msg.Origin,
					HopLimit:    HOP_LIMIT - 1,
					HashValue:   nextChunkHash,
				}

				newEncoded, err := protobuf.Encode(&GossipPacket{DataRequest: &newMsg})
				printerr("Gossiper Encode Error", err)

				nextHop := getNextHop(newMsg.Destination)
				sendPacket(newEncoded, nextHop, gossip)

				putInFileMemory(*fileInfo)

				go downloadCountDown(hex.EncodeToString(msg.HashValue), nextChunkHash, newMsg, gossip)
			}
		}
		//fileMemory.Files[fileInfo.metahash] = *fileInfo
	}

}

func getDataFor(hash []byte) ([]byte, bool) {

	key := hex.EncodeToString(hash)

	metafileToSend, isMeta := getFromFileMemory(key)

	if isMeta {
		return metafileToSend.metafile, true

	} else {
		chunk, isChunk := searchChunksFor(key)
		return chunk, isChunk
	}

}

func checkHashBeingFetched(hash []byte) (*FileInfo, bool) {
	fileMemory.mu.Lock()
	defer fileMemory.mu.Unlock()
	for _, fileInfo := range fileMemory.Files {
		if Equal(hash, fileInfo.hashCurrentlyBeingFetched) {
			return &fileInfo, true
		}
	}
	return nil, false
}

func checkHashBeingFetched_unsynched(hash []byte) (*FileInfo, bool) {
	for _, fileInfo := range fileMemory.Files {
		if Equal(hash, fileInfo.hashCurrentlyBeingFetched) {
			return &fileInfo, true
		}
	}
	return nil, false
}

func findFileWithHash(hash []byte) (*FileInfo, bool, bool) {
	hashString := hex.EncodeToString(hash)

	metafileToSend, isMeta := getFromFileMemory(hashString)

	if isMeta {
		return &metafileToSend, isMeta, true
	}

	fileInfo, isValidFile := checkHashBeingFetched(hash)

	return fileInfo, isMeta, isValidFile
}

func findFileWithHash_unsynched(hash []byte) (*FileInfo, bool, bool) {
	hashString := hex.EncodeToString(hash)

	metafileToSend, isMeta := fileMemory.Files[hashString]

	if isMeta {
		return &metafileToSend, isMeta, true
	}

	fileInfo, isValidFile := checkHashBeingFetched_unsynched(hash)

	return fileInfo, isMeta, isValidFile
}

func downloadFile(fileInfo FileInfo) {

	data := make([]byte, 0)

	for i := 0; i < len(fileInfo.orderedChunks); i++ {
		chunk := fileInfo.orderedChunks[i]
		data = append(data, chunk...)
	}

	err := ioutil.WriteFile(DOWNLOAD_FOLDER+fileInfo.filename, data, 0644)
	printerr("Download Error", err)
	//err = ioutil.WriteFile(FILE_FOLDER+fileInfo.filename, data, 0644)
	//checkErr(err)

	fmt.Println("RECONSTRUCTED file " + fileInfo.filename)

}

func downloadCountDown(key string, hash []byte, msg DataRequest, peerGossiper *Gossiper) {

	ticker := time.NewTicker(DOWNLOAD_COUNTDOWN_TIME * time.Second)
	<-ticker.C

	fileInfo, isMeta, found := findFileWithHash(hash)
	if found && !fileInfo.downloadComplete && !fileInfo.downloadInterrupted {

		newEncoded, err := protobuf.Encode(&GossipPacket{DataRequest: &msg})
		printerr("Gossiper Encode Error", err)

		if isMeta {
			fmt.Println("DOWNLOADING metafile of " + fileInfo.filename + " from " + msg.Destination)
		} else {
			fmt.Println("DOWNLOADING " + fileInfo.filename + " chunk " + strconv.Itoa(fileInfo.chunkIndexBeingFetched+1) + " from " + msg.Destination)
		}

		nextHop := getNextHop(msg.Destination)
		sendPacket(newEncoded, nextHop, peerGossiper)
		go downloadCountDown(key, hash, msg, peerGossiper)
	}

}

func Equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
