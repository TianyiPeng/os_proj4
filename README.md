# os_proj4
project4 for operating systems course

Our source code of paxos library is in ./src/paxos, our source code for kvpaxos server and tests are in ./src/kvpaxos

Our start_server and stop_server scripts are in ./bin

In compile.sh, we do following things:
1. build kvpaxos server executable bin/server
2. build kvpaxos test executable bin/test
3. build paxos teset executable bin/paxostest
4. change start_server and stop_server scripts to executable
5. add current working path(I suppose you run compile.sh under this path where it belongs to) into GOPATH, this is to add paxos library, if it doesn't work, please manually add GOPATH 

For paxos test:
Our tests are two parts:
1. In ./src/kvpaxos/paxostest.go, just run bin/paxostest
2. In ./src/paxos/paxos.go, tests are in main function, just change paxo.go's package to main and run it

The paxos test includes:
1. statusTest: check all servers will decide the same value
2. doneTest: check server can clean the memory if we call done()
3. maximum-minimum Test: check the current maximum seq and minimum seq is correct
4. killTest: check when servers are killed, the consensus can continue to work
5. startTheSameSEQ: check the client can propose value with same seq, and consensus are not influenced
6. disorder propose Test: check the client can propose with arbitrary order of seq number, and consensus are not influenced  

For kvpaxos test:
Our kvpaxos test are in bin/test, just run bin/test
(Our test takes a bit long time(approxiamately 1min), please wait)

The kvpaxos test includes:
1. BasicTest: random 1000 requests, check whether all server store the same data
2. RequestIDTest: check whether requestid works




