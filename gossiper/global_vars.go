package gossiper

var Files = make(map[string]FileInfo)

var messages = ""
var nodes = ""
var PeerName = ""
var PeerUIPort = ""
var KnownPeers = make(map[string]bool)
var Keys = make([]string, 0)
var routingTable = InitRoutingTable()
var mongeringMessages = make(map[string]map[string][]uint32) // map (ip we monger to) -> (origin of mongered message) -> (ids of mongered messages from origin)

var AntiEntropy = 10

var rumorID uint32 = 0

var RTimer = 0

