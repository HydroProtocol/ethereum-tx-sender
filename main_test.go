package main

import (
	"encoding/hex"
	"github.com/davecgh/go-spew/spew"
	"golang.org/x/crypto/sha3"
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	_ = os.Setenv("KUBE_NAMESPACE", "sandbox")
	_ = os.Setenv("KUBE_APP_NAME", "ethereum-launcher-kovan")
	_ = os.Setenv("ETCD_SERVER_ADDRESS", "http://0.0.0.0:8888")

	config = &Config{
		DatabaseURL: "postgres://david:@localhost:5432/launcher",
	}

	connectDB()

	var reloadedLog LaunchLog
	db.Model(&reloadedLog).Debug().Set("gorm:query_option", "FOR UPDATE").Where("id = ?", 1).Scan(&reloadedLog)
	//run()
}

func TestHash(t *testing.T) {
	x, _ := hex.DecodeString("f8aa820606843b9aca0082c9a794e5f527f02a688f7227850c890403fa35a9d8c50580b844a9059cbb000000000000000000000000b8b0b11883639a93287ec51da22e5a4741d381d50000000000000000000000000000000000000000000000056bc75e2d631000001ca0176c66321ae18a19c53a4391a6387b6e7a8414f7f1b0dd673a0555e869f669cba077ee83c53c94ce3af2dc6e14f81d8f4bdf68fbf5facdfd8174811924bd6a1f7a")
	spew.Dump(Keccak256(x))
}
