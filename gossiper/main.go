package gossiper

//./gossiper -UIPort=10000 -gossipAddr=127.0.0.1:5000 -name=nodeA
// -peers=127.0.0.1:5001,10.1.1.7:5002 -simple

import (
	"flag"
	"net"
	"strconv"
	"sync"
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

	clientAddress := net.ParseIP("127.0.0.1")
	clientPort, _ := strconv.Atoi(*uiport)

	gossiper := NewGossiper(*gossipAddr, *name)

	var wg sync.WaitGroup

	channelForMessages, _, _ := InitListening(&wg, gossiper.conn)

	msg := <-channelForMessages

	if gossiper.address.IP.Equal(clientAddress) && gossiper.address.Port == clientPort {
		print("CLIENT MESSAGE " + msg.Simple.Contents)
		/*If it comes from a client, the gossiper sends the message to all known peers, ​ sets
		the OriginalName of the message to its own name and sets relay peer to its
		own address*/
	} else {
		print("SIMPLE MESSAGE origin " +
			msg.Simple.OriginalName + " from " +
			msg.Simple.RelayPeerAddr + " contents " + msg.Simple.Contents)
		/*If it comes from another peer A, the gossiper (1) stores ​ A’s address from the relay
		peer field​ in the list of known peers, and (2) ​ changes the relay peer field to its
		own address​ , and (3) sends the message to all known peers ​ besides peer A,
		leaving the original sender field unchanged
		*/

	}

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

func getPeers(peers string) {
}
