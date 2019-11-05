package gossiper


import "sync"

//var Files = make(map[string]FileInfo)
var fileMemory = FileMemory{
	mu:    sync.Mutex{},
	Files: make(map[string]FileInfo),
}

var mongerer = Mongering{
	mu:                sync.Mutex{},
	mongeringMessages: make(map[string]map[string][]uint32),
}

var NodeID IDStruct
var messages = ""
var nodes = ""
var PeerName = ""
var PeerUIPort = ""
var KnownPeers = make(map[string]bool)
var Keys = make([]string, 0)
var routingTable = InitRoutingTable()
//var mongeringMessages = make(map[string]map[string][]uint32) // map (ip we monger to) -> (origin of mongered message) -> (ids of mongered messages from origin)

var AntiEntropy = 10

var rumorID uint32 = 0

var RTimer = 0

type FileMemory struct {
	mu sync.Mutex
	Files map[string]FileInfo
}

type Mongering struct {
	mu sync.Mutex
	mongeringMessages map[string]map[string][]uint32

}

var debug = false


