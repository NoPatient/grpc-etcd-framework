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

