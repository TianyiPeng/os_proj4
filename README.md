# os_proj4
project4 for operating systems course

Our source code of paxos library is in ./src/paxos, our source code for kvpaxos server and tests are in ./src/kvpaxos

Our start_server and stop_server scripts are in ./src/kvpaxos

In compile.sh, we do following things:
1. build kvpaxos server executable src/kvpaxos/server
2. build kvpaxos test executable src/kvpaxos/test
3. build paxos teset executable src/kvpaxos/paxostest
4. change start_server and stop_server scripts to executable
5. add current working path(I suppose you run compile.sh under this path where it belongs to) into GOPATH, this is to add paxos library, if it doesn't work, please manually add GOPATH 

For paxos test:
Our tests are two parts:
1. In ./src/kvpaxos/paxostest.go, just run paxostest
2. In ./src/paxos/paxos.go, tests are in main function, just change paxo.go's package to main and run it

For kvpaxos test:
Our kvpaxos test are in ./src/kvpaxos/test.go, just run test




