package client

//go:generate protoc -I ./../../proto --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. ../../proto/anonymizer.proto
