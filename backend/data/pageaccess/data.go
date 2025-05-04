package pageaccess

import (
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/caches"
	"github.com/M1chaCH/deployment-controller/framework/logs"
)

func InitCache() {
	logs.Info(nil, "initializing pageaccess cache")

	initialCacheEntries := make([]UserPageAccess, 0)

	tx, err := framework.DB().Beginx()

	if err != nil {
		logs.Panic(nil, "failed to begin transaction during pageaccess cache initialisation: %v", err)
	}

	var pageAccessResult []userPageAccessResult
	err = tx.Select(&pageAccessResult, `
select p.id as page_id, u.id as user_id, p.technical_name, p.private_page,
       CASE
           WHEN (up.user_id IS NOT NULL AND u.onboard AND NOT u.blocked)
               OR p.private_page IS NOT TRUE
               THEN TRUE
           ELSE FALSE
           END AS has_access
from users as u
    full join pages as p on true
    left join public.user_page up on p.id = up.page_id and u.id = up.user_id
`)
	if err != nil {
		logs.Panic(nil, "failed to initialize pageaccess cache: %v", err)
	}

	userPageAccess := map[string][]PageAccessPage{}
	for _, result := range pageAccessResult {
		currentUserAccess, ok := userPageAccess[result.UserId]
		if !ok {
			currentUserAccess = []PageAccessPage{}
		}

		currentUserAccess = append(currentUserAccess, PageAccessPage{
			PageId:        result.PageId,
			Access:        result.HasAccess,
			TechnicalName: result.TechnicalName,
			PrivatePage:   result.PrivatePage,
		})
		userPageAccess[result.UserId] = currentUserAccess
	}

	for key, value := range userPageAccess {
		initialCacheEntries = append(initialCacheEntries, UserPageAccess{
			UserId: key,
			Pages:  value,
		})
		if err != nil {
			logs.Panic(nil, "failed to cache pageaccess: %v", err)
		}
	}

	// setup cache try for not logged-in requests
	anonPageAccess := make([]PageAccessPage, 0)
	err = tx.Select(&anonPageAccess, `
SELECT p.id as page_id, p.technical_name, NOT p.private_page as has_access
FROM pages p
`)
	if err != nil {
		logs.Panic(nil, "failed to initialize anon pageaccess cache: %v", err)
	}

	initialCacheEntries = append(initialCacheEntries, UserPageAccess{
		UserId: AnonymousUserId,
		Pages:  anonPageAccess,
	})

	err = cache.Initialize(initialCacheEntries)
	if err != nil {
		logs.Panic(nil, "failed to initialize pageaccess cache: %v", err)
	}
	logs.Info(nil, "successfully initialized pageaccess cache")
}

func LoadUserPageAccess(txFunc framework.LoadableTx, userId string) (UserPageAccess, error) {
	access, found := cache.Get(userId)
	if found {
		return access, nil
	}

	tx, err := txFunc()
	if err != nil {
		return UserPageAccess{}, err
	}

	var pageAccessResult []PageAccessPage
	err = tx.Select(&pageAccessResult, `
SELECT p.id as page_id, p.technical_name, p.private_page,
       CASE
           WHEN (up.user_id IS NOT NULL AND u.onboard AND NOT u.blocked)
               OR p.private_page IS NOT TRUE
               THEN TRUE
           ELSE FALSE
           END AS has_access
FROM pages AS p
         LEFT JOIN user_page up ON p.id = up.page_id AND up.user_id = $1
		 LEFT JOIN users u ON u.id = up.user_id
`, userId)
	if err != nil {
		return UserPageAccess{}, err
	}

	userPageAccess := UserPageAccess{
		UserId: userId,
		Pages:  pageAccessResult,
	}

	err = cache.Store(userPageAccess)
	if err != nil {
		if err.Error() == caches.ErrCacheNotInitialized {
			InitCache()
		} else {
			return UserPageAccess{}, err
		}
	}
	return userPageAccess, nil
}

func DeleteUserPageAccessCache(userId string) {
	err := cache.Remove(userId)
	if err != nil {
		logs.Warn(nil, "failed to delete user page access cache entry: %v", err)
	}
}

func ClearCache() {
	cache.Clear()
}
