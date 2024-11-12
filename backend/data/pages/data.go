package pages

import (
	"database/sql"
	"github.com/M1chaCH/deployment-controller/framework"
)

func LoadPages(txFunc framework.LoadableTx) ([]Page, error) {
	tx, err := txFunc()
	if err != nil {
		return nil, err
	}

	var pages []Page
	err = tx.Select(&pages, "select * from pages")

	return pages, err
}

func InsertNewPage(txFunc framework.LoadableTx, page Page) error {
	tx, err := txFunc()
	if err != nil {
		return err
	}
	_, err = tx.NamedExec("insert into pages (id, technical_name, url, title, description, private_page) VALUES (:id, :technical_name, :url, :title, :description, :private_page)", page)

	return err
}

func UpdatePage(txFunc framework.LoadableTx, page Page) error {
	tx, err := txFunc()
	if err != nil {
		return err
	}

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

func DeletePage(txFunc framework.LoadableTx, pageId string) error {
	tx, err := txFunc()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`delete from pages where id = $1`, pageId)
	return err
}
