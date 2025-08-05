package repository

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"strconv"
	"time"
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
	
	// Get the underlying SQL database to configure connection pooling
	sqlDB, err := da.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL database: %v", err)
	}
	
	// Configure connection pool with environment variables, defaulting to lowest values
	maxOpenConns := getEnvAsInt("DB_MAX_OPEN_CONNS", 1)
	maxIdleConns := getEnvAsInt("DB_MAX_IDLE_CONNS", 1)
	connMaxLifetimeMinutes := getEnvAsInt("DB_CONN_MAX_LIFETIME_MINUTES", 5)
	connMaxIdleTimeMinutes := getEnvAsInt("DB_CONN_MAX_IDLE_TIME_MINUTES", 1)
	
	// Set maximum number of open connections
	sqlDB.SetMaxOpenConns(maxOpenConns)
	
	// Set maximum number of idle connections
	sqlDB.SetMaxIdleConns(maxIdleConns)
	
	// Set maximum lifetime of a connection
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetimeMinutes) * time.Minute)
	
	// Set maximum idle time for a connection
	sqlDB.SetConnMaxIdleTime(time.Duration(connMaxIdleTimeMinutes) * time.Minute)
	
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

// Query the final query to get the results.
func (da *Access) Query(query string, dest any, arguments ...any) error {
	if err := da.DB.Raw(query, arguments...).Scan(dest).Error; err != nil {
		return fmt.Errorf("failed to fetch data values: %v", err)
	}

	return nil
}

func (da *Access) Close() error {
	return nil
}

// getEnvAsInt reads an environment variable as an integer with a default value
func getEnvAsInt(name string, defaultValue int) int {
	valueStr, exists := os.LookupEnv(name)
	if !exists {
		return defaultValue
	}
	
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Warning: Invalid integer value for %s: %s, using default: %d", name, valueStr, defaultValue)
		return defaultValue
	}
	
	return value
}
