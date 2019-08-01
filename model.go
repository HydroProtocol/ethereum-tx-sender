package main

import (
	"database/sql"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"net/url"
	"os"
)

var db *gorm.DB

const STATUS_CREATED = "created"
const STATUS_PENDING = "pending"
const STATUS_RETRIED = "retried"
const STATUS_SUCCESS = "success"

type LaunchLog struct {
	gorm.Model

	Hash     sql.NullString
	From     string
	To       string
	Value    decimal.Decimal `gorm:"type:text"`
	GasLimit uint64
	Status   string
	GasPrice decimal.Decimal `gorm:"type:text"`
	Nonce    sql.NullInt64
	Data     []byte

	ItemType string
	ItemID   string
}

func getAllLogsWithStatus(status string) []*LaunchLog {
	var launchLogs []*LaunchLog
	db.Where("status = ?", status).Find(&launchLogs)
	return launchLogs
}

func connectDB() {
	dbUrl := os.Getenv("DATABASE_URL")

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
