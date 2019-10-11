package gossiper

import (
	"fmt"
	"github.com/dedis/protobuf"
)

func HandleSimpleMessagesFrom(gossip *Gossiper, name *string, gossipAddr *string) {
	for {

		pkt, _ := getAndDecodePacket(gossip)
		msg := pkt.Simple
		newOriginalName := msg.OriginalName
		newRelayPeerAddr := msg.RelayPeerAddr
		newContents := msg.Contents

		originalRelay := msg.RelayPeerAddr

		fmt.Println("SIMPLE MESSAGE origin " +
			msg.OriginalName + " from " +
			msg.RelayPeerAddr + " contents " + msg.Contents)

		AddPeer(msg.RelayPeerAddr)
		newRelayPeerAddr = *gossipAddr
		fmt.Println("PEERS " + FormatPeers(KnownPeers))

		newMsg := SimpleMessage{newOriginalName, newRelayPeerAddr, newContents}

		newPacketBytes, err := protobuf.Encode(&GossipPacket{Simple: &newMsg})
		if err != nil {
			panic("Gossiper Encode Error: " + err.Error() + "\n")
		}

		for _, dst := range KnownPeers {
			if dst != originalRelay  {
				sendPacket(newPacketBytes, dst, gossip)
			}
		}

		messages = messages + newOriginalName + ": " + msg.Contents + "\n"

	}
}

func HandleSimpleClientMessagesFrom(gossip *Gossiper, name *string, gossipAddr *string) {
	for {

		pkt := getAndDecodeFromClient(gossip)
		text := pkt.Text

		fmt.Println("CLIENT MESSAGE " + text)

		fmt.Println("PEERS " + FormatPeers(KnownPeers))

		newMsg := SimpleMessage{*name, *gossipAddr, text}

		newPacketBytes, err := protobuf.Encode(&GossipPacket{Simple: &newMsg})
		if err != nil {
			panic("Gossiper Encode Error: " + err.Error() + "\n")
		}

		for _, dst := range KnownPeers {
			sendPacket(newPacketBytes, dst, gossip)
		}

		messages = messages + gossip.Name + ": " + text + "\n"

	}
}
