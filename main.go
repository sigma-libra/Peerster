package main

//./gossiper -UIPort=10000 -gossipAddr=127.0.0.1:5000 -name=nodeA
// -peers=127.0.0.1:5001,10.1.1.7:5002 -simple

import (
	"flag"
	"github.com/dedis/protobuf"
	"net"
	"strings"
	"time"
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

	//logger := log.New(os.Stdout, "", 0)

	uiport = flag.String("UIPort",
		"8080", "port for the UI client (default \"8080\")")
	gossipAddr = flag.String("gossipAddr",
		"127.0.0.1:5000", "ip:port for the gossiper (default \"127.0.0.1:5000\")")
	name = flag.String("name", "", "name of the gossiper")
	peers := flag.String("peers", "", "comma separated list of peers of the form ip:port")
	simple := flag.Bool("simple", false, "run gossiper in simple broadcast mode")

	flag.Parse()

	peerGossip := NewGossiper(*gossipAddr, *name)
	clientGossip := NewGossiper(clientAddress+":"+*uiport, *name)

	knownPeers = strings.Split(*peers, ",")

	if *simple {
		go handleMessagesFrom(peerGossip, false)
		go handleMessagesFrom(clientGossip, true)
		time.Sleep(10 * time.Second)

	}

}


func handleMessagesFrom(gossip *Gossiper, isClient bool) {

	for {

		packetBytes := make([]byte, 1024)
		_, _, err := gossip.conn.ReadFromUDP(packetBytes)
		if err != nil {
			panic("Gossiper Read Error: " + err.Error() + "\n")
		}

		pkt := GossipPacket{}
		protobuf.Decode(packetBytes, &pkt)
		/*if err != nil {
			panic("Gossiper Protobuf Decode Error: " + err.Error() + "\n")
		} else {
			print("Fine\n")
		}*/

		msg := pkt.Simple
		newOriginalName := msg.OriginalName
		newRelayPeerAddr := msg.RelayPeerAddr
		newContents := msg.Contents

		originalRelay := msg.RelayPeerAddr

		if isClient {
			print("CLIENT MESSAGE "+ msg.Contents + "\n")
			newOriginalName = *name
			newRelayPeerAddr = *gossipAddr

		} else {
			print("SIMPLE MESSAGE origin " +
				msg.OriginalName + " from " +
				msg.RelayPeerAddr + " contents " + msg.Contents + "\n")

			if !stringInSlice(msg.RelayPeerAddr, knownPeers) {
				knownPeers = append(knownPeers, msg.RelayPeerAddr)
			}
			newRelayPeerAddr = *gossipAddr
		}
		print("PEERS\n" + formatPeers(knownPeers) + "\n")

		newMsg := SimpleMessage{newOriginalName, newRelayPeerAddr, newContents}


		newPacketBytes, err := protobuf.Encode(&GossipPacket{&newMsg})
		if err != nil {
			panic("Gossiper Encode Error: " + err.Error() + "\n")
		}

		for _, dst := range knownPeers {
			if dst != originalRelay {
				udpAddr, _ := net.ResolveUDPAddr("udp4", dst)
				_, err = gossip.conn.WriteToUDP(newPacketBytes, udpAddr)
				if err != nil {
					print("Gossiper Write to UDP Error: " + err.Error() + "\n")
				}
			}
		}

	}
}

func formatPeers(peerSlice[]string) string {
	peers := ""
	for _, peer := range peerSlice {
		peers = peers + peer + ","
	}
	peers = peers[:len(peers)-1]
	return peers
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
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

