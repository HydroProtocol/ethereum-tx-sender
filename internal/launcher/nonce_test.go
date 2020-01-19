package launcher

import (
	models2 "git.ddex.io/infrastructure/ethereum-launcher/internal/models"
	"github.com/onrik/ethrpc"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

const user2 = "0x31ebd457b999bf99759602f5ece5aa5033cb56b3"
const user3 = "0x3eb06f432ae8f518a957852aa44776c234b4a84a"

func TestLoadLastNonce(t *testing.T) {
	_ = models2.ConnectDB("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	tLancher := newTestLauncher()

	nonce := tLancher.loadLastNonce(user2)
	assert.EqualValues(t, 3, nonce)
}

func TestGetNextNonce(t *testing.T) {
	_ = models2.ConnectDB("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	tLancher := newTestLauncher()

	currentNonce := tLancher.getNextNonce(user2)
	assert.EqualValues(t, 4, currentNonce)
}

func TestIncreaseNextNonce(t *testing.T) {
	_ = models2.ConnectDB("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	tLancher := newTestLauncher()

	nonce := tLancher.getNextNonce(user2)
	assert.EqualValues(t, 4, nonce)

	tLancher.increaseNextNonce(user2)
	currentNonce := tLancher.getNextNonce(user2)
	assert.EqualValues(t, 5, currentNonce)
}

func TestDeleteCachedNonce(t *testing.T) {
	_ = models2.ConnectDB("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	tLancher := newTestLauncher()

	nonce := tLancher.getNextNonce(user2)
	assert.EqualValues(t, 4, nonce)

	tLancher.increaseNextNonce(user2)
	currentNonce := tLancher.getNextNonce(user2)
	assert.EqualValues(t, 5, currentNonce)

	tLancher.deleteCachedNonce(user2)
	currentNonce = tLancher.getNextNonce(user2)
	assert.EqualValues(t, 4, currentNonce)
}

func TestConcurrency(t *testing.T) {
	_ = models2.ConnectDB("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	tLancher := newTestLauncher()

	nonce := tLancher.getNextNonce(user2)
	assert.EqualValues(t, 4, nonce)

	for i:=0; i < 20; i++ {
		go tLancher.increaseNextNonce(user2)
	}

	<-time.After(time.Second * 5)
	currentNonce := tLancher.getNextNonce(user2)
	assert.EqualValues(t, 24, currentNonce)
}

func newTestLauncher() *launcher {
	ethrpcClient := ethrpc.New("http://localhost:8545")
	tLauncher := launcher{
			ethrpcClient:ethrpcClient,
			pkm:nil,
			nonceCache:make(map[string]int64),
			nonceCacheMutex: &sync.Mutex{},
		}

	return &tLauncher
}
