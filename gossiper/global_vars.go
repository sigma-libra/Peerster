package gossiper

var Files = make(map[string]FileInfo)

var NodeID IDStruct
var Messages = ""
var Nodes = ""
var PeerName = ""
var PeerUIPort = ""
var KnownPeers = make(map[string]bool)
var Keys = make([]string, 0)
var RoutingTable = InitRoutingTable()
var mongeringMessages = make(map[string]map[string][]uint32) // map (ip we monger to) -> (origin of mongered message) -> (ids of mongered Messages from origin)

var AntiEntropy = 10

var rumorID uint32 = 0

var RTimer = 0

