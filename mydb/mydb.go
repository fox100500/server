package mydb

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//TExtOrder ...
type TExtOrder struct {
	//	gorm.Model
	Ticker                string
	Figi                  string
	Typed                 string
	Operation             string
	Price                 float64
	Lots                  int
	TakeprofitEnabled     bool
	TakeprofitPrice       float64
	TakeprofitLots        int
	TakeprofitEndDataTime time.Time
	StoplossEnabled       bool
	StoplossPrice         float64
	StoplossLots          int
	StopLossEndDataTime   time.Time
	TrailingstopEnabled   bool
	TrailingstopSize      float64
	OrderStartDateTime    time.Time
}

var mydb *gorm.DB

//Init ...
func Init() error {

	dbURL := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		"172.17.0.1",
		"postgres",
		"postgres",
		"openapidb",
		"5432",
		"disable",
		"Asia/Shanghai",
	)

	mydb, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dbURL,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
		return err
	}

	err = mydb.AutoMigrate(&TExtOrder{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "AutoMigrate: %v\n", err)
		os.Exit(1)
		return err
	}

	fmt.Println("Init database - done")
	log.Println("Init database - done")
	return nil
}

//Insert ...
func Insert(order TExtOrder) {
	result := mydb.Create(&order)
	log.Println(result)
}

//Delete ...
func Delete(figi string) {

	result := mydb.Where("Figi LIKE ?", figi).Delete(&TExtOrder{})
	log.Println(result)
}

//SelectAllRows ...
func SelectAllRows() []TExtOrder {

	var orders []TExtOrder

	result := mydb.Find(&orders)
	log.Println(result)
	return orders
}

/*
	layout := "2014-09-12T11:45:26.371Z"
	str := "2014-11-12T11:45:26.371Z"
	t := time.Now()
*/
