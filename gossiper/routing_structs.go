package gossiper

var RTimer = 0

type RoutingTable struct {
	Table map[string]string
}

func InitRoutingTable() RoutingTable {
	return RoutingTable{Table: make(map[string]string)}
}
