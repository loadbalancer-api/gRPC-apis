# LoadBalancer module
allows applicaton to provision 3rd party LoadBalancer via gRPC APIs

## Install and have Radware VM up and has management IP

## How to install protoc
```
sudo apt install protobuf-compiler
export GO111MODULE=on
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
go get -u google.golang.org/grpc
export PATH="$PATH:$(go env GOPATH)/bin"
```
## How to compile Loadbalancer protobuf
```
./generate.sh
```

## How to compile server module for Radware
```
cd radware/server
go build server.go
```

## How to compile client module for Radware
```
cd radware/client
go build client.go
```
## How to run server unit testing code
go to the server folder
```
go test -v -vet=off
```
## How to run code coverage
```
go test -coverprofile=coverage.out
go tool cover -func=coverage.out
```

## using grpcurl to verify loadbalancer service

Need to install grpcurl on your workspace
```
wget https://github.com/fullstorydev/grpcurl/releases/download/v1.7.0/grpcurl_1.7.0_linux_x86_64.tar.gz

tar -xvf grpcurl_1.7.0_linux_x86_64.tar.gz

chmod +x grpcurl
```
Run server
```
./radware/server/server
```

```
grpcurl --plaintext localhost:8080 describe LoadBalancer.LoadBalancerService
LoadBalancer.LoadBalancerService is a service:
service LoadBalancerService {
  rpc ConfigL3InterfacesService ( .LoadBalancer.CfgL3InterfacesRequest ) returns ( .LoadBalancer.CfgL3InterfacesResponse );
  rpc ConfigL4FilterService ( .LoadBalancer.CfgL4FilterRequest ) returns ( .LoadBalancer.CfgL4FilterResponse );
  rpc CreateService ( .LoadBalancer.CreateInstanceRequest ) returns ( .LoadBalancer.CreateInstanceResponse );
  rpc DestroyService ( .LoadBalancer.DestroyInstanceRequest ) returns ( .LoadBalancer.DestroyInstanceResponse );
  rpc ProvisionEndPointService ( .LoadBalancer.ProvisionEndPointRequest ) returns ( .LoadBalancer.ProvisionEndPointResponse );
  rpc QueryInstanceService ( .LoadBalancer.QueryInstanceRequest ) returns ( .LoadBalancer.QueryInstanceResponse );
}
```
## How to build & publish loadbalancersApi container image using gradle Kotlin
```
export REPO=<docker hub repo to upload the imaeg to>
export USER=<user Id to access the docker hub repo>
export KEY=<authentication key to access docker hub>

gradle lbbuild -Pversion=v1.0
gradle lbpublish -Pversion=v1.0
```

## How to test
start conatiner 
```
docker run  -p 8080:8080 <docker image>
```

to attach
```
docker exec -it <container name> bash
```
to test using k8s
```
kubectl apply -f config/lb_server.yaml
```
and from the worker node
```
./client --ip=nodeIP --port="31000"
```
Where 192.168.0.17 is the worker node IP address "kubectl get nodes -o wide"

to test with docker 
./client 
the default port is 8080 and host is localhost

## How to connect to vdirect running in the container
from browser running on ur worker node
connect to 
```
https://<eth0 IP inside server container>:2189/
user:<vdirect user id>
password:<vdirect password>
```

