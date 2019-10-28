package gossiper

var RTimer = 0
const HopLimit = 10

type RoutingTable struct {
	Table map[string]string
	LastMsgID map[string]uint32
}

func InitRoutingTable() RoutingTable {
	return RoutingTable{
		Table: make(map[string]string),
		LastMsgID: make(map[string]uint32),
	}
}


