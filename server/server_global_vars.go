package server

import (
	"sync"
	"sync/atomic"
	"time"
)

var Simple_File_Share bool
var Round_based_TLC bool
var Ex4 bool
var AckAll bool

var PeerGossiper *Gossiper
//var Files = make(map[string]FileInfo)
var fileMemory = FileMemory{
	mu:    sync.Mutex{},
	Files: make(map[string]FileInfo),
}

var mongerer = Mongering{
	mu:                sync.Mutex{},
	mongeringMessages: make(map[string]map[string][]uint32),
}

var searchRequestTracker = SearchRequestTracking{
	mu:       sync.RWMutex{},
	messages: make(map[string]map[string]time.Time),
}

var SearchReplyTracker = SearchReplyTracking{
	Mu:       sync.Mutex{},
	Messages: make(map[string]map[string]*SearchResult),
}

var NodeID IDStruct
var messages = ""
var nodes = ""
var confirmed = ""
var rounds = ""
var matchingFiles = ""
var messageableNodes = ""
var PeerName = ""
var PeerUIPort = ""
var KnownPeers = make(map[string]bool)
var Keys = make([]string, 0)
var routingTable = InitRoutingTable()

var N = 3
var StubbornTimeout = 5
var Hoplimit = HOP_LIMIT

//var mongeringMessages = make(map[string]map[string][]uint32) // map (ip we monger to) -> (origin of mongered message) -> (ids of mongered messages from origin)

var AntiEntropy = 10

var rumorID uint32 = 0

var matchCounter uint32 = 0

var RTimer = 0

type FileMemory struct {
	mu    sync.Mutex
	Files map[string]FileInfo
}

type Mongering struct {
	mu                sync.Mutex
	mongeringMessages map[string]map[string][]uint32
}

type SearchRequestTracking struct {
	mu       sync.RWMutex
	messages map[string]map[string]time.Time //origin -> keywords as string -> time arrived

}

type SearchReplyTracking struct {
	Mu       sync.Mutex
	Messages map[string]map[string]*SearchResult //filename -> origin -> SearchResult
}

type TCLAckTracking struct {
	mu sync.Mutex
	acks map[uint32]uint32 //TLCMessage IC -> number of acks collected
}

var debug = true

func updateAndGetMatchCounter() uint32 {

	return atomic.AddUint32(&matchCounter, 1)

}

func restartMatchCounter(gossiper Gossiper) {
	atomic.StoreUint32(&matchCounter, 0)
	gossiper.mu.Lock()
	gossiper.lastMatch = make(map[string]map[string]bool)
	gossiper.mu.Unlock()

}

func getAndUpdateRumorID() uint32 {

	return atomic.AddUint32(&rumorID, 1)

}

func getMatchCounter() uint32 {
	return atomic.LoadUint32(&matchCounter)
}

func addToMessageableNodes(node string) {
	messageableNodes += "<span onclick='openMessageWindow((this.textContent || this.innerText))'>" + node + "</span>\n"
}

func addToMatchingFiles(file string) {
	matchingFiles += "<span ondblclick='downloadSelectedFile((this.textContent || this.innerText))'>" + file + "</span>\n"
}
