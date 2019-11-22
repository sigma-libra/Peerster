package server

import (
	"net"
	"sync"
)

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

type PrivateMessage struct {
	Origin      string
	ID          uint32
	Text        string
	Destination string
	HopLimit    uint32
}

type GossipPacket struct {
	Simple        *SimpleMessage
	Rumor         *RumorMessage
	Status        *StatusPacket
	Private       *PrivateMessage
	DataRequest   *DataRequest
	DataReply     *DataReply
	SearchRequest *SearchRequest
	SearchReply   *SearchReply
}

type Gossiper struct {
	address         *net.UDPAddr
	conn            *net.UDPConn
	Name            string
	mu              sync.Mutex
	wantMap         map[string]PeerStatus
	earlyMessages   map[string]map[uint32]RumorMessage
	orderedMessages map[string][]RumorMessage
}

func NewGossiper(address, name string) *Gossiper {
	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	printerr("Gossiper Error Resolve Address", err)

	udpConn, err := net.ListenUDP("udp4", udpAddr)

	printerr("Gossiper Error Listen UDP", err)

	return &Gossiper{
		address:         udpAddr,
		conn:            udpConn,
		Name:            name,
		mu:              sync.Mutex{},
		wantMap:         make(map[string]PeerStatus),
		earlyMessages:   make(map[string]map[uint32]RumorMessage),
		orderedMessages: make(map[string][]RumorMessage),
	}
}

