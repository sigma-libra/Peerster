//Author: Sabrina Kall
package gossiper

import (
	"fmt"
	"math/rand"
)

func get_peer_with_group(msg_groups []string, gossip Gossiper, message string) string {
	//pick interested peer to send to
	var destPeer string
	foundNext := false
	var matchGroup string

	if FilterIncomingPackets && message != "" {
		fmt.Println("FIND FORWARDING PEER FOR GROUPS " + GroupsToString(msg_groups))
		for peerName, peerGroups := range gossip.groupMap {
			if peerName != gossip.Name {
				for _, peerGroup := range peerGroups {
					for _, group := range msg_groups {
						if peerGroup == group {
							destPeer = peerName
							foundNext = true
							matchGroup = peerGroup
							break
						}
					}
					if foundNext {
						break
					}
				}
				if foundNext {
					break
				}
			}
		}
	}

	var nextPeer string

	if !foundNext || !FilterIncomingPackets {
		nextPeer = Keys[rand.Intn(len(Keys))]
		if message != "" {
			fmt.Println("NEXT PEER RANDOM " + nextPeer)
		}
	} else {
		nextPeer = getNextHop(destPeer)
		fmt.Println("NEXT PEER INTERESTED " + nextPeer + "(" + destPeer + " with group " + matchGroup + ")")
	}
	return nextPeer
}

func GroupsToString(groups []string) string {
	str := ""
	for _, group := range groups {
		str += group + ", "
	}
	if len(groups) > 0 {
		str = str[:len(str)-2]
	}
	return str
}
