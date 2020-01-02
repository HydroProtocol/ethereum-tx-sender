package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlockNumberDao_GetCurrentBlockNumberAndIncreaseBlockNumber(t *testing.T) {
	ConnectDB("postgres://ddex@localhost:5432/ddex?sslmode=disable")
	nextBlockNumber, err := BlockNumberDao.IncreaseBlockNumber(106)
	assert.Nil(t, err)
	assert.EqualValues(t,106, nextBlockNumber)

	blockNumber, err := BlockNumberDao.GetCurrentBlockNumber()
	assert.Nil(t, err)
	assert.EqualValues(t,106, blockNumber)

	nextBlockNumber2, err := BlockNumberDao.IncreaseBlockNumber(107)
	assert.Nil(t, err)
	assert.EqualValues(t,107, nextBlockNumber2)

	blockNumber2, err := BlockNumberDao.GetCurrentBlockNumber()
	assert.Nil(t, err)
	assert.EqualValues(t,107, blockNumber2)
}


