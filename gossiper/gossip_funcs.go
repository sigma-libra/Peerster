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

func getAndDecodePacket(gossip *Gossiper) GossipPacket {
	packetBytes := make([]byte, 1024)
	_, _, err := gossip.conn.ReadFromUDP(packetBytes)
	if err != nil {
		panic("Gossiper Read Error: " + err.Error() + "\n")
	}

	pkt := GossipPacket{}
	protobuf.Decode(packetBytes, &pkt)
	return pkt
}


func sendPacket(pkt []byte, dst string, gossip *Gossiper) {
	udpAddr, _ := net.ResolveUDPAddr("udp4", dst)
	_, err := gossip.conn.WriteToUDP(pkt, udpAddr)
	if err != nil {
		println("Gossiper Write to UDP Error: " + err.Error())
	}
}

func isRumorPacket(pkt GossipPacket) bool {
	pktHasSimpleMsg := (pkt.Simple != nil) && (pkt.Rumor == nil) && (pkt.Status == nil)
	pktHasRumorMsg := (pkt.Simple == nil) && (pkt.Rumor != nil) && (pkt.Status == nil)
	pktHasStatusMsg := (pkt.Simple == nil) && (pkt.Rumor == nil) && (pkt.Status != nil)

	if !pktHasSimpleMsg && ! pktHasRumorMsg && ! pktHasStatusMsg {
		panic("Packet has no message")
	}

	return pktHasRumorMsg
}
