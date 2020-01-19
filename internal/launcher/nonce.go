package launcher

import (
	models2 "git.ddex.io/infrastructure/ethereum-launcher/internal/models"
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
		logrus.Errorf("%s load transactions count error: %+v", from, err)
	}

	nonce := int64(n) - 1

	maxNonceInDB := models2.LaunchLogDao.GetAddressMaxNonce(from)

	var res int64

	if nonce > maxNonceInDB {
		res = nonce
	} else {
		res = maxNonceInDB
	}

	logrus.Infof("load last nonce for %s %d", from, res)

	return res
}
