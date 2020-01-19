ethereum-sender:
	go mod download
	go build -ldflags "-w -linkmode external -extldflags -static" -v -o build/app

rmdb:
	psql postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable < db/migrations/0001-init.down.sql
