package models

type BlockNumber struct {
	BlockNumber int `gorm:"column:block_number"`
}

type SetBlockNumberRet struct {
	BlockNumber int `gorm:"column:setval"`
}

type blockNumberDao struct{
}

var BlockNumberDao *blockNumberDao

func init(){
	BlockNumberDao = &blockNumberDao{}
}

func (*blockNumberDao) GetCurrentBlockNumber() (int,error) {
	var lastBlockNumber BlockNumber
	err := DB.Raw("select last_value as block_number from block_number_serial").Find(&lastBlockNumber).Error
	return lastBlockNumber.BlockNumber, err
}

func (*blockNumberDao) IncreaseBlockNumber(blockNumber int) (int,error) {
	var setBlockNumberRet SetBlockNumberRet
	err := DB.Raw("select setval('block_number_serial', ?, true)", blockNumber).Find(&setBlockNumberRet).Error
	return setBlockNumberRet.BlockNumber, err
}
