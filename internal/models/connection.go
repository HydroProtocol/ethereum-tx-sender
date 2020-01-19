package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sirupsen/logrus"
	"net/url"
)

var DB *gorm.DB

func ConnectDB(dbUrl string) error {
	if dbUrl == "" {
		return fmt.Errorf("empty DB url")
	}

	_url, err := url.Parse(dbUrl)

	if err != nil {
		logrus.Errorf("parse DB url failed %s", dbUrl)
		return err
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
		logrus.Errorf("failed to connect database args: %s err: %+v", args, err)
		return err
	}

	DB = _db
	DB.LogMode(true)
	DB.AutoMigrate(&LaunchLog{})
	DB.AutoMigrate(&LastBlockNumber{})

	return nil
}
