#!/bin/bash

go build -o src/kvpaxos/server src/kvpaxos/server.go
go build -o src/kvpaxos/test src/kvpaxos/test.go
go build -o src/kvpaxos/paxostest src/kvpaxos/paxostest.go
chmod +x src/kvpaxos/start_server
chmod +x src/kvpaxos/stop_server
pwd=$(pwd)
suffix="export GOPATH="
echo $suffix$pwd >> ~/.bashrc
