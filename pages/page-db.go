package pages

import (
	"github.com/M1chaCH/deployment-controller/framework"
)

var pageCache = make(map[string]Page)

// TODO change all DB Access to transactions
func LoadPages() ([]Page, error) {
	if len(pageCache) == 0 {
		db := framework.DB()
		pages := make([]Page, 0)

		err := db.Select(&pages, "SELECT * FROM pages")
		if err != nil {
			return nil, err
		}

		for _, page := range pages {
			pageCache[page.ID] = page
		}

		return pages, nil
	}

	return getCachedPages(), nil
}

func InsertPage(page Page) error {
	db := framework.DB()
	_, err := db.NamedExec("INSERT INTO pages (id, url, title, description, private_page) VALUES (:id, :url, :title, :description, :private_page)", page)

	if err == nil {
		pageCache[page.ID] = page
	}

	return err
}

func UpdatePage(page Page) error {
	db := framework.DB()
	_, err := db.NamedExec(`
UPDATE pages
SET url = :url, title = :title, description = :description, private_page = :private_page
WHERE id = :id
`, page)

	if err == nil {
		pageCache[page.ID] = page
	}

	return err
}

func DeletePage(id string) error {
	db := framework.DB()
	_, err := db.Exec("DELETE FROM pages WHERE id = $1", id)

	if err == nil {
		delete(pageCache, id)
	}

	return err
}

func PageExists(id string) bool {
	_, ok := pageCache[id]
	return ok
}

func SimilarPageExists(id string, title string, description string) bool {
	if PageExists(id) {
		return true
	}

	for _, page := range pageCache {
		if page.Title == title && page.Description == description {
			return true
		}
	}

	return false
}

func getCachedPages() []Page {
	var pages []Page
	for _, page := range pageCache {
		pages = append(pages, page)
	}
	return pages
}
