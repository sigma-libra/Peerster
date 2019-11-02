package gossiper

import (
	"net"
	"sync/atomic"
)

type IDStruct struct {
	Name    string
	Port    string
	Guiport string
}

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
	Simple      *SimpleMessage
	Rumor       *RumorMessage
	Status      *StatusPacket
	Private     *PrivateMessage
	DataRequest *DataRequest
	DataReply   *DataReply
}

type Gossiper struct {
	address           *net.UDPAddr
	conn              *net.UDPConn
	Name              string
	wantMap           map[string]PeerStatus
	earlyMessages     map[string]map[uint32]RumorMessage
	orderedMessages   map[string][]RumorMessage
}

func NewGossiper(address, name string) *Gossiper {
	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	if err != nil {
		println("Gossiper Error Resolve Address: " + err.Error())
	}

	udpConn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		println("Gossiper Error Listen UDP: " + err.Error())
	}

	return &Gossiper{
		address:           udpAddr,
		conn:              udpConn,
		Name:              name,
		wantMap:           make(map[string]PeerStatus),
		earlyMessages:     make(map[string]map[uint32]RumorMessage),
		orderedMessages:   make(map[string][]RumorMessage),
	}
}

func getAndUpdateRumorID() uint32 {

	return atomic.AddUint32(&rumorID, 1)

}
