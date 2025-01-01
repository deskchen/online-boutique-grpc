module github.com/appnetorg/OnlineBoutique

go 1.22.1

replace github.com/appnetorg/OnlineBoutique/services => ./services

replace github.com/appnetorg/OnlineBoutique/util => ./util

require (
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.3
	google.golang.org/grpc v1.69.2
	google.golang.org/protobuf v1.36.1
)

require (
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241015192408-796eee8c2d53 // indirect
)
