package pages

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
)

func LoadPages(tx *sqlx.Tx) ([]Page, error) {
	var pages []Page
	err := tx.Select(&pages, "select * from pages")
	return pages, err
}

func InsertNewPage(tx *sqlx.Tx, page Page) error {
	_, err := tx.NamedExec("insert into pages (id, technical_name, url, title, description, private_page) VALUES (:id, :technical_name, :url, :title, :description, :private_page)", page)
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
	return nil
}

func DeletePage(tx *sqlx.Tx, pageId string) error {
	_, err := tx.Exec(`delete from pages where id = $1`, pageId)
	return err
}
