package gossiper

import (
	"fmt"
	"github.com/SabrinaKall/Peerster/helper"
	"github.com/dedis/protobuf"
)


func HandleSimpleMessagesFrom(gossip *Gossiper, isClient bool, name *string, gossipAddr *string, knownPeers []string, peerSharingChan chan string) {

	for {

		pkt, _ := getAndDecodePacket(gossip)
		msg := pkt.Simple
		newOriginalName := msg.OriginalName
		newRelayPeerAddr := msg.RelayPeerAddr
		newContents := msg.Contents

		originalRelay := msg.RelayPeerAddr

		if isClient {
			select {
			case newPeer := <-peerSharingChan:
				knownPeers = append(knownPeers, newPeer)

			default:

			}
			fmt.Println("CLIENT MESSAGE " + msg.Contents)
			newOriginalName = *name
			newRelayPeerAddr = *gossipAddr

		} else {
			fmt.Println("SIMPLE MESSAGE origin " +
				msg.OriginalName + " from " +
				msg.RelayPeerAddr + " contents " + msg.Contents)

			if !helper.StringInSlice(msg.RelayPeerAddr, knownPeers) {
				knownPeers = append(knownPeers, msg.RelayPeerAddr)
				peerSharingChan <- msg.RelayPeerAddr
			}
			newRelayPeerAddr = *gossipAddr
		}
		fmt.Println("PEERS\n" + FormatPeers(knownPeers))

		newMsg := SimpleMessage{newOriginalName, newRelayPeerAddr, newContents}

		newPacketBytes, err := protobuf.Encode(&GossipPacket{Simple: &newMsg})
		if err != nil {
			panic("Gossiper Encode Error: " + err.Error() + "\n")
		}

		for _, dst := range knownPeers {
			if dst != originalRelay {
				sendPacket(newPacketBytes, dst, gossip)
			}
		}

	}
}
