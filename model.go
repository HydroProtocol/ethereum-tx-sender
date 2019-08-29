package main

import (
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"net/url"
)

var db *gorm.DB

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

func getAllLogsWithStatus(status string) []*LaunchLog {
	var launchLogs []*LaunchLog
	db.Where("status = ?", status).Find(&launchLogs)
	return launchLogs
}

func connectDB() {
	dbUrl := config.DatabaseURL

	if dbUrl == "" {
		logrus.Fatal("empty db url")
	}

	_url, err := url.Parse(dbUrl)

	if err != nil {
		logrus.Fatalf("parse db url failed %s", dbUrl)
	}

	host := _url.Hostname()
	port := _url.Port()
	username := _url.User.Username()
	database := _url.Path[1:]
	password, _ := _url.User.Password()
	sslmode := _url.Query().Get("sslmode")

	if sslmode == "" {
		sslmode = "disable"
	}

	args := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s", host, port, username, database, sslmode)

	if password != "" {
		args = fmt.Sprintf("%s password=%s", args, password)
	}

	_db, err := gorm.Open("postgres", args)

	if err != nil {
		logrus.Fatalf("failed to connect database args: %s err: %+v", args, err)
	}

	db = _db

	db.AutoMigrate(&LaunchLog{})
}
