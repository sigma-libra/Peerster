package gossiper

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	chunker2 "github.com/restic/chunker"
	"io"
	"io/ioutil"
	"math"
)

func ReadFileIntoChunks(filename string) {

	data, err := ioutil.ReadFile(FILEFOLDER + filename)
	checkErr(err)

	nbChunks := int(math.Ceil(float64(len(data) / CHUNK_SIZE)))

	fileStruct := FileInfo{
		filename: filename,
		filesize: len(data),
		metafile: make([]byte, 32*nbChunks),
		chunks:   make([][32]byte, nbChunks),
		metahash: [32]byte{},
	}

	// create a chunker
	chunker := chunker2.New(bytes.NewReader(data), chunker2.Pol(0x3DA3358B4DC173))

	// reuse this buffer
	buf := make([]byte, 8*1024) //8kB

	for i := 0; i < nbChunks; i++ {
		chunk, err := chunker.Next(buf)
		if err != nil {
			fmt.Println(err)
		}
		chunkSha := sha256.Sum256(chunk.Data)
		for j := 0; j < len(chunkSha); j++ {
			fileStruct.metafile[i+j] = chunkSha[j]
		}
		fileStruct.chunks[i] = chunkSha
		if err == io.EOF {
			break
		}

		fmt.Printf("%d %02x\n", chunk.Length, sha256.Sum256(chunk.Data))
	}

	fileStruct.metahash = sha256.Sum256(fileStruct.metafile)

	Files[fileStruct.metahash] = fileStruct
}

func getDataFor(hash []byte) []byte {
	metahash := [32]byte{}
	for i := 0; i < 32; i++ {
		metahash[i] = hash[i]
	}

	metafileToSend, isMeta := Files[metahash]

	if isMeta {
		return metafileToSend.metafile

	} else {
		//chunk requested
		for _, fileInfo := range Files {
			for chunkIndex, chunk := range fileInfo.chunks {
				if chunk == metahash {
					fileData, err := ioutil.ReadFile(FILEFOLDER + fileInfo.filename)
					checkErr(err)
					chunker := chunker2.New(bytes.NewReader(fileData), chunker2.Pol(0x3DA3358B4DC173))

					// reuse this buffer
					buf := make([]byte, 8*1024) //8kB

					for i := 0; i < chunkIndex; i++ {
						_, _ = chunker.Next(buf)
					}
					chunkToSend, err := chunker.Next(buf)
					checkErr(err)
					return chunkToSend.Data

				}
			}
		}

	}
	return nil

}

func checkErr(err error) {
	if err != nil {
		println("File func error: " + err.Error())
	}
}
