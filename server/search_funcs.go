package server

import (
	"encoding/hex"
	"fmt"
	"github.com/dedis/protobuf"
	"sort"
	"strconv"
	"time"
)
import "strings"

func handleSearchRequest(msg *SearchRequest, gossiper *Gossiper) {

	//check if duplicate

	timeArrival := time.Now()
	origin := msg.Origin
	sort.Strings(msg.Keywords)
	keywordString := stringify(msg.Keywords)

	searchRequestTracker.mu.Lock()
	keywordLists, containsOrigin := searchRequestTracker.messages[origin]
	if containsOrigin {
		lastTime, containsKeyString := keywordLists[keywordString]
		duplicate := containsKeyString && timeArrival.Sub(lastTime) <= time.Millisecond*500
		if duplicate {
			searchRequestTracker.mu.Unlock()
			return
		}
	} else {
		searchRequestTracker.messages[origin] = make(map[string]time.Time)
	}

	searchRequestTracker.messages[origin][keywordString] = timeArrival
	searchRequestTracker.mu.Unlock()

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
			hash, err := hex.DecodeString(match.metahash)
			printerr("handle search request, decoding match", err)
			res := SearchResult{
				FileName:     match.filename,
				MetafileHash: hash,
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

		if reply.Destination == gossiper.Name {
			handleSearchReply(&reply, gossiper)
		} else {

			newEncoded, err := protobuf.Encode(&GossipPacket{SearchReply: &reply})
			printerr("Gossiper Encode Error", err)
			nextHop := getNextHop(reply.Destination)
			sendPacket(newEncoded, nextHop, gossiper)
		}

	}

	msg.Budget -= 1
	SendSearchRequest(msg, gossiper)
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

		handleSearchRequest(msg, gossiper)

		ticker := time.NewTicker(time.Duration(SEARCH_REQUEST_COUNTDOWN_TIME) * time.Second)
		<-ticker.C

		msg.Budget = msg.Budget * 2
		if msg.Budget >= MAX_SEARCH_BUDGET || getMatchCounter() >= MATCH_THRESHOLD {
			break
		}

	}

}

func SendSearchRequest(msg *SearchRequest, gossiper *Gossiper) {
	distribution := divideBudget(msg.Budget)
	for peer, budgetForPeer := range distribution {
		if budgetForPeer > 0 {
			msg.Budget = budgetForPeer
			newEncoded, err := protobuf.Encode(&GossipPacket{SearchRequest: msg})
			printerr("Rumor Gossiper Error", err)
			sendPacket(newEncoded, peer, gossiper)
		}
	}

}

func divideBudget(budget uint64) map[string]uint64 {
	distribution := make(map[string]uint64)
	peers := KnownPeers
	for k := range peers {
		distribution[k] = 0
	}
	distr := uint64(0)
	for distr < budget {
		for k := range peers {
			distribution[k] += 1
			distr += 1
			if distr >= budget {
				break
			}
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

	matchedNames := make(map[string]bool)
	SearchReplyTracker.Mu.Lock()
	for _, result := range msg.Results {

		chunkString := ""
		for _, chunkNb := range result.ChunkMap {
			chunkString += strconv.FormatUint(chunkNb, 10) + ","
		}
		fmt.Println("FOUND match " + result.FileName + " at " + msg.Origin + " metafile=" + hex.EncodeToString(result.MetafileHash) + " chunks=" + chunkString[:len(chunkString)-1])
		_, containsFile := SearchReplyTracker.Messages[result.FileName]
		if !containsFile {
			SearchReplyTracker.Messages[result.FileName] = make(map[string]*SearchResult)
		}
		SearchReplyTracker.Messages[result.FileName][msg.Origin] = result

		_, alreadyMatched := matchedNames[result.FileName]

		if result.ChunkCount == uint64(len(result.ChunkMap)) && !alreadyMatched {
			addToMatchingFiles(result.FileName)
			matchedNames[result.FileName] = true
			nbMatches := getAndUpdateMatchCounter()
			if nbMatches >= MATCH_THRESHOLD {
				fmt.Println("SEARCH FINISHED")
				break
			}
		}
	}

	/*for filename := range matchedNames {

		senders := searchReplyTracker.messages[filename]
		for _, reply := range senders {
			if reply.ChunkCount == uint64(len(reply.ChunkMap)) {
				hash := reply.MetafileHash
				fileInfo, _, _ := findFileWithHash(hash)
				downloadFile(*fileInfo)
				break
			}
		}
	}*/
	SearchReplyTracker.Mu.Unlock()

}
