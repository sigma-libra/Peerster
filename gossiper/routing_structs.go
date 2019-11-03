package gossiper

import "sync"

type RoutingTable struct {
	mu sync.Mutex
	Table map[string]string
	LastMsgID map[string]uint32
}

func InitRoutingTable() RoutingTable {
	return RoutingTable{
		mu:        sync.Mutex{},
		Table: make(map[string]string),
		LastMsgID: make(map[string]uint32),
	}
}


