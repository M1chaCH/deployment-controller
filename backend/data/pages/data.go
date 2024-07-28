package pages

import (
	"database/sql"
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/jmoiron/sqlx"
)

func InitCache() {
	logs.Info("initialized pages cache")
	tx, err := framework.DB().Beginx()
	if err != nil {
		logs.Panic(fmt.Sprintf("failed to begin tx for page cache: %v", err))
	}

	items, err := LoadPages(tx)
	if err != nil {
		logs.Panic(fmt.Sprintf("failed to load items for page cache: %v", err))
	}

	err = cache.Initialize(items)
	if err != nil {
		logs.Panic(fmt.Sprintf("failed to initialize page cache: %v", err))
	}

	err = tx.Commit()
	if err != nil {
		logs.Panic(fmt.Sprintf("failed to commit tx: %v", err))
	}

	logs.Info("successfully initialized pages cache")
}

func LoadPages(tx *sqlx.Tx) ([]Page, error) {
	if cache.IsInitialized() {
		result, err := cache.GetAllAsArray()
		if err != nil || len(result) > 0 {
			return result, err
		}
	}

	logs.Info("no pages found in cache, checking db")

	var pages []Page
	err := tx.Select(&pages, "select * from pages")

	if err == nil {
		err = cache.Initialize(pages)
		if err != nil {
			logs.Warn(fmt.Sprintf("failed to cache pages: %v", err))
			return nil, err
		}
	}

	return pages, err
}

func InsertNewPage(tx *sqlx.Tx, page Page) error {
	_, err := tx.NamedExec("insert into pages (id, technical_name, url, title, description, private_page) VALUES (:id, :technical_name, :url, :title, :description, :private_page)", page)

	if err == nil {
		cache.StoreSafeBackground(page)
	}

	return err
}

func UpdatePage(tx *sqlx.Tx, page Page) error {
	result, err := tx.NamedExec("update pages set technical_name = :technical_name, url = :url, title = :title, description = :description, private_page = :private_page where id = :id", page)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	cache.StoreSafeBackground(page)

	return nil
}

func DeletePage(tx *sqlx.Tx, pageId string) error {
	_, err := tx.Exec(`delete from pages where id = $1`, pageId)
	if err == nil {
		err = cache.Remove(pageId)
	}
	return err
}
