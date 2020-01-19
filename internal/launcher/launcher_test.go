package launcher

import (
	"context"
	"database/sql"
	"git.ddex.io/infrastructure/ethereum-launcher/internal/api"
	"git.ddex.io/infrastructure/ethereum-launcher/internal/config"
	messages2 "git.ddex.io/infrastructure/ethereum-launcher/internal/messages"
	models2 "git.ddex.io/infrastructure/ethereum-launcher/internal/models"
	"github.com/jinzhu/gorm"
	"github.com/onrik/ethrpc"
	uuid "github.com/satori/go.uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestStartLauncher(t *testing.T) {
	// docker-compose -f docker-db-eth-node.yaml down -v
	// docker-compose -f docker-compose-localhost-source.yaml up db ethereum-node

	_ = os.Setenv("ETHEREUM_NODE_URL", "http://localhost:8545")
	_ = os.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	_ = os.Setenv("MAX_GAS_PRICE_FOR_RETRY", "50000000000")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD", "90")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD_FOR_URGENT", "60")
	_ = os.Setenv("PRIVATE_KEYS", "b7a0c9d2786fc4dd080ea5d619d36771aeb0c8c26c290afd3451b92ba2b7bc2c")

	config.InitConfig()
	api.InitSubscribeHub()
	ethrpcClient := ethrpc.New(config.Config.EthereumNodeUrl)
	_ = models2.ConnectDB("postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	log := &models2.LaunchLog{
		From:       "0x31ebd457b999bf99759602f5ece5aa5033cb56b3",
		To:         "0x3eb06f432ae8f518a957852aa44776c234b4a84a",
		Value:      decimal.New(1, 18),
		GasLimit:   100000,
		GasUsed:    0,
		ExecutedAt: 0,
		Status:     messages2.LaunchLogStatus_CREATED.String(),
		GasPrice:   decimal.New(9, 9),
		Data:       nil,
		ItemType:   "test",
		ItemID:     uuid.NewV4().String(),
		Hash: sql.NullString{
			String: "0xd300eccb6998e9102dc4bdb0f621c49276e7bfc267d24e3c7e802523288b71f7",
			Valid:  true,
		},
		ErrMsg:   "",
		IsUrgent: false,
		Nonce: sql.NullInt64{
			Int64: int64(4),
			Valid: true,
		},
	}
	err := models2.LaunchLogDao.InsertLaunchLog(log)
	assert.Nil(t,err)

	balanceUser2, _:= ethrpcClient.EthGetBalance(user2, "latest")
	assert.EqualValues(t, "8999999999999999821008", balanceUser2.String())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 20)
	go StartLauncher(ctx, ethrpcClient)
	<- time.After(time.Second * 20)
	cancel()

	balanceUser2AfterLaunch, _:= ethrpcClient.EthGetBalance(user2, "latest")
	assert.EqualValues(t, "8998999726999999821008", balanceUser2AfterLaunch.String())
}

func TestPickLaunchLogsPendingTooLongWithNoUrgent(t *testing.T) {
	_ = os.Setenv("ETHEREUM_NODE_URL", "http://localhost:8545")
	_ = os.Setenv("DATABASE_URL", "postgres://localhost:5432/launcher")
	_ = os.Setenv("MAX_GAS_PRICE_FOR_RETRY", "50")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD", "90")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD_FOR_URGENT", "60")

	config.InitConfig()

	var logs []*models2.LaunchLog
	for i := 0; i <= 10; i++ {
		// -100, -85, -70, ..., 50
		log := &models2.LaunchLog{
			Model: gorm.Model{ID: uint(i), CreatedAt: time.Now().Add(-100 * time.Second).Add(time.Duration(i*15) * time.Second)},
			Nonce: sql.NullInt64{Valid: true, Int64: int64(i)},
		}

		logs = append(logs, log)
	}

	resendingLogs := pickLaunchLogsPendingTooLong(logs)

	assert.Len(t, resendingLogs, 1)
	assert.Equal(t, uint(0), resendingLogs[0].ID)
}

func TestPickLaunchLogsPendingTooLongWithUrgent(t *testing.T) {
	_ = os.Setenv("ETHEREUM_NODE_URL", "http://localhost:8545")
	_ = os.Setenv("DATABASE_URL", "postgres://localhost:5432/launcher")
	_ = os.Setenv("MAX_GAS_PRICE_FOR_RETRY", "50")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD", "90")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD_FOR_URGENT", "60")

	config.InitConfig()

	var logs []*models2.LaunchLog
	for i := 0; i <= 10; i++ {
		// -100, -85, -70, -55, ..., 50
		log := &models2.LaunchLog{
			Model: gorm.Model{ID: uint(i), CreatedAt: time.Now().Add(-100 * time.Second).Add(time.Duration(i*15) * time.Second)},
			Nonce: sql.NullInt64{Valid: true, Int64: int64(i)},
		}

		logs = append(logs, log)
	}

	logs[2].IsUrgent = true

	resendingLogs := pickLaunchLogsPendingTooLong(logs)

	assert.Len(t, resendingLogs, 3)
	assert.Equal(t, uint(0), resendingLogs[0].ID)
	assert.Equal(t, uint(1), resendingLogs[1].ID)
	assert.Equal(t, uint(2), resendingLogs[2].ID)
}

func TestPickLaunchLogsPendingTooLongWhenNoLogs(t *testing.T) {
	_ = os.Setenv("ETHEREUM_NODE_URL", "http://localhost:8545")
	_ = os.Setenv("DATABASE_URL", "postgres://localhost:5432/launcher")
	_ = os.Setenv("MAX_GAS_PRICE_FOR_RETRY", "50")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD", "90")
	_ = os.Setenv("RETRY_PENDING_SECONDS_THRESHOLD_FOR_URGENT", "60")

	config.InitConfig()

	var logs []*models2.LaunchLog
	resendingLogs := pickLaunchLogsPendingTooLong(logs)

	assert.Len(t, resendingLogs, 0)
}

func TestNilSlice(t *testing.T) {
	nilSlice := returnNilSlice()
	assert.Nil(t, nilSlice)
	assert.Equal(t, 0, len(nilSlice))
}

func returnNilSlice() []int {
	return nil
}
