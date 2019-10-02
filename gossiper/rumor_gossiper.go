package gossiper

import (
	"fmt"
	"github.com/SabrinaKall/Peerster/helper"
	"github.com/dedis/protobuf"
	"math/rand"
	"net"
	"time"
)

func HandleRumorMessagesFrom(gossip *Gossiper, name string, gossipAddr string, knownPeers []string, isClient bool) {

	wantMap := make(map[string]PeerStatus)
	bufferedMsgs := make(map[string][]RumorMessage)

	for {

		pkt := getAndDecodePacket(gossip)

		isRumorPkt := isRumorPacket(pkt)

		if isRumorPkt {
			msg := pkt.Rumor

			receivedBefore := (wantMap[msg.Origin].NextID > msg.ID)

			if !receivedBefore {
				s1 := rand.NewSource(time.Now().UnixNano())
				r1 := rand.New(s1)
				randomPeer := knownPeers[r1.Intn(len(knownPeers))]
				newEncoded, err := protobuf.Encode(&GossipPacket{Rumor: msg})
				if err != nil {
					println("Gossiper Encode Error: " + err.Error())
				}

				sendPacket(newEncoded, randomPeer, gossip)

				if wantMap[msg.Origin].NextID == msg.ID {
					wantMap[msg.Origin] = PeerStatus{
						Identifier: msg.Origin,
						NextID:     msg.ID,
					}
					fastForward(wantMap, bufferedMsgs[msg.Origin])
				} else {
					bufferedMsgs[msg.Origin] = append(bufferedMsgs[msg.Origin], *msg)

				}

			}

		} else {
			msg := pkt.Status
		}


		originalRelay := msg.RelayPeerAddr

		if isClient {
			fmt.Println("CLIENT MESSAGE " + msg.Contents)
			msg.OriginalName = name
			msg.RelayPeerAddr = gossipAddr

		} else {
			fmt.Println("SIMPLE MESSAGE origin " +
				msg.OriginalName + " from " +
				msg.RelayPeerAddr + " contents " + msg.Contents)

			if !helper.StringInSlice(msg.RelayPeerAddr, knownPeers) {
				knownPeers = append(knownPeers, msg.RelayPeerAddr)
			}
			msg.RelayPeerAddr = gossipAddr
		}
		fmt.Println("PEERS\n" + formatPeers(knownPeers))

		newPacketBytes, err := protobuf.Encode(&GossipPacket{Simple: msg})
		if err != nil {
			println("Gossiper Encode Error: " + err.Error())
		}

		for _, dst := range knownPeers {
			if dst != originalRelay {
				udpAddr, _ := net.ResolveUDPAddr("udp4", dst)
				_, err = gossip.conn.WriteToUDP(newPacketBytes, udpAddr)
				if err != nil {
					println("Gossiper Write to UDP Error: " + err.Error())
				}
			}
		}

	}
}

func fastForward(statuses map[string]PeerStatus, messages []RumorMessage) (map[string]PeerStatus, []RumorMessage) {
	updated := false
	for

	return statuses, messages

}
