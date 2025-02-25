package data

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
)

type Access struct {
	DSN string
	DB  *gorm.DB
}

func NewDataAccess(dsn string) *Access {
	da := &Access{
		DSN: dsn,
	}
	err := da.Connect()
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	return da
}

func NewDataAccessFromEnvDSN() *Access {
	dsn, ok := os.LookupEnv("DSN")
	if !ok {
		dsn = "host=localhost port=5432 user=postgres password=postgres1 dbname=postgres sslmode=disable"
	}

	dataAccess := NewDataAccess(dsn)
	if dataAccess == nil {
		panic(fmt.Errorf("unable to connect to database, check database is accessible and dsn: %s", dsn))
	}

	return dataAccess
}

func (da *Access) Connect() error {
	var err error
	da.DB, err = gorm.Open(postgres.Open(da.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}
	return nil
}

func (da *Access) Execute(sql string, values ...interface{}) error {
	var tx *gorm.DB
	if len(values) > 0 {
		tx = da.DB.Exec(sql, values)
	} else {
		tx = da.DB.Exec(sql)
	}

	if err := tx.Error; err != nil {
		return fmt.Errorf("unable to execute sql %s, err: %v", sql, err)
	}
	return nil
}

func (da *Access) Close() error {
	return nil
}
