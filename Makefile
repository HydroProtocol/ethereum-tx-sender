build-ethereum-sender:
	go mod download
	go build ./cmd/ethereum-sender -ldflags "-w -linkmode external -extldflags -static" -v -o build/bin/ethereum-sender


