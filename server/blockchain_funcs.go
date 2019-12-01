package server

import (
	"encoding/hex"
	"fmt"
	"github.com/dedis/protobuf"
	"math/rand"
	"strconv"
	"time"
)

func CreateBlock(info FileInfo) BlockPublish {
	hash, err := hex.DecodeString(info.metahash)
	printerr("HandleBlock", err)

	tx := TxPublish{
		Name:         info.filename,
		Size:         int64(info.filesize),
		MetafileHash: hash,
	}

	block := BlockPublish{
		PrevHash:    [32]byte{},
		Transaction: tx,
	}
	return block
}

func HandleBlock(block BlockPublish, gossiper Gossiper) {

	gossiper.mu.Lock()
	defer gossiper.mu.Unlock()
	msg := TLCMessage{
		Origin:      gossiper.Name,
		ID:          getAndUpdateRumorID(),
		Confirmed:   -1,
		TxBlock:     block,
		VectorClock: nil,
		Fitness:     0,
	}

	gossiper.wantMap[gossiper.Name] = PeerStatus{
		Identifier: gossiper.Name,
		NextID:     msg.ID + 1,
	}
	wrapper := RumorableMessage{
		Origin:   msg.Origin,
		ID:       msg.ID,
		isTLC:    true,
		rumorMsg: nil,
		tclMsg:   &msg,
	}

	gossiper.orderedMessages[gossiper.Name] = append(gossiper.orderedMessages[gossiper.Name], wrapper)
	gossiper.tclAcks[msg.ID] = make([]string, 1)
	gossiper.tclAcks[msg.ID][0] = gossiper.Name
	fmt.Println("UNCONFIRMED GOSSIP origin " + msg.Origin + " ID " + strconv.FormatUint(uint64(msg.ID), 10) +
		" file name " + msg.TxBlock.Transaction.Name + " size " + strconv.FormatInt(msg.TxBlock.Transaction.Size, 10) +
		" metahash " + hex.EncodeToString(msg.TxBlock.Transaction.MetafileHash))

	if Simple_File_Share || (Round_based_TLC && !gossiper.tlcSentForCurrentTime && len(gossiper.tlcBuffer) == 0) {
		gossiper.tlcSentForCurrentTime = true
		msg := wrapper.tclMsg
		newEncoded, err := protobuf.Encode(&GossipPacket{TLCMessage: msg})
		printerr("Rumor Gossiper Error", err)

		if len(KnownPeers) > 0 {
			randomPeer := Keys[rand.Intn(len(Keys))]
			sendPacket(newEncoded, randomPeer, &gossiper)
			addToMongering(randomPeer, msg.Origin, msg.ID)

			go statusCountDown(wrapper, randomPeer, &gossiper)
			go stubbornRepeat(msg, gossiper)
		}
	} else {
		gossiper.tlcBuffer = append(gossiper.tlcBuffer, wrapper.tclMsg.TxBlock)
	}

}

func stubbornRepeat(msg *TLCMessage, gossiper Gossiper) {
	for {
		ticker := time.NewTicker(time.Duration(StubbornTimeout) * time.Second)
		<-ticker.C
		if len(gossiper.tclAcks[msg.ID]) >= (N/2+1) || (Round_based_TLC && len(gossiper.tlcBuffer) > 0) {
			return
		} else {
			randomPeer := Keys[rand.Intn(len(Keys))]
			newEncoded, err := protobuf.Encode(&GossipPacket{TLCMessage: msg})
			printerr("Rumor Gossiper Error", err)
			sendPacket(newEncoded, randomPeer, &gossiper)
		}
	}
}

func MakeTClMessageForBlock(block BlockPublish, gossiper Gossiper) RumorableMessage {

	gossiper.mu.Lock()
	defer gossiper.mu.Unlock()
	msg := TLCMessage{
		Origin:      gossiper.Name,
		ID:          getAndUpdateRumorID(),
		Confirmed:   -1,
		TxBlock:     block,
		VectorClock: nil,
		Fitness:     0,
	}

	gossiper.wantMap[gossiper.Name] = PeerStatus{
		Identifier: gossiper.Name,
		NextID:     msg.ID + 1,
	}
	wrapper := RumorableMessage{
		Origin:   msg.Origin,
		ID:       msg.ID,
		isTLC:    true,
		rumorMsg: nil,
		tclMsg:   &msg,
	}

	return wrapper
}
