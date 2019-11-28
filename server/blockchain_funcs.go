package server

import (
	"encoding/hex"
	"fmt"
	"github.com/dedis/protobuf"
	"math/rand"
	"strconv"
)

func CreateTLCMessage(info FileInfo, gossiper Gossiper) {

	/*fileMetaHash, err := hex.DecodeString(info.metahash)
	printerr("BroadcastTCLMessage: metahash string -> []byte", err)
	tx := TxPublish{
		Name:         info.filename,
		Size:         int64(info.filesize),
		MetafileHash: fileMetaHash,
	}

	block := BlockPublish{
		PrevHash:    [32]byte{},
		Transaction: TxPublish{},
	}*/

	gossiper.mu.Lock()
	msg := TLCMessage{
		Origin:      gossiper.Name,
		ID:          getAndUpdateRumorID(),
		Confirmed:   -1,
		TxBlock:     BlockPublish{},
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
	gossiper.mu.Unlock()
	fmt.Println("UNCONFIRMED GOSSIP origin " + msg.Origin + " ID " + strconv.FormatUint(uint64(msg.ID), 10) +
		" file name " + msg.TxBlock.Transaction.Name + " size " + strconv.FormatInt(msg.TxBlock.Transaction.Size, 10) +
		" metahash " + hex.EncodeToString(msg.TxBlock.Transaction.MetafileHash))

	newEncoded, err := protobuf.Encode(&GossipPacket{TLCMessage: &msg})
	printerr("Rumor Gossiper Error", err)

	if len(KnownPeers) > 0 {
		randomPeer := Keys[rand.Intn(len(Keys))]
		sendPacket(newEncoded, randomPeer, &gossiper)
		addToMongering(randomPeer, msg.Origin, msg.ID)

		go statusCountDown(wrapper, randomPeer, &gossiper)
	}

}

func handleTLCMessage(msg *TLCMessage, gossiper *Gossiper) {
	//check if filename is valid

	if msg.Confirmed != -1 {
		fmt.Println("CONFIRMED GOSSIP origin " + msg.Origin + " ID " + strconv.FormatUint(uint64(msg.ID), 10) +
			" file name " + msg.TxBlock.Transaction.Name + " size " + strconv.FormatInt(msg.TxBlock.Transaction.Size, 10) +
			" metahash " + hex.EncodeToString(msg.TxBlock.Transaction.MetafileHash))
	} else {

		fmt.Println("SENDING ACK origin " + msg.Origin + " ID " + strconv.FormatUint(uint64(msg.ID), 10))
		//send ack
		ack := PrivateMessage{
			Origin:      gossiper.Name,
			ID:          msg.ID,
			Text:        "",
			Destination: msg.Origin,
			HopLimit:    Hoplimit,
		}

		newPacketBytes, err := protobuf.Encode(&GossipPacket{Private: &ack})
		printerr("TLC Message ack protobuf encoding", err)

		sendPacket(newPacketBytes, ack.Destination, gossiper)
	}

}
