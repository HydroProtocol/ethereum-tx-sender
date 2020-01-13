package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const LocalDBUrl = "postgres://postgres@localhost:5432/postgres?sslmode=disable"
func TestConnectDBFail(t *testing.T) {
	err := ConnectDB("")
	assert.NotNil(t,err)

	err = ConnectDB(LocalDBUrl)
	assert.NotNil(t, err)
}

func TestConnectDB(t *testing.T) {
	// docker-compose -f docker-db-eth-node.yaml down -v
	// docker-compose -f docker-compose-localhost-source.yaml up db ethereum-node
	err := ConnectDB(LocalDBUrl)
	assert.Nil(t, err)
	var ret struct{
		RetValue int
	}
	err = DB.Raw("select 1 as ret_value").Find(&ret).Error

	assert.Nil(t, err)
	assert.EqualValues(t, 1, ret.RetValue)
}
