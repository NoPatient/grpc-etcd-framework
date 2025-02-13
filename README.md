# v1.0.0
This is a simple demo showing how to say "Hello World!" of GRPC.

To make this repo more guided, I list the steps of coding:
* coding .porto file
* generating .go based on .proto
* coding server and client
* running server then running client

Some notes:
before executing protoc command to generate .go, this two plugins should be installed:
```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```
Figure out meaning of parameters when you apply protoc.
```shell
protoc --go_out=. --go-grpc_out=. ./proto/service.proto
# go_out and go-grpc_out defines the target path of output file.
# `option go_package = "gen/go/v1";` in .proto defines the import path based on the target path.
# For example, in this code parameters are '.' meaning root path of project and go_package is 'gen/go/v1'
# Therefore the generated file will at ./gen/go/v1
```
# v1.1.0
Implement service registry and service discovery using etcd.

Running etcd on localhost using docker.
```shell
docker run -d \
  -p 2379:2379 \
  -p 2380:2380 \
  --name etcd \
  gcr.io/etcd-development/etcd:v3.5.0-arm64 \
  /usr/local/bin/etcd \
  --name my-etcd-1 \
  --data-dir /etcd-data \
  --listen-client-urls http://0.0.0.0:2379 \
  --advertise-client-urls http://0.0.0.0:2379 \
  --listen-peer-urls http://0.0.0.0:2380 \
  --initial-advertise-peer-urls http://0.0.0.0:2380 \
  --initial-cluster my-etcd-1=http://0.0.0.0:2380 \
  --initial-cluster-token my-etcd-token \
  --initial-cluster-state new
# Note: the version of image should be corresponding with architecture of localhost
```

Install etcd client and check.
```shell
brew install etcd  
etcdcli --version  
```

List all key valur pairs in etcd.
```shell
etcdctl --endpoints=http://localhost:2379 get --prefix ""
```


# v1.2.0
Add circuit breaker for client.

Circuit breaker is a design pattern used to detect failures and 
prevent an application from repeatedly trying to execute an operation
that's likely to fail.

The circuit breaker has three states:
1. Closed: The system is operating normally, and requests are allowed.
2. Open: The system has detected too many failures, and requests are blocked.
3. HalfOpen: The system is testing whether the underlying issue has been resolved by allowing a limited number of requests.

Logic summary:
1. Initial State: Closed (requests are allowed).
2. Failures Occur: Each failure increments the failure count. 
3. Threshold Reached: If failures exceed failureThreshold, the state changes to Open (requests are blocked). 
4. Timeout Elapses: After openTimeout, the state changes to HalfOpen (limited requests are allowed). 
5. Recovery: If a request succeeds in HalfOpen, the state returns to Closed. If it fails, the state goes back to Open.