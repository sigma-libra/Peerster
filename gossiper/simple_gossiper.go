package gossiper

import (
	"github.com/SabrinaKall/Peerster/helper"
	"github.com/dedis/protobuf"
)

func HandleSimpleMessagesFrom(print chan string, gossip *Gossiper, outGossip *Gossiper, name string, knownPeers []string,
	isClient bool, peerSharingChan chan string) {

	gossipAddr := gossip.address.String()
	for {

		pkt, _ := getAndDecodePacket(gossip)
		msg := pkt.Simple

		originalRelay := msg.RelayPeerAddr

		if isClient {
			select {
			case newPeer := <-peerSharingChan:
				knownPeers = append(knownPeers, newPeer)
			default:
				print <- ("CLIENT MESSAGE " + msg.Contents)
				msg.OriginalName = name
				msg.RelayPeerAddr = gossipAddr
			}

		} else {
			print <- ("SIMPLE MESSAGE origin " +
				msg.OriginalName + " from " +
				msg.RelayPeerAddr + " contents " + msg.Contents)

			if !helper.StringInSlice(msg.RelayPeerAddr, knownPeers) {
				knownPeers = append(knownPeers, msg.RelayPeerAddr)
				peerSharingChan <- msg.RelayPeerAddr
			}
			msg.RelayPeerAddr = gossipAddr
		}
		print <- ("PEERS\n" + formatPeers(knownPeers))

		newPacketBytes, err := protobuf.Encode(&GossipPacket{Simple: msg})
		if err != nil {
			println("Gossiper Encode Error: " + err.Error())
		}

		for _, dst := range knownPeers {
			if dst != originalRelay {
				sendPacket(newPacketBytes, dst, outGossip)
			}
		}

	}
}
