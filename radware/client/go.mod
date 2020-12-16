module client

go 1.13

require (
	github.com/golang/protobuf v1.4.2 // indirect
	google.golang.org/grpc v1.31.1
	lbservice v0.0.0
)

replace lbservice => ../../api/lbservice
