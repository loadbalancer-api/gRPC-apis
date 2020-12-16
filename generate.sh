#/bin/bash
if [ -z "$GO111MODULE" ]
then
export GO111MODULE=on
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
go get -u google.golang.org/grpc
export PATH="$PATH:$(go env GOPATH)/bin"
fi
protoc api/lbservice/lbservice.proto --go_out=plugins=grpc:.
