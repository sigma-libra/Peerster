package gossiper

import (
	"fmt"
	"math/rand"
	"strconv"
)

func get_peer_with_group(msg_groups[]string, gossip Gossiper) string {
	//pick interested peer to send to
	var destPeer string
	foundNext := false

	fmt.Println("TESTING groups: " + GroupsToString(msg_groups) + " among " + strconv.Itoa(len(gossip.groupMap)))
	for peerName, peerGroups := range gossip.groupMap {
		fmt.Println("TESTING PEER " + peerName + "(" + strconv.Itoa(len(peerGroups)) + " options)")
		for _, peerGroup := range peerGroups {
			for _, group := range msg_groups {
				fmt.Println("TESTING " + peerGroup + " (local) vs " + group)
				if peerGroup == group {
					destPeer = peerName
					foundNext = true
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

	var nextPeer string

	if !foundNext || destPeer == gossip.Name {
		nextPeer = Keys[rand.Intn(len(Keys))]
		fmt.Println("NEXT PEER RANDOM " + nextPeer)
	} else {
		nextPeer = getNextHop(destPeer);
		fmt.Println("NEXT PEER INTERESTED " + nextPeer + "(" + destPeer + ")")
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
