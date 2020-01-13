package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlockNumberDao_GetCurrentBlockNumberAndIncreaseBlockNumber(t *testing.T) {
	// docker-compose -f docker-db-eth-node.yaml down -v
	// docker-compose -f docker-compose-localhost-source.yaml up db ethereum-node
	_ = ConnectDB(LocalDBUrl)
	blockNumber, err := BlockNumberDao.GetCurrentBlockNumber()
  assert.EqualValues(t, 1, blockNumber)
  assert.Nil(t, err)

	nextBlockNumber, err := BlockNumberDao.IncreaseBlockNumber(106)
	assert.Nil(t, err)
	assert.EqualValues(t,106, nextBlockNumber)

	blockNumber, err = BlockNumberDao.GetCurrentBlockNumber()
	assert.Nil(t, err)
	assert.EqualValues(t,106, blockNumber)

	nextBlockNumber2, err := BlockNumberDao.IncreaseBlockNumber(107)
	assert.Nil(t, err)
	assert.EqualValues(t,107, nextBlockNumber2)

	blockNumber2, err := BlockNumberDao.GetCurrentBlockNumber()
	assert.Nil(t, err)
	assert.EqualValues(t,107, blockNumber2)
}


