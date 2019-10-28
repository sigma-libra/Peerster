package gossiper

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	chunker2 "github.com/restic/chunker"
	"io"
	"io/ioutil"
	"math"
	"os"
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

	fileStruct := FileInfo{
		filename: filename,
		filesize: len(data),
		metafile: make([]byte, 32*nbChunks),
		chunks:   make(map[[32]byte][]byte),
		metahash: [32]byte{},
	}

	// create a chunker
	chunker := chunker2.New(bytes.NewReader(data), chunker2.Pol(0x3DA3358B4DC173))

	// reuse this buffer
	buf := make([]byte, CHUNK_SIZE) //8kB

	fmt.Println("Reading chunks into buffers")
	for i := 0; i < nbChunks; i++ {
		chunk, err := chunker.Next(buf)
		if err != nil {
			fmt.Println(err)
		}
		chunkSha := sha256.Sum256(chunk.Data)
		for j := 0; j < len(chunkSha); j++ {
			fileStruct.metafile[i+j] = chunkSha[j]
		}
		fileStruct.chunks[chunkSha] = chunk.Data
		fmt.Println("printed chunk " + strconv.Itoa(i))
		if err == io.EOF {
			break
		}

		fmt.Println("Metahash: " + string(fileStruct.metahash[:32]))
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
			data, isHere := fileInfo.chunks[metahash]
			if isHere {
				return data
			}
		}

	}
	return []byte{}

}

func downloadFile(progress DownloadInProgress) error{

		file, err := os.Create(DOWNLOAD_FOLDER + progress.filename)
		if err != nil {
		return err
	}
		defer file.Close()

		for _, data := range progress.chunks {
			_, err = io.WriteString(file, string(data))
			if err != nil {
				return err
			}
		}

		fmt.Println("RECONSTRUCTED file " + progress.filename)
		return file.Sync()

}

func checkErr(err error) {
	if err != nil {
		println("File func error: " + err.Error())
	}
}
