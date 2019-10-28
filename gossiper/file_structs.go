package gossiper

var Files = make(map[[32]byte]FileInfo)
var DownloadsInProgress = make(map[[32]byte]DownloadInProgress)

const FILE_FOLDER = "./_SharedFiles/"
const DOWNLOAD_FOLDER = "./_Downloads/"
const CHUNK_SIZE = 8 * 1024

type FileInfo struct {
	filename string //File name on the local machine.
	filesize int    //File size in bytes.
	metafile []byte //The metafile computed as described above.
	chunks   map[[32]byte][]byte
	metahash [32]byte //The SHA-256 hash of the metafile.
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

type DownloadInProgress struct {
	metafile                  []byte
	chunks                    [][]byte
	metafileFetched           bool
	chunkIndexBeingFetched    int
	hashCurrentlyBeingFetched []byte
	nbChunks                  int
	filename                  string
}
