flix-dns-util: main.go
	CGO_ENABLED=1 go build -tags=netcgo -o $@ $^
