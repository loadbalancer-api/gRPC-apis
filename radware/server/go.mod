module server

go 1.13

require (
	github.com/golang/mock v1.4.4
	google.golang.org/grpc v1.34.0
	lbservice v0.0.0
)

replace lbservice => ../../api/lbservice
