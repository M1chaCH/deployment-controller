package pages

type Page struct {
	Id            string `json:"id" db:"id"`
	TechnicalName string `json:"technicalName" db:"technical_name"`
	Url           string `json:"url" db:"url"`
	Title         string `json:"title" db:"title"`
	Description   string `json:"description" db:"description"`
	PrivatePage   bool   `json:"privatePage" db:"private_page"`
}
