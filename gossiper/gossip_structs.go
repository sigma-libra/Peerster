package gossiper

import "net"

type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string
}

type PeerStatus struct {
	Identifier string
	NextID     uint32
}

type StatusPacket struct {
	Want []PeerStatus
}

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

type GossipPacket struct {
	Simple *SimpleMessage
	Rumor  *RumorMessage
	Status *StatusPacket
}

type Gossiper struct {
	address *net.UDPAddr
	conn    *net.UDPConn
	Name    string
}

func NewGossiper(address, name string) *Gossiper {
	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		print("Gossiper Error Resolve Address: " + err.Error() + "\n")
	}

	udpConn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		print("Gossiper Error Listen UDP: " + err.Error() + "\n")
	}
	return &Gossiper{
		address: udpAddr,
		conn:    udpConn,
		Name:    name}
}
