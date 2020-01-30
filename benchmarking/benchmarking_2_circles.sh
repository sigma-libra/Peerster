#!/usr/bin/env bash

cd ./../

go build
cd client
go build
cd ..

RED='\033[0;31m'
NC='\033[0m'
DEBUG="true"

outputFiles=()
message=I_love_cats
grps=cats




UIPort=12345
gossipPort=5000
name='A'

# General peerster (gossiper) command
#./Peerster -UIPort=12345 -gossipPort=127.0.0.1:5001 -name=A -peers=127.0.0.1:5002 > A.out &

for i in `seq 1 5`;
do
	outFileName="benchmarking/$name.out"
	peerPort=$((($gossipPort+1)%5+5000))
	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
	if [[ "$name" == 'B' ]]
	then
		./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -filterIn=false --peers=$peer -rtimer=5 -groups=$grps > $outFileName &
	elif [[ "$name" == 'E' ]]; then
		./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -filterIn=false --peers="127.0.0.1:5000,127.0.0.1:5005" -rtimer=5 > $outFileName &
	else
		./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name  -filterIn=false --peers=$peer -rtimer=5 > $outFileName &
	fi	
	outputFiles+=("$outFileName")
	if [[ "$DEBUG" == "true" ]] ; then
		echo "$name running at UIPort $UIPort and gossipPort $gossipPort with peer $peer"
	fi
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	name=$(echo "$name" | tr "A-Y" "B-Z")
done

for i in `seq 1 5`;
do
	outFileName="benchmarking/$name.out"
	peerPort=$((($gossipPort+1)%5 +5005))
	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
	if [[ "$name" == 'J' ]];
	then
		./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -filterIn=false --peers="127.0.0.1:5004" -rtimer=5 -groups=$grps > $outFileName &
	else
		./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name -filterIn=false --peers=$peer -rtimer=5 > $outFileName &
	fi	
	outputFiles+=("$outFileName")
	if [[ "$DEBUG" == "true" ]] ; then
		if [[ "$name" == 'J' ]];
			then
				echo "$name running at UIPort $UIPort and gossipPort $gossipPort with peer 127.0.0.1:5004"
			else
				echo "$name running at UIPort $UIPort and gossipPort $gossipPort with peer $peer"
			fi	
	fi
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	name=$(echo "$name" | tr "A-Y" "B-Z")
done

for j in `seq 1 1000`;
do
    ./client/client -UIPort=12349 -groups=$grps -msg=$message
done

pkill -f Peerster



