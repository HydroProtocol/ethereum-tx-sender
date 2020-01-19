build-ethereum-sender:
	go mod download
	go build -o bin/app -v -ldflags '-s -w' cmd/ethereum-tx-sender/main.go



