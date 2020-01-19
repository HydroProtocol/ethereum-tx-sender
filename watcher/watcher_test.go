package watcher

import (
	"context"
	"database/sql"
	"git.ddex.io/infrastructure/ethereum-launcher/api"
	"git.ddex.io/infrastructure/ethereum-launcher/messages"
	"git.ddex.io/infrastructure/ethereum-launcher/signer"
	"time"

	"git.ddex.io/infrastructure/ethereum-launcher/models"
	"git.ddex.io/infrastructure/ethereum-launcher/pkm"
	"github.com/onrik/ethrpc"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"

	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

const user2 = "0x31ebd457b999bf99759602f5ece5aa5033cb56b3"
const user3 = "0x3eb06f432ae8f518a957852aa44776c234b4a84a"

func TestNewWatcher(t *testing.T) {
	// docker-compose -f docker-compose.yaml down -v
	// docker-compose -f docker-compose.yaml up db ethereum-node
	api.InitSubscribeHub()
	ethrpcClient := ethrpc.New("http://localhost:8545")
	_ = models.ConnectDB("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")

	signer.InitPKM("b7a0c9d2786fc4dd080ea5d619d36771aeb0c8c26c290afd3451b92ba2b7bc2c")

	balanceUser2, _:= ethrpcClient.EthGetBalance(user2, "latest")
	balanceUser3, _:= ethrpcClient.EthGetBalance(user3,"latest")

	assert.EqualValues(t, "8999999999999999821008", balanceUser2.String())
	assert.EqualValues(t, "8999999999999999821008", balanceUser3.String())

	nonce, err := ethrpcClient.EthGetTransactionCount("0x31ebd457b999bf99759602f5ece5aa5033cb56b3", "latest")
	assert.EqualValues(t, 4, nonce)
	assert.Nil(t, err)

	transaction := ethrpc.T{
		From:     "0x31ebd457b999bf99759602f5ece5aa5033cb56b3",
		To:       "0x3eb06f432ae8f518a957852aa44776c234b4a84a",
		Gas:      100000,
		GasPrice: big.NewInt(9000000000),
		Value:    big.NewInt(10000000000000000),
		Data:     "",
		Nonce:    nonce,
	}

	raw,err:= signer.LocalPKM.Sign(&transaction)
	assert.Nil(t, err)
	hash, err := ethrpcClient.EthSendRawTransaction(raw)
	assert.Nil(t, err)
	assert.EqualValues(t, "0xd300eccb6998e9102dc4bdb0f621c49276e7bfc267d24e3c7e802523288b71f7",hash)

	log := &models.LaunchLog{
		From:       "0x31ebd457b999bf99759602f5ece5aa5033cb56b3",
		To:         "0x3eb06f432ae8f518a957852aa44776c234b4a84a",
		Value:      decimal.New(1,18),
		GasLimit:   100000,
		GasUsed:    0,
		ExecutedAt: 0,
		Status:     messages.LaunchLogStatus_PENDING.String(),
		GasPrice:   decimal.New(9,9),
		Data:       nil,
		ItemType:   "test",
		ItemID:     uuid.NewV4().String(),
		Hash:       sql.NullString{
			String:hash,
			Valid:true,
		},
		ErrMsg:     "",
		IsUrgent:   false,
		Nonce:      sql.NullInt64{
			Int64:int64(nonce),
			Valid:true,
		},
	}

	err = models.LaunchLogDao.InsertLaunchLog(log)
	assert.Nil(t, err)


	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 10)
	watcherClient := NewWatcher(ctx, "http://localhost:8545", ethrpcClient)
	go watcherClient.StartWatcher()

	<- time.After(time.Second * 5)
	cancel()
	watcherLog := models.LaunchLogDao.FindLogByHash(hash)

	assert.EqualValues(t, messages.LaunchLogStatus_SUCCESS.String(), watcherLog.Status)
}
