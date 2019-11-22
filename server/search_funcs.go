package server

import (
	"encoding/hex"
	"fmt"
	"github.com/dedis/protobuf"
	"strconv"
	"time"
)
import "strings"

func handleSearchRequest(msg *SearchRequest, gossiper *Gossiper) {

	//check if duplicate

	duplicate := false
	timeArrival := time.Now()
	origin := msg.Origin
	keywordString := stringify(msg.Keywords)

	searchReqestTracker.mu.Lock()
	keywordLists, containsOrigin := searchReqestTracker.messages[origin]
	if containsOrigin {
		lastTime, containsKeyString := keywordLists[keywordString]
		duplicate = containsKeyString && timeArrival.Sub(lastTime) < time.Millisecond*500
	} else {
		searchReqestTracker.messages[origin] = make(map[string]time.Time)
	}
	searchReqestTracker.messages[origin][keywordString] = timeArrival

	searchReqestTracker.mu.Unlock()

	if !duplicate {
		//search for keywords
		matches := SearchForKeywords(msg.Keywords)
		//handle matches
		if len(matches) > 0 {
			results := make([]*SearchResult, 0)
			for _, match := range matches {
				chunkMap := make([]uint64, 0)
				for i := uint64(1); i <= uint64(match.chunkIndexBeingFetched); i++ {
					chunkMap = append(chunkMap, i)
				}
				res := SearchResult{
					FileName:     match.filename,
					MetafileHash: match.metafile,
					ChunkMap:     chunkMap,
					ChunkCount:   uint64(match.nbChunks),
				}

				results = append(results, &res)
			}

			reply := SearchReply{
				Origin:      gossiper.Name,
				Destination: msg.Origin,
				HopLimit:    HOP_LIMIT - 1,
				Results:     results,
			}

			newEncoded, err := protobuf.Encode(&GossipPacket{SearchReply: &reply})

			printerr("Gossiper Encode Error", err)

			nextHop := getNextHop(reply.Destination)

			sendPacket(newEncoded, nextHop, gossiper)
		}

		msg.Budget -= 1
		SendSearchRequest(msg, gossiper)
	}
}

func stringify(keywords []string) string {
	keywordString := ""
	for _, key := range keywords {
		keywordString += key
	}
	return keywordString
}

func SearchForKeywords(keywords []string) []FileInfo {

	fileMemory.mu.Lock()
	defer fileMemory.mu.Unlock()

	matches := make([]FileInfo, 0)
	for _, file := range fileMemory.Files {
		for _, keyword := range keywords {
			if strings.Contains(file.filename, keyword) {
				matches = append(matches, file)
				break
			}
		}
	}
	return matches
}

func SendRepeatedSearchRequests(msg *SearchRequest, gossiper *Gossiper) {
	/*
		Then, the node
		subtracts 1 from the incoming request's budget, then only if the budget B is still greater than
		zero, redistributes the request to up to B of the node's neighbors, subdividing the remaining
		budget B as evenly as possible (i.e., plus-or-minus 1) among the recipients of these
		redistributed search requests.
	*/

	restartMatchCounter()

	for {

		SendSearchRequest(msg, gossiper)

		ticker := time.NewTicker(time.Duration(SEARCH_REQUEST_COUNTDOWN_TIME) * time.Second)
		<-ticker.C
		msg.Budget = msg.Budget * 2

		if getMatchCounter() >= 2 {
			fmt.Println("SEARCH FINISHED")
			break
		}
	}

}

func SendSearchRequest(msg *SearchRequest, gossiper *Gossiper) {
	if msg.Budget > MAX_SEARCH_BUDGET {
		distribution := divideBudget(msg.Budget, gossiper)
		for peer, budgetForPeer := range distribution {
			msg.Budget = budgetForPeer
			newEncoded, err := protobuf.Encode(&GossipPacket{SearchRequest: msg})
			printerr("Rumor Gossiper Error", err)
			sendPacket(newEncoded, peer, gossiper)
		}
	}

}

func divideBudget(budget uint64, gossiper *Gossiper) map[string]uint64 {
	distribution := make(map[string]uint64)
	peers := KnownPeers
	for k := range peers {
		distribution[k] = 0
	}
	for budget > 0 {
		for k := range peers {
			distribution[k] += 1
			budget -= 1
		}
	}

	return distribution

}

func handleSearchReply(msg *SearchReply, gossiper *Gossiper) {
	// not mine
	if msg.Destination != gossiper.Name {
		if msg.HopLimit > 0 {
			msg.HopLimit -= 1
			newEncoded, err := protobuf.Encode(&GossipPacket{SearchReply: msg})
			printerr("Gossiper Encode Error", err)
			nextHop := getNextHop(msg.Destination)
			sendPacket(newEncoded, nextHop, gossiper)
			return
		}
	}


	searchReplyTracker.mu.Lock()
	for _, result := range msg.Results {

		chunkString := ""
		for _, chunkNb := range result.ChunkMap {
			chunkString += strconv.FormatUint(chunkNb, 10) + ",";
		}
		fmt.Println("FOUND match "+ result.FileName + " at " + msg.Origin + " metafile=" + hex.EncodeToString(result.MetafileHash) + " chunks=" + chunkString[:-1])
		if result.ChunkCount == uint64(len(result.ChunkMap)) {
			getAndUpdateMatchCounter()
			addToMatchingFiles(result.FileName)

		}
		_, containsOrigin := searchReplyTracker.messages[result.FileName]
		if !containsOrigin {
			searchReplyTracker.messages[result.FileName] = make(map[string]*SearchResult)
		}
		searchReplyTracker.messages[result.FileName][msg.Origin] =  result
	}

	searchReplyTracker.mu.Unlock()

}
