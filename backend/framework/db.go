package framework

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"net/http"
	"time"
)

func GetTx(c *gin.Context) *sqlx.Tx {
	value, ok := c.Get(txContextKey)
	if !ok {
		panic("transaction was never set?!?")
	}

	tx, ok := value.(*sqlx.Tx)
	if !ok {
		panic("set transaction is not a transaction?!?")
	}
	return tx
}

func DB() *sqlx.DB {
	if configuredDb != nil {
		return configuredDb
	}
	config := getAndValidateDbConfig()

	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Db.Host, config.Db.Port, config.Db.User, config.Db.Password, config.Db.Name))
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

	logs.Info(fmt.Sprintf("Connected to database: %s:%d %s", config.Db.Host, config.Db.Port, config.Db.Name))
	configuredDb = db
	return configuredDb
}

const txContextKey = "DB_TRANSACTION"

// todo maybe not middleware but part of GetTx(), because this transaction is not always used if line 71 is changed
func TransactionMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		db := DB()

		tx, err := db.Beginx()
		if err != nil {
			logs.Warn(fmt.Sprintf("failed to begin transaction: %s", err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "could not begin transaction"})
			return
		}

		c.Set(txContextKey, tx)
		c.Next()

		// TODO, why only when request succeeded?
		// because if there was no panic, but i am aborting a request with code 404 or so, then I want the changes to be reverted, right?
		// -> can't always store client updates (might get reverted later ):). -> maybe don't use transaction there?
		if c.Writer.Status() < 300 {
			err = tx.Commit()
			if err != nil {
				logs.Info(fmt.Sprintf("failed to commit transaction: %s", err))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "changes could not be saved"})
				return
			}
		} else {
			err = tx.Rollback()
			if err != nil {
				logs.Info(fmt.Sprintf("failed to rollback transaction: %s", err))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal DB error"})
				return
			}
		}
	}
}

var configuredDb *sqlx.DB

func getAndValidateDbConfig() *AppConfig {
	config := Config()

	if config.Db.Name == "" {
		logs.Info("DB Name is not configured, using default: 'deployment_controller'")
		config.Db.Name = "deployment_controller"
	}

	if config.Db.User == "" {
		panic("DB User not configured")
	}

	if config.Db.Password == "" {
		panic("DB Password not configured")
	}

	if config.Db.Host == "" {
		logs.Info("DB Host is not configured, using default: 'localhost'")
		config.Db.Host = "localhost"
	}

	if config.Db.Port == 0 {
		logs.Info("DB Port is not configured, using default: '5432'")
		config.Db.Port = 5432
	}

	return config
}
