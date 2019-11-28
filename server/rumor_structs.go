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

type RumorableMessage struct {
	Origin   string
	ID       uint32
	isTLC    bool
	rumorMsg *RumorMessage
	tclMsg   *TLCMessage
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
	Origin      string //TLC Ack: peer acking the msg
	ID          uint32 //TLC Ack: the​ same ID as the acked TLCMessage
	Text        string //TLC Ack: can be empty
	Destination string //TLC Ack: the source of the acked TLCMessage
	HopLimit    uint32 //TLC Ack: default 10, otherwise -hopLimit flag
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
	TLCMessage    *TLCMessage
	Ack           *TLCAck
}

type Gossiper struct {
	address               *net.UDPAddr
	conn                  *net.UDPConn
	Name                  string
	mu                    sync.Mutex
	wantMap               map[string]PeerStatus
	earlyMessages         map[string]map[uint32]RumorableMessage
	orderedMessages       map[string][]RumorableMessage
	tclAcks               map[uint32][]string
	my_time               int
	tlcSentForCurrentTime bool
	tlcBuffer             []BlockPublish
}

func NewGossiper(address, name string) *Gossiper {
	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	printerr("Gossiper Error Resolve Address", err)

	udpConn, err := net.ListenUDP("udp4", udpAddr)

	printerr("Gossiper Error Listen UDP", err)

	return &Gossiper{
		address:               udpAddr,
		conn:                  udpConn,
		Name:                  name,
		mu:                    sync.Mutex{},
		wantMap:               make(map[string]PeerStatus),
		earlyMessages:         make(map[string]map[uint32]RumorableMessage),
		orderedMessages:       make(map[string][]RumorableMessage),
		tclAcks:               make(map[uint32][]string),
		my_time:               0,
		tlcBuffer:             make([]BlockPublish, 0),
		tlcSentForCurrentTime: false,
	}
}
