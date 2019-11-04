package gossiper

import (
	"github.com/dedis/protobuf"
	. "net"
)

func FormatPeers(peerSlice []string) string {
	peers := ""
	for _, peer := range peerSlice {
		peers = peers + peer + ","
	}
	if len(peerSlice) > 0 {
		peers = peers[:len(peers)-1]
	}
	return peers
}

func getAndDecodePacket(gossip *Gossiper) (GossipPacket, string) {

	packetBytes := make([]byte, PACKET_SIZE)
	_, sender, err := gossip.conn.ReadFromUDP(packetBytes)
	printerr("Get and Decode Error", err)

	pkt := GossipPacket{}
	protobuf.Decode(packetBytes, &pkt)
	return pkt, sender.String()
}

func getAndDecodeFromClient(gossip *Gossiper) Message {

	packetBytes := make([]byte, 1024)
	_, _, err := gossip.conn.ReadFromUDP(packetBytes)
	printerr("Get and Decode Error", err)

	pkt := Message{}
	protobuf.Decode(packetBytes, &pkt)
	return pkt
}

func sendPacket(pkt []byte, dst string, gossip *Gossiper) {
	udpAddr, err := ResolveUDPAddr("udp4", dst)
	printerr("Send Error", err)
	_, err = gossip.conn.WriteToUDP(pkt, udpAddr)
	printerr("Send Error", err)

}

func AddPeer(peer string) {
	if _, here := KnownPeers[peer]; !here {
		KnownPeers[peer] = true
		Keys = append(Keys, peer)
		nodes += peer + "\n"
	}

	if _, tracked := mongeringMessages[peer]; !tracked {
		mongeringMessages[peer] = make(map[string][]uint32)
	}
}


func printerr(errMsg string, err error) {
	if debug && err != nil {
		println(errMsg + ": " + err.Error())
	}
}