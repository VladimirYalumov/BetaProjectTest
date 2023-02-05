package orm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"regexp"
)

var db *gorm.DB

const PrefixRecordNotFound = "record not found"

type Model interface {
	FindBy(field string, value any) (bool, error)
	Create() error
	Delete() error
}

func InitDB(host string, user string, password string, dbName string, port string) {
	var err error
	dbinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)
	db, err = gorm.Open("postgres", dbinfo)
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&Vote{})
}

func IfEmpty(err string) bool {
	matched, _ := regexp.MatchString(PrefixRecordNotFound, err)
	if matched {
		return true
	}
	return false
}
