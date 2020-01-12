package models


import (
	"database/sql"
	"git.ddex.io/infrastructure/ethereum-launcher/messages"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/shopspring/decimal"
)

type LaunchLog struct {
	gorm.Model

	From       string          `gorm:"not null;type:text;index:idx_launch_logs_from"`
	To         string          `gorm:"not null;type:text"`
	Value      decimal.Decimal `gorm:"not null;type:text"`
	GasLimit   uint64          `gorm:"not null"`
	GasUsed    uint64          `gorm:"not null;default:0"`
	ExecutedAt uint64          `gorm:"not null;default:0"`
	Status     string          `gorm:"not null;index:idx_launch_logs_status"`
	GasPrice   decimal.Decimal `gorm:"not null;type:text"`
	Data       []byte          `gorm:"not null"`
	ItemType   string          `gorm:"not null;index:idx_launch_logs_item"`
	ItemID     string          `gorm:"not null;index:idx_launch_logs_item"`
	Hash       sql.NullString  `gorm:"unique_index"`
	ErrMsg     string          `gorm:"type:text"`
	IsUrgent   bool            `gorm:"not null;default:false"`
	Nonce      sql.NullInt64
}

type launchLogDao struct {
}

var LaunchLogDao *launchLogDao

func init(){
	LaunchLogDao = &launchLogDao{}
}

func (*launchLogDao)InsertLaunchLog(launchLog *LaunchLog) error {
	return DB.Create(launchLog).Error
}

func (*launchLogDao)InsertRetryLaunchLog(tx *gorm.DB, launchLog *LaunchLog) error {
	newLog := &LaunchLog{
		ItemType: launchLog.ItemType,
		ItemID:   launchLog.ItemID,
		Status:   messages.LaunchLogStatus_PENDING.String(),
		From:     launchLog.From,
		To:       launchLog.To,
		Value:    launchLog.Value,
		GasLimit: launchLog.GasLimit,
		Data:     launchLog.Data,
		Nonce:    launchLog.Nonce,
		Hash:     launchLog.Hash,
		GasPrice: launchLog.GasPrice,
		IsUrgent: launchLog.IsUrgent,
	}

	if err := tx.Save(newLog).Error; err != nil {
		return err
	}

	// TODO use subscribe instead
	// err = updateTransactionAndTrades(newLog)

	return nil
}

func (*launchLogDao)GetAllLogsWithStatus(status string) []*LaunchLog {
	var launchLogs []*LaunchLog
	DB.Where("status = ?", status).Find(&launchLogs)
	return launchLogs
}

func (*launchLogDao)FindLogByHash(hash string) *LaunchLog {
	var launchLog LaunchLog
	DB.Where("hash = ?", hash).First(&launchLog)
	return &launchLog
}

func (*launchLogDao)GetAddressMaxNonce(address string) int64 {
	var maxNonceInDB sql.NullInt64
	DB.Raw(`select max(nonce) from launch_logs where "from" = ?`, address).Scan(&maxNonceInDB)

	if maxNonceInDB.Valid {
		return maxNonceInDB.Int64
	}
  return 0
}
