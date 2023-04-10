package utils

import (
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB1 *gorm.DB
var DB2 *gorm.DB
var DB *gorm.DB

// todo : replace last variable with spread notation "..."
func ConnectDatabase(DBDriver string, DBSource1 string, DBSource2 string) {

	var db1 *gorm.DB
	var err error

	if DBDriver == "postgres" {
		db1, err = gorm.Open(postgres.Open(DBSource1), &gorm.Config{})
	} else {
		db1, err = gorm.Open(mysql.Open(DBSource1), &gorm.Config{})
	}

	if err != nil {
		panic("Failed to connect to database!")
	}

	DB1 = db1

	if DBSource2 != "" {
		db2, err := gorm.Open(postgres.Open(DBSource2), &gorm.Config{})
		if err != nil {
			panic("Failed to connect to database!")
		}
		DB2 = db2
	}

	DB = db1
}
