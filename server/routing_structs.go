package server

import "sync"

type RoutingTable struct {
	mu sync.RWMutex
	Table map[string]string
	LastMsgID map[string]uint32
}

func InitRoutingTable() RoutingTable {
	return RoutingTable{
		mu:        sync.RWMutex{},
		Table: make(map[string]string),
		LastMsgID: make(map[string]uint32),
	}
}


