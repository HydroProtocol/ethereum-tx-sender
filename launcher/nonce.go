package launcher

import (
	"database/sql"
	"git.ddex.io/infrastructure/ethereum-launcher/models"
	"github.com/onrik/ethrpc"
	"github.com/sirupsen/logrus"
	"sync"
)

var nonceCacheMutex = &sync.Mutex{}
var nonceCache = make(map[string]int64)
var ethrpcClient *ethrpc.EthRPC

func loadLastNonce(from string) int64 {
	n, err := ethrpcClient.EthGetTransactionCount(from, "pending")

	if err != nil {
		logrus.Errorf("%s load transcations count error: %+v", from, err)
	}

	nonce := int64(n) - 1

	var maxNonceInDB sql.NullInt64
	models.DB.Raw(`select max(nonce) from launch_logs where "from" = ?`, from).Scan(&maxNonceInDB)

	if !maxNonceInDB.Valid {
		return nonce
	}

	var res int64

	if nonce > maxNonceInDB.Int64 {
		res = nonce
	} else {
		res = maxNonceInDB.Int64
	}

	logrus.Infof("load last nonce for %s %d", from, res)

	return res
}

func deleteCachedNonce(from string) {
	nonceCacheMutex.Lock()
	defer nonceCacheMutex.Unlock()
	delete(nonceCache, from)
}

func getNextNonce(from string) int64 {
	nonceCacheMutex.Lock()
	defer nonceCacheMutex.Unlock()

	if _, exist := nonceCache[from]; !exist {
		nonce := loadLastNonce(from)
		nonceCache[from] = nonce
	}

	return nonceCache[from] + 1
}

func increaseNextNonce(from string) {
	nonceCacheMutex.Lock()
	defer nonceCacheMutex.Unlock()

	nonceCache[from] = nonceCache[from] + 1
}
