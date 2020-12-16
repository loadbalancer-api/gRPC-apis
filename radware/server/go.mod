module server

go 1.13

require (
	github.com/golang/mock v1.4.4 // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	google.golang.org/grpc v1.31.1
	google.golang.org/protobuf v1.25.0 // indirect
	lbservice v0.0.0
)

replace lbservice => ../../api/lbservice
