package gossiper

//./gossiper -UIPort=10000 -gossipAddr=127.0.0.1:5000 -name=nodeA
// -peers=127.0.0.1:5001,10.1.1.7:5002 -simple

import (
	"flag"
	"fmt"
)


func main() {
	uiport := flag.String("UIPort", 
	"8080", "port for the UI client (default \"8080\")")
	gossipAddr := flag.String("gossipAddr", 
	"127.0.0.1:5000", "ip:port for the gossiper (default \"127.0.0.1:5000\")")
	name := flag.String("name", "", "name of the gossiper")
	peers := flag.String("peers", "", "comma separated list of peers of the form ip:port")
	simple := flag.Bool("simple", false, "run gossiper in simple broadcast mode")

	flag.Parse()
}



packetToSend := GossipPacket{Simple: simplemessage}

func NewGossiper(address, name string) *Gossiper {
	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	udpConn, err := net.ListenUDP("udp4", udpAddr)
	return &Gossiper{
		address:udpAddr,
		conn:udpConn,
		Name:name,}
}