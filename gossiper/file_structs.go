package gossiper

import "encoding/hex"

type FileInfo struct {
	filename      string         //File name on the local machine.
	filesize      int            //File size in bytes.
	metafile      []byte         //The metafile computed as described above. [32 * NbChunks] byte array
	orderedHashes map[string]int //mapping of chunk hash to index
	orderedChunks map[int][]byte //list of chunks by index
	metahash      string         //The hex encoded SHA-256 hash of the metafile.

	downloadComplete          bool //if complete file on node
	downloadInterrupted       bool
	metafileFetched           bool   //if metafile on node
	chunkIndexBeingFetched    int    //highest unknown chunk
	hashCurrentlyBeingFetched []byte //hash of highest unknown chunk
	nbChunks                  int    //nb chunks in file
}

type DataRequest struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
}

type DataReply struct {
	Origin      string
	Destination string
	HopLimit    uint32
	HashValue   []byte
	Data        []byte
}

func InitFileInfo(filename string, metahash []byte) FileInfo {
	return FileInfo{
		filename:                  filename,
		filesize:                  0,
		metafile:                  nil,
		orderedHashes:             make(map[string]int),
		orderedChunks:             make(map[int][]byte),
		metahash:                  hex.EncodeToString(metahash),
		downloadComplete:          false,
		downloadInterrupted:       false,
		metafileFetched:           false,
		chunkIndexBeingFetched:    0,
		hashCurrentlyBeingFetched: metahash,
		nbChunks:                  0,
	}
}
