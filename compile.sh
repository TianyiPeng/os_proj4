#!/bin/bash

go build -o bin/server src/kvpaxos/server.go
go build -o bin/test src/kvpaxos/test.go
go build -o bin/paxostest src/kvpaxos/paxostest.go
chmod +x bin/start_server
chmod +x bin/stop_server
pwd=$(pwd)
suffix="export GOPATH="
echo $suffix$pwd >> ~/.bashrc
