#!/bin/bash

COUNTER=$1
DEGREE=$2

# initialize the ip file
if [ ! -d "metadata" ]; then
  mkdir metadata
fi
cd metadata
rm ip_list
export IP_PATH=$(pwd)
if [ -d "ip_list" ]; then
  rm ip_list
fi
for i in `seq 0 $COUNTER`
do
  port=$(($i+11000))
  echo 127.0.0.1:$port >> ip_list
done
cd ..

# start a thread representing bulletinboard
go run ../networking/test/bulletinboard.go -d $DEGREE -c $COUNTER -path $IP_PATH &

# start threads representing nodes
for i in `seq 1 $COUNTER`;
do
  go run ../networking/test/nodes.go -l $i -c $COUNTER -d $DEGREE -path $IP_PATH &
done

# wait some time for all the nodes to finish initializing
sleep 6

# send the clock message to bulletinboard to start an epoch
go run ../networking/test/clock.go -path $IP_PATH

# wait some time for the protocol to finish running
# LASTPORT=$(($COUNTER + 11000))
# for i in `seq 11000 $LASTPORT`;
# do
#   lsof -t -i tcp:$i | xargs kill
# done
