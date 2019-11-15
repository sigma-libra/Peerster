package server

import (
	"fmt"
	"github.com/dedis/protobuf"
)

func HandleSimpleMessagesFrom(gossip *Gossiper, gossipAddr *string) {
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
		//fmt.Println("PEERS " + FormatPeers(Keys))

		newMsg := SimpleMessage{newOriginalName, newRelayPeerAddr, newContents}

		newPacketBytes, err := protobuf.Encode(&GossipPacket{Simple: &newMsg})
		if err != nil {
			printerr("Simple Gossiper Error", err)
		}

		for dst, _ := range KnownPeers {
			if dst != originalRelay  {
				sendPacket(newPacketBytes, dst, gossip)
			}
		}

		messages = messages + newOriginalName + ": " + msg.Contents + "\n"

	}
}

func HandleSimpleClientMessagesFrom(gossip *Gossiper, gossipAddr *string, peerGossip *Gossiper) {
	for {

		pkt := getAndDecodeFromClient(gossip)
		text := pkt.Text

		fmt.Println("CLIENT MESSAGE " + text)

		//fmt.Println("PEERS " + FormatPeers(Keys))

		newMsg := SimpleMessage{peerGossip.Name, *gossipAddr, text}

		newPacketBytes, err := protobuf.Encode(&GossipPacket{Simple: &newMsg})
		if err != nil {
			printerr("Simple Gossiper Error", err)
		}

		for dst, _ := range KnownPeers {
			sendPacket(newPacketBytes, dst, peerGossip)
		}

		messages = messages + gossip.Name + ": " + text + "\n"

	}
}
