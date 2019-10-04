package gossiper

import (
	"github.com/dedis/protobuf"
	"net"
)

func formatPeers(peerSlice []string) string {
	peers := ""
	for _, peer := range peerSlice {
		peers = peers + peer + ","
	}
	peers = peers[:len(peers)-1]
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

func sendPacket(pkt []byte, dst string, gossip *Gossiper) {
	udpAddr, err := net.ResolveUDPAddr("udp", dst)
	if err != nil {
		println("Gossiper Funcs Resolve UDP Address Error: " + err.Error())
	}
	_, err = gossip.conn.WriteToUDP(pkt, udpAddr)
	if err != nil {
		println("Gossiper Funcs Write to UDP Error: " + err.Error())
	}
}

func isRumorPacket(pkt GossipPacket) bool {
	pktHasSimpleMsg := (pkt.Simple != nil) && (pkt.Rumor == nil) && (pkt.Status == nil)
	pktHasRumorMsg := (pkt.Simple == nil) && (pkt.Rumor != nil) && (pkt.Status == nil)
	pktHasStatusMsg := (pkt.Simple == nil) && (pkt.Rumor == nil) && (pkt.Status != nil)

	if !pktHasSimpleMsg && !pktHasRumorMsg && !pktHasStatusMsg {
		panic("Packet has no message")
	}

	return pktHasRumorMsg
}
