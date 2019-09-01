package sorm

import (
	"fmt"
	"narabel/plogger"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var dbInstance *gorm.DB
var DBDebugMode = false

func New() *gorm.DB {
	if dbInstance != nil {
		return dbInstance
	}

	var err error
	dbInstance, err = NewWithError()

	if err != nil {
		panic(err)
	}

	if DBDebugMode {
		dbInstance.LogMode(true)
	}

	return dbInstance
}

func NewWithError() (*gorm.DB, error) {

	driver := os.Getenv("DB_DRIVER_STRING")
	connect := os.Getenv("DB_CONNECTION_STRING")

	isLocal := os.Getenv("ENVIROMENT") == "LOCAL"
	if isLocal {
		connect = os.Getenv("DB_CONNECTION_STRING_LOCAL")
	}

	db, err := gorm.Open(driver, connect)
	if err != nil {
		return db, fmt.Errorf("%s : %s (%s)", driver, connect, err.Error())
	}

	if isLocal {
		db.SetLogger(plogger.DefaultPlainLogger)
	}
	db.DB().SetConnMaxLifetime(time.Minute * 1)
	db.DB().SetMaxIdleConns(4)

	err = db.Exec("SET TRANSACTION ISOLATION LEVEL READ COMMITTED;").Error

	return db, err
}

func NewProd() (*gorm.DB, error) {

	driver := os.Getenv("DB_DRIVER_STRING")

	connect := os.Getenv("DB_CONNECTION_STRING_PROD")
	fmt.Println(driver, "con:", connect)

	db, err := gorm.Open(driver, connect)
	if err != nil {
		return db, fmt.Errorf("%s : %s (%s)", driver, connect, err.Error())
	}

	db.DB().SetConnMaxLifetime(time.Minute * 1)
	db.DB().SetMaxIdleConns(4)

	err = db.Exec("SET TRANSACTION ISOLATION LEVEL READ COMMITTED;").Error

	return db, err
}

func Transact(db *gorm.DB, txFunc func(*gorm.DB) error) (err error) {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit().Error
		}
	}()
	return txFunc(tx)
}
