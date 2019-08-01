package main

import "database/sql"

var nonceCache = make(map[string]uint64)

func loadLastNonce(from string) int64 {
	n, err := ethrpcClient.EthGetTransactionCount(from, "latest")
	nonce := int64(n)

	if err != nil {
		panic(err)
	}

	var maxNonceInDB sql.NullInt64
	db.Raw("select max(nonce) from launch_logs where from = ?", from).Scan(&maxNonceInDB)

	if !maxNonceInDB.Valid {
		return nonce
	}

	if nonce > maxNonceInDB.Int64 {
		return nonce
	} else {
		return maxNonceInDB.Int64
	}
}

func getNextNonce(from string) uint64 {
	if _, exist := nonceCache[from]; !exist {
		nonce := loadLastNonce(from)
		nonceCache[from] = nonce
	}

	return nonceCache[from] + 1
}

func increaseNextNonce(from string) {
	nonceCache[from] = nonceCache[from] + 1
}
