package gossiper

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/SabrinaKall/Peerster/helper"
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
