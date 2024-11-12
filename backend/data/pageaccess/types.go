package pageaccess

import "github.com/M1chaCH/deployment-controller/framework/caches"

type UserPageAccess struct {
	UserId string
	Pages  []PageAccessPage
}

type PageAccessPage struct {
	PageId        string `db:"page_id" json:"pageId"`
	Access        bool   `db:"has_access" json:"hasAccess"`
	TechnicalName string `db:"technical_name" json:"technicalName"`
	PrivatePage   bool   `db:"private_page" json:"privatePage"`
}

func (pa UserPageAccess) GetCacheKey() string {
	return pa.UserId
}

func (p PageAccessPage) GetTechnicalName() string {
	return p.TechnicalName
}

func (p PageAccessPage) GetAccessAllowed() bool {
	return p.Access
}

var cache = caches.GetCache[UserPageAccess]()

type userPageAccessResult struct {
	UserId        string `db:"user_id"`
	PageId        string `db:"page_id"`
	HasAccess     bool   `db:"has_access"`
	TechnicalName string `db:"technical_name"`
	PrivatePage   bool   `db:"private_page"`
}

const AnonymousUserId = "anon"
