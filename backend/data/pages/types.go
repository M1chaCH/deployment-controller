package pages

import "github.com/M1chaCH/deployment-controller/framework/caches"

type Page struct {
	Id            string `json:"id" db:"id"`
	TechnicalName string `json:"technicalName" db:"technical_name"`
	Url           string `json:"url" db:"url"`
	Title         string `json:"title" db:"title"`
	Description   string `json:"description" db:"description"`
	PrivatePage   bool   `json:"privatePage" db:"private_page"`
}

func (p Page) GetCacheKey() string {
	return p.Id
}

func (p Page) GetTechnicalName() string {
	return p.TechnicalName
}

func (p Page) GetAccessAllowed() bool {
	return !p.PrivatePage
}

var cache = caches.GetCache[Page]()
