package gossiper

import (
	"github.com/dedis/protobuf"
	"net"
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

	packetBytes := make([]byte, 1024)
	_, sender, err := gossip.conn.ReadFromUDP(packetBytes)
	if err != nil {
		print("Gossiper funcs Read Error: " + err.Error() + "\n")
	}

	pkt := GossipPacket{}
	protobuf.Decode(packetBytes, &pkt)
	return pkt, sender.String()
}

func getAndDecodeFromClient(gossip *Gossiper) (Message) {

	packetBytes := make([]byte, 1024)
	_, _, err := gossip.conn.ReadFromUDP(packetBytes)
	if err != nil {
		print("Gossiper funcs Read Error: " + err.Error() + "\n")
	}

	pkt := Message{}
	protobuf.Decode(packetBytes, &pkt)
	return pkt
}



func sendPacket(pkt []byte, dst string, gossip *Gossiper) {
	udpAddr, err := net.ResolveUDPAddr("udp4", dst)
	if err != nil {
		println("Gossiper Funcs Resolve UDP Address Error: " + err.Error())
	}
	_, err = gossip.conn.WriteToUDP(pkt, udpAddr)
	if err != nil {
		println("Gossiper Funcs Write to UDP Error: " + err.Error())
	}

}

func AddPeer(peer string) {
	if _, here := KnownPeers[peer]; !here {
		KnownPeers[peer] = true
		Keys = append(Keys, peer)
		nodes += peer +"\n"
	}

	if _, tracked := mongeringMessages[peer]; !tracked {
		mongeringMessages[peer] = make(map[string][]uint32)
	}
}


func initNode(name string) {

	_, wantsKnownForSender := wantMap[name]
	if !wantsKnownForSender {
		wantMap[name] = PeerStatus{
			Identifier: name,
			NextID:     1,
		}
	}

	_, listExists := orderedMessages[name]
	if !listExists {
		orderedMessages[name] = make([]RumorMessage, 0)
	}

	_, listExists = earlyMessages[name]
	if !listExists {
		earlyMessages[name] = make(map[uint32]RumorMessage)
	}

}

func getMessage(origin string, id uint32) RumorMessage {
	isInOrdered := (wantMap[origin].NextID > id)
	if isInOrdered {
		return orderedMessages[origin][id-1]
	}

	return earlyMessages[origin][id]

}

func addToMongering(dst string, origin string, ID uint32) {
	_, wasMongering := mongeringMessages[dst][origin]
	if !wasMongering {
		mongeringMessages[dst][origin] = make([]uint32, 0)
	}
	mongeringMessages[dst][origin] = append(mongeringMessages[dst][origin], ID)
}