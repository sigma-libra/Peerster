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

type FileStruct struct {
	filename string   //File name on the local machine.
	filesize int      //File size in bytes.
	metafile []byte   //The metafile computed as described above.
	metahash [32]byte //The SHA-256 hash of the metafile.
}

type DataRequest struct {
	Origin         string
	Destination    string
	stringHopLimit uint32
	HashValue      []byte
}

type DataReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
	Data        []byte
}

var Files = make(map[[32]byte]FileStruct)

const FILEFOLDER = "/_SharedFiles/"
const METAFILE = "MetaHash"

const CHUNK_SIZE = 8 * 1024

func ReadFileIntoChunks(filename string) {

	data, err := ioutil.ReadFile(FILEFOLDER + filename)
	checkErr(err)

	nbChunks := int(math.Ceil(float64(len(data) / CHUNK_SIZE)))

	fileStruct := FileStruct{
		filename: filename,
		filesize: len(data),
		metafile: make([]byte, 32*nbChunks),
		metahash: [32]byte{},
	}

	// create a chunker
	chunker := chunker2.New(bytes.NewReader(data), chunker2.Pol(0x3DA3358B4DC173))

	// reuse this buffer
	buf := make([]byte, 8*1024) //8kB

	for i := 0; i < nbChunks; i++ {
		chunk, err := chunker.Next(buf)
		if err != nil {
			panic(err)
		}

		chunkSha := sha256.Sum256(chunk.Data)
		for j := 0; j < len(chunkSha); j++ {
			fileStruct.metafile[i+j] = chunkSha[j]
		}
		if err == io.EOF {
			break
		}

		fmt.Printf("%d %02x\n", chunk.Length, sha256.Sum256(chunk.Data))
	}

	fileStruct.metahash = sha256.Sum256(fileStruct.metafile)

	Files[fileStruct.metahash] = fileStruct
}

func checkErr(err error) {
	if err != nil {
		println("File func error: " + err.Error())
	}
}
