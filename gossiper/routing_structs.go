package gossiper


type RoutingTableStruct struct {
	Table map[string]string
	LastMsgID map[string]uint32
}

func InitRoutingTable() RoutingTableStruct {
	return RoutingTableStruct{
		Table: make(map[string]string),
		LastMsgID: make(map[string]uint32),
	}
}


