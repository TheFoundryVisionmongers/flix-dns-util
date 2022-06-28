flix-dns-util:main.go proto/flix_transfer.pb.go
	GODEBUG=netdns=cgo CGO_ENABLED=1 go build -o $@ $<

proto/flix_transfer.pb.go:proto/flix_transfer.proto
	protoc -I proto --go_out=plugins=grpc:proto proto/flix_transfer.proto