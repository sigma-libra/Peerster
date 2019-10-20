package gossiper

import "github.com/dedis/protobuf"

func initNode(name string) {

	_, wantsKnownForSender := wantMap[name]
	if !wantsKnownForSender {
		wantMap[name] = PeerStatus{
			Identifier: name,
			NextID:     1,
		}
	}

	_, listExists := orderedMessages[name]
	if !listExists {
		orderedMessages[name] = make([]RumorMessage, 0)
	}

	_, listExists = earlyMessages[name]
	if !listExists {
		earlyMessages[name] = make(map[uint32]RumorMessage)
	}

}

func getMessage(origin string, id uint32) RumorMessage {
	isInOrdered := (wantMap[origin].NextID > id)
	if isInOrdered {
		return orderedMessages[origin][id-1]
	}

	return earlyMessages[origin][id]

}

func addToMongering(dst string, origin string, ID uint32) {
	_, wasMongering := mongeringMessages[dst][origin]
	if !wasMongering {
		mongeringMessages[dst][origin] = make([]uint32, 0)
	}
	mongeringMessages[dst][origin] = append(mongeringMessages[dst][origin], ID)
}

func makeStatusPacket() []byte {
	wants := make([]PeerStatus, 0)
	for _, status := range wantMap {
		wants = append(wants, status)
	}
	wantPacket := StatusPacket{Want: wants}
	newEncoded, err := protobuf.Encode(&GossipPacket{Status: &wantPacket})
	if err != nil {
		println("Gossiper Encode Error: " + err.Error())
	}
	return newEncoded

}

func fastForward(origin string) {
	currentNext := wantMap[origin].NextID
	updated := false
	indexesDelivered := make([]uint32, 0)
	for {
		for id, savedMsg := range earlyMessages[origin] {
			if savedMsg.ID == currentNext {
				currentNext += 1
				updated = true
				indexesDelivered = append(indexesDelivered, id)
				orderedMessages[origin] = append(orderedMessages[origin], savedMsg)
			}

		}
		if !updated {
			break
		} else {
			updated = false
		}
	}
	for _, index := range indexesDelivered {
		delete(earlyMessages[origin], index)
	}
	wantMap[origin] = PeerStatus{
		origin,
		currentNext,
	}

}
