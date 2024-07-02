package framework

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strconv"
	"time"
)

func DB() *sqlx.DB {
	if configuredDb != nil {
		return configuredDb
	}
	config := loadDbConfig()

	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.Name))
	if err != nil {
		panic(fmt.Sprintf("failed to open DB: %s", err))
	}

	err = db.Ping()
	if err != nil {
		panic(fmt.Sprintf("failed to ping DB: %s", err))
	}

	db.SetMaxIdleConns(12)
	db.SetMaxOpenConns(20)
	db.SetConnMaxIdleTime(time.Hour)
	db.SetConnMaxLifetime(8 * time.Hour)

	log.Printf("Connected to database: %s:%d %s", config.Host, config.Port, config.Name)
	configuredDb = db
	return configuredDb
}

type dbConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

var configuredDb *sqlx.DB

func loadDbConfig() dbConfig {
	var config dbConfig

	config.Name = os.Getenv("DB_NAME")
	if config.Name == "" {
		log.Println("DB Name is not configured, using default: 'deployment_controller'")
		config.Name = "deployment_controller"
	}

	config.User = os.Getenv("DB_USER")
	if config.User == "" {
		panic("DB User not configured")
	}

	config.Password = os.Getenv("DB_PASS")
	if config.Password == "" {
		panic("DB Password not configured")
	}

	config.Host = os.Getenv("DB_HOST")
	if config.Host == "" {
		log.Println("DB Host is not configured, using default: 'localhost'")
		config.Host = "localhost"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		log.Println("DB Port is not configured, using default: '5432'")
		config.Port = 5432
	} else {
		portNumber, err := strconv.Atoi(port)
		if err != nil {
			panic(fmt.Sprintf("configured DB Port is not a number: %s", err))
		}
		config.Port = portNumber
	}

	return config
}
