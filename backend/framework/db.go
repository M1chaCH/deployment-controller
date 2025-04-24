package framework

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework/config"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"net/http"
	"time"
)

const ErrTooManyTokens = "too many tokens found per user"
const ErrNotValidated = "token is not validated"

type LoadableTx func() (*sqlx.Tx, error)

// GetTx lazy loads the transaction for the request.
// the transaction is only started on the first execution of the inner function
// the transaction will automatically be committed / rollback if it was started
func GetTx(c *gin.Context) LoadableTx {
	return func() (*sqlx.Tx, error) {
		tx := getTxFromContext(c)
		if tx != nil {
			return tx, nil
		}

		tx, err := DB().Beginx()
		if err != nil {
			return nil, err
		}
		c.Set(txContextKey, tx)
		return tx, nil
	}
}

// DB creates and tests a new connection to the DB.
// TODO remove panics, this might be called from within requests, don't want to crash backend
func DB() *sqlx.DB {
	if configuredDb != nil {
		return configuredDb
	}
	cnf := getAndValidateDbConfig()

	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cnf.Db.Host, cnf.Db.Port, cnf.Db.User, cnf.Db.Password, cnf.Db.Name))
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

	logs.Info(nil, "Connected to database: %s:%d %s", cnf.Db.Host, cnf.Db.Port, cnf.Db.Name)
	configuredDb = db
	return configuredDb
}

const txContextKey = "DB_TRANSACTION"

func getTxFromContext(c *gin.Context) *sqlx.Tx {
	value, ok := c.Get(txContextKey)
	if !ok {
		return nil
	}

	tx, ok := value.(*sqlx.Tx)
	if !ok {
		logs.Error(c, "set transaction is not a transaction?!?")
	}
	return tx
}

func TransactionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		tx := getTxFromContext(c)
		if tx == nil {
			// no tx was used in this request, so no commit required
			return
		}

		// why only when request succeeded?
		// because if there was no panic, but I am aborting a request with code 404 or so, then I want the changes to be reverted
		// -> can't always store client updates -> don't use transaction for things that must always persist.
		if c.Writer.Status() < 400 {
			err := tx.Commit()
			if err != nil {
				logs.Info(c, "failed to commit transaction: %s", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "changes could not be saved"})
				return
			}
		} else {
			err := tx.Rollback()
			if err != nil {
				logs.Info(c, "failed to rollback transaction: %s", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "internal DB error"})
				return
			}
			logs.Info(c, "transaction rolled back")
		}
	}
}

var configuredDb *sqlx.DB

func getAndValidateDbConfig() *config.AppConfig {
	cnf := config.Config()

	if cnf.Db.Name == "" {
		logs.Info(nil, "DB Name is not configured, using default: 'deployment_controller'")
		cnf.Db.Name = "deployment_controller"
	}

	if cnf.Db.User == "" {
		panic("DB User not configured")
	}

	if cnf.Db.Password == "" {
		panic("DB Password not configured")
	}

	if cnf.Db.Host == "" {
		logs.Info(nil, "DB Host is not configured, using default: 'localhost'")
		cnf.Db.Host = "localhost"
	}

	if cnf.Db.Port == 0 {
		logs.Info(nil, "DB Port is not configured, using default: '5432'")
		cnf.Db.Port = 5432
	}

	return cnf
}
