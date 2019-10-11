package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
	"net"
)

func HandleSimpleMessagesFrom(gossip *Gossiper, isClient bool, name *string, gossipAddr *string) {
	for {

		pkt, _ := getAndDecodePacket(gossip)
		msg := pkt.Simple
		newOriginalName := msg.OriginalName
		newRelayPeerAddr := msg.RelayPeerAddr
		newContents := msg.Contents

		originalRelay := msg.RelayPeerAddr

		if isClient {
			fmt.Println("CLIENT MESSAGE " + msg.Contents)
			newOriginalName = *name
			newRelayPeerAddr = *gossipAddr

		} else {
			fmt.Println("SIMPLE MESSAGE origin " +
				msg.OriginalName + " from " +
				msg.RelayPeerAddr + " contents " + msg.Contents)

			AddPeer(msg.RelayPeerAddr)
			newRelayPeerAddr = *gossipAddr
		}
		fmt.Println("PEERS " + FormatPeers(KnownPeers))

		newMsg := SimpleMessage{newOriginalName, newRelayPeerAddr, newContents}

		newPacketBytes, err := protobuf.Encode(&GossipPacket{Simple: &newMsg})
		if err != nil {
			panic("Gossiper Encode Error: " + err.Error() + "\n")
		}

		for _, dst := range KnownPeers {
			if dst != originalRelay && dst != "0" && dst != "" {
				sendPacket(newPacketBytes, dst, gossip)
			}
		}

		messages = messages + newOriginalName + ": " + msg.Contents + "\n"

	}
}

func SendClientMessage(msg *string, uiport *string) {
	simplePacket := SimpleMessage{"client", clientAddr, *msg}
	packetToSend, err := protobuf.Encode(&GossipPacket{
		Simple: &simplePacket,
		Rumor:  nil,
		Status: nil,
	})
	if err != nil {
		print("Client Encode Error: " + err.Error() + "\n")
	}

	msgTest := GossipPacket{}
	err = protobuf.Decode(packetToSend, &msgTest)
	if err != nil {
		println("Client Protobuf Decode Error: " + err.Error())
	}

	clientUdpAddr, err := net.ResolveUDPAddr("udp4", clientAddr)
	gossiperUdpAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:"+*uiport)
	if err != nil {
		println("Client Resolve Addr Error: " + err.Error())
	}
	udpConn, err := net.ListenUDP("udp4", clientUdpAddr)
	if err != nil {
		print("Client ListenUDP Error: " + err.Error() + "\n")
	}
	_, err = udpConn.WriteToUDP(packetToSend, gossiperUdpAddr)
	if err != nil {
		println("Client Write To UDP: " + err.Error())
	}

	udpConn.Close()
}
