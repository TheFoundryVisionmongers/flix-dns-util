flix-dns-util: main.go
	GODEBUG=netdns=cgo CGO_ENABLED=1 go build -o $@ $^
