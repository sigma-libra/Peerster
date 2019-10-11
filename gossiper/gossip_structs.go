package gossiper

import (
	"net"
)

const clientAddr = "127.0.0.1:10000"


var messages = ""
var nodes = ""
var PeerName = ""
var PeerUIPort = ""
var ClientUIPort = ""

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

type GossipWaiter struct {
	dst  string
	done chan bool
	msg  RumorMessage
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
		address: udpAddr,
		conn:    udpConn,
		Name:    name}
}
