package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBlockNumberDao_GetCurrentBlockNumberAndIncreaseBlockNumber(t *testing.T) {
	// docker-compose -f docker-compose.yaml down -v
	// docker-compose -f docker-compose.yaml up db ethereum-node
	_ = ConnectDB(LocalDBUrl)

	blockNumber, err := BlockNumberDao.GetCurrentBlockNumber()
  assert.EqualValues(t, 0, blockNumber)
  assert.Nil(t, err)

	err = BlockNumberDao.IncreaseBlockNumber(106)
	assert.Nil(t, err)

	blockNumber, err = BlockNumberDao.GetCurrentBlockNumber()
	assert.Nil(t, err)
	assert.EqualValues(t,106, blockNumber)

	err = BlockNumberDao.IncreaseBlockNumber(107)
	assert.Nil(t, err)

	blockNumber2, err := BlockNumberDao.GetCurrentBlockNumber()
	assert.Nil(t, err)
	assert.EqualValues(t,107, blockNumber2)
}
