package server

import (
	"encoding/hex"
	"fmt"
	"github.com/dedis/protobuf"
	"strconv"
	"time"
)
import "strings"

func handleSearchRequest(msg *SearchRequest, gossiper *Gossiper, timeArrival time.Time, reduceBudget bool) {

	if debug {
		fmt.Println("Got search request")
	}
	origin := msg.Origin
	keywordString := stringify(msg.Keywords)

	if origin != gossiper.Name {

		//check if duplicate
		searchRequestTracker.mu.Lock()
		keywordLists, containsOrigin := searchRequestTracker.messages[origin]

		if !containsOrigin {
			searchRequestTracker.messages[origin] = make(map[string]time.Time)
		} else {
			lastTime, containsKeyString := keywordLists[keywordString]
			if containsKeyString {
				diff := timeArrival.Sub(lastTime).Seconds()
				if diff <= 0.5 {
					searchRequestTracker.mu.Unlock()
					if debug {
						fmt.Println("Too soon - dropping search request")
					}
					return
				}
			}
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

			newEncoded, err := protobuf.Encode(&GossipPacket{SearchReply: &reply})
			printerr("Gossiper Encode Error", err)
			nextHop := getNextHop(reply.Destination)
			sendPacket(newEncoded, nextHop, gossiper)

		}
	}

	if reduceBudget {
		msg.Budget -= 1
	}
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
				if debug {
					fmt.Println("Found match: " + keyword + " -> " + file.filename)
				}
				matches = append(matches, file)
				break
			}
		}
	}
	return matches
}

func SendRepeatedSearchRequests(msg *SearchRequest, gossiper *Gossiper, increasingBudget bool) {
	/*
		Then, the node
		subtracts 1 from the incoming request's budget, then only if the budget B is still greater than
		zero, redistributes the request to up to B of the node's neighbors, subdividing the remaining
		budget B as evenly as possible (i.e., plus-or-minus 1) among the recipients of these
		redistributed search requests.
	*/

	restartMatchCounter()

	firstTime := true

	for {

		if debug {
			fmt.Println("Looping repeated search requests: budget: ")
		}
		handleSearchRequest(msg, gossiper, time.Now(), firstTime)

		firstTime = false;
		ticker := time.NewTicker(time.Duration(SEARCH_REQUEST_COUNTDOWN_TIME) * time.Second)
		<-ticker.C

		if increasingBudget {
			msg.Budget = msg.Budget * 2
			if debug {
				fmt.Println("Searching with budget: " + strconv.FormatUint(msg.Budget, 10))
			}

		}
		if msg.Budget > MAX_SEARCH_BUDGET || getMatchCounter() >= MATCH_THRESHOLD {
			if debug {
				fmt.Println("Done repeating search with budget: " + strconv.FormatUint(msg.Budget, 10) +
					" / matches: " + strconv.FormatUint(uint64(getMatchCounter()), 10))
			}
			break
		}

	}

}

func SendSearchRequest(msg *SearchRequest, gossiper *Gossiper) {
	distribution := divideBudget(msg.Budget)
	for peer, budgetForPeer := range distribution {
		if budgetForPeer > 0 {
			if debug {
				fmt.Println("Send search request to " + peer)
			}
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

	//matchedNames := make(map[string]bool)
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

		//_, alreadyMatched := matchedNames[result.FileName]

		if result.ChunkCount == uint64(len(result.ChunkMap)) { //&& !alreadyMatched {
			addToMatchingFiles(result.FileName)
			//matchedNames[result.FileName] = true
			nbMatches := getAndUpdateMatchCounter()
			if nbMatches >= MATCH_THRESHOLD {
				fmt.Println("SEARCH FINISHED")
				break
			}
		}
	}
	SearchReplyTracker.Mu.Unlock()

}
