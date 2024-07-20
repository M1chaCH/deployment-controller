package pages

import (
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/jmoiron/sqlx"
)

// fixme, UI allows PageIds to change, use readonly UUID for Id and then editable Human Readable Id
type Page struct {
	Id          string `json:"id" db:"id"`
	Url         string `json:"url" db:"url"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`
	PrivatePage bool   `json:"private_page" db:"private_page"`
}

func (p Page) GetCacheKey() string {
	return p.Id
}

var cache framework.ItemsCache[Page] = &framework.LocalItemsCache[Page]{}

func LoadPages(tx *sqlx.Tx) ([]Page, error) {
	if !cache.IsInitialized() {
		pages := make([]Page, 0)
		err := tx.Select(&pages, "SELECT * FROM pages")
		if err != nil {
			return nil, err
		}

		cache.Initialize(pages)
		return pages, nil
	}

	return cache.GetAll(), nil
}

func InsertPage(tx *sqlx.Tx, page Page) error {
	_, err := tx.NamedExec("INSERT INTO pages (id, url, title, description, private_page) VALUES (:id, :url, :title, :description, :private_page)", page)

	if err == nil {
		go cache.Store(page)
	}

	return err
}

func UpdatePage(tx *sqlx.Tx, page Page) error {
	_, err := tx.NamedExec(`
UPDATE pages
SET url = :url, title = :title, description = :description, private_page = :private_page
WHERE id = :id
`, page)

	if err == nil {
		go cache.Store(page)
	}

	return err
}

func DeletePage(tx *sqlx.Tx, id string) error {
	_, err := tx.Exec("DELETE FROM pages WHERE id = $1", id)

	if err == nil {
		cache.Remove(id)
	}

	return err
}

func PageExists(id string) bool {
	_, ok := cache.Get(id)
	return ok
}

func SimilarPageExists(id string, title string, description string) bool {
	if PageExists(id) {
		return true
	}

	for _, page := range cache.GetAll() {
		if page.Title == title && page.Description == description {
			return true
		}
	}

	return false
}
