package models


import (
	"database/sql"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/shopspring/decimal"
)

const STATUS_CREATED = "created"
const STATUS_PENDING = "pending"
const STATUS_RETRIED = "retried"
const STATUS_SUCCESS = "success"

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

