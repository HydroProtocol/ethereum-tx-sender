package launcher

import (
	"database/sql"
	"git.ddex.io/infrastructure/ethereum-launcher/models"
	"github.com/sirupsen/logrus"
)

func (l*launcher)getNextNonce(from string) int64 {
	l.nonceCacheMutex.Lock()
	defer l.nonceCacheMutex.Unlock()

	if _, exist := l.nonceCache[from]; !exist {
		nonce := l.loadLastNonce(from)
		l.nonceCache[from] = nonce
	}

	return l.nonceCache[from] + 1
}

func (l*launcher)deleteCachedNonce(from string) {
	l.nonceCacheMutex.Lock()
	defer l.nonceCacheMutex.Unlock()
	delete(l.nonceCache, from)
}

func (l*launcher)increaseNextNonce(from string) {
	l.nonceCacheMutex.Lock()
	defer l.nonceCacheMutex.Unlock()

	l.nonceCache[from] = l.nonceCache[from] + 1
}

func (l*launcher)loadLastNonce(from string) int64 {
	n, err := l.ethrpcClient.EthGetTransactionCount(from, "pending")

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
