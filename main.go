package main
//./gossiper -UIPort=10000 -gossipAddr=127.0.0.1:5000 -name=nodeA
// -peers=127.0.0.1:5001,10.1.1.7:5002 -simple

import (
	"flag"
	"github.com/dedis/protobuf"
	"net"
	"strconv"
	"strings"
)

type SimpleMessage struct {
	OriginalName  string
	RelayPeerAddr string
	Contents      string
}

type GossipPacket struct {
	Simple *SimpleMessage
}


var name *string
var gossipAddr *string
var knownPeers []string
var uiport *string

const clientAddress = "127.0.0.1"


func main() {

	uiport = flag.String("UIPort",
		"8080", "port for the UI client (default \"8080\")")
	gossipAddr = flag.String("gossipAddr",
		"127.0.0.1:5000", "ip:port for the gossiper (default \"127.0.0.1:5000\")")
	name = flag.String("name", "", "name of the gossiper")
	peers := flag.String("peers", "", "comma separated list of peers of the form ip:port")
	simple := flag.Bool("simple", false, "run gossiper in simple broadcast mode")

	flag.Parse()

	gossip := NewGossiper(*gossipAddr, *name)

	knownPeers = strings.Split(*peers, ",")

	if *simple {
		stopLoopChannel := make(chan bool, 1)
		go handleMessagesFrom(gossip, stopLoopChannel)
	}

}

func handleMessagesFrom(gossip *Gossiper, finish <-chan bool) {

	gossipedIP := gossip.address.IP.String()
	gossipedPort := strconv.Itoa(gossip.address.Port)
	fromClient := (gossipedIP == clientAddress) && (gossipedPort == *uiport)

	var originalRelay string
	var msg *GossipPacket
	for {
		select {
		case <-finish:
			gossip.conn.Close()
			return
		default:
			packetBytes := make([]byte, 1024)
			gossip.conn.ReadFromUDP(packetBytes)
			protobuf.Decode(packetBytes, msg)
			originalRelay = msg.Simple.RelayPeerAddr
			if fromClient {
				print("CLIENT MESSAGE " + msg.Simple.Contents)
				msg.Simple.OriginalName = *name
				msg.Simple.RelayPeerAddr = *gossipAddr

			} else {
				print("SIMPLE MESSAGE origin " +
					msg.Simple.OriginalName + " from " +
					msg.Simple.RelayPeerAddr + " contents " + msg.Simple.Contents)

				originalRelay := msg.Simple.RelayPeerAddr
				if !stringInSlice(originalRelay, knownPeers) {
					knownPeers = append(knownPeers, msg.Simple.RelayPeerAddr)
				}
				msg.Simple.RelayPeerAddr = *gossipAddr
			}

			packetBytes, _ = protobuf.Encode(msg)

			for _, dst := range knownPeers {
				if dst != originalRelay {
					udpAddr, _ := net.ResolveUDPAddr("udp4", dst)
					gossip.conn.WriteToUDP(packetBytes, udpAddr)
				}
			}

		}
	}
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

//packetToSend := GossipPacket{Simple: simplemessage}

type Gossiper struct {
	address *net.UDPAddr
	conn    *net.UDPConn
	Name    string
}

func NewGossiper(address, name string) *Gossiper {
	udpAddr, _ := net.ResolveUDPAddr("udp4", address)

	udpConn, _ := net.ListenUDP("udp4", udpAddr)
	return &Gossiper{
		address: udpAddr,
		conn:    udpConn,
		Name:    name}
}
