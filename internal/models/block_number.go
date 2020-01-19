package models

import (
	"github.com/jinzhu/gorm"
	"strings"
)

type LastBlockNumber struct {
	gorm.Model
	BlockNumber int `gorm:"column:last_block_number"`
}

type blockNumberDao struct{
}

var BlockNumberDao *blockNumberDao

func init(){
	BlockNumberDao = &blockNumberDao{}
}

func (*blockNumberDao) GetCurrentBlockNumber() (int,error) {
	var lastBlockNumber LastBlockNumber
	err := DB.Where("id = 1").First(&lastBlockNumber).Error
	if err != nil && strings.Contains(err.Error(), "record not found") {
		return 0, nil
	}
	return lastBlockNumber.BlockNumber, err
}

func (*blockNumberDao) IncreaseBlockNumber(blockNumber int) error {
	var lastBlockNumber LastBlockNumber
	err := DB.Where("id = 1").First(&lastBlockNumber).Error

	if err != nil && !strings.Contains(err.Error(), "record not found"){
		return err
	}

	if lastBlockNumber.ID == 0 {
		lastBlockNumber = LastBlockNumber{}
		lastBlockNumber.ID = 1
	}

	lastBlockNumber.BlockNumber = blockNumber
	return DB.Save(&lastBlockNumber).Error
}
