package gossiper

import (
	"net"
	"sync"
	"sync/atomic"
)

type RumorMessage struct {
	Origin string
	ID     uint32
	Text   string
	Groups []string //project
}

type PeerStatus struct {
	Identifier string
	NextID     uint32
	Groups     []string //projet
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
	Simple      *SimpleMessage
	Rumor       *RumorMessage
	Status      *StatusPacket
	Private     *PrivateMessage
	DataRequest *DataRequest
	DataReply   *DataReply
}

type Gossiper struct {
	address         *net.UDPAddr
	conn            *net.UDPConn
	Name            string
	mu              sync.Mutex
	wantMap         map[string]PeerStatus
	groupMap        map[string][]string
	earlyMessages   map[string]map[uint32]RumorMessage
	orderedMessages map[string][]RumorMessage
	groupMessages   map[string][]RumorMessage
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
		groupMap:        make(map[string][]string),
		earlyMessages:   make(map[string]map[uint32]RumorMessage),
		orderedMessages: make(map[string][]RumorMessage),
		groupMessages: make(map[string][]RumorMessage),
	}
}

func getAndUpdateRumorID() uint32 {

	return atomic.AddUint32(&rumorID, 1)

}
