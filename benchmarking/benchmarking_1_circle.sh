#!/usr/bin/env bash

cd ./../

go build
cd client
go build
cd ..

RED='\033[0;31m'
NC='\033[0m'
DEBUG="false"

outputFiles=()
message=I_love_cats
grps=cats




UIPort=12345
gossipPort=5000
name='A'

# General peerster (gossiper) command
#./Peerster -UIPort=12345 -gossipPort=127.0.0.1:5001 -name=A -peers=127.0.0.1:5002 > A.out &

for i in `seq 1 10`;
do
	outFileName="benchmarking/$name.out"
	peerPort=$((($gossipPort+1)%10+5000))
	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
	if [[ "$name" == 'A' ]]
	then
		./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name --peers=$peer -rtimer=5 -groups=$grps > $outFileName &
	else
		./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -name=$name --peers=$peer -rtimer=5 > $outFileName &
	fi	
	outputFiles+=("$outFileName")
	if [[ "$DEBUG" == "true" ]] ; then
		echo "$name running at UIPort $UIPort and gossipPort $gossipPort"
	fi
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	name=$(echo "$name" | tr "A-Y" "B-Z")
done

SECONDS=0
for j in `seq 1 1000`;
do
    ./client/client -UIPort=12349 -groups=$grps -msg=$message
done
ELAPSED="Elapsed: $(($SECONDS / 3600))hrs $((($SECONDS / 60) % 60))min $(($SECONDS % 60))sec"

pkill -f Peerster



