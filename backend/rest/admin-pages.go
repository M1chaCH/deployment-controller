package rest

import (
	"database/sql"
	"errors"
	"github.com/M1chaCH/deployment-controller/auth"
	"github.com/M1chaCH/deployment-controller/data/pageaccess"
	"github.com/M1chaCH/deployment-controller/data/pages"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

var invalidCharsForTechnicalNameRegex = regexp.MustCompile(`[^a-zA-Z0-9-_]`)

func getPages(c *gin.Context) {
	data, err := pages.LoadPages(framework.GetTx(c))
	if err != nil {
		logs.Warn(c, "failed to load pages: %v", err)
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to read pages"})
		return
	}
	auth.RespondWithCookie(c, http.StatusOK, data)
}

func postPage(c *gin.Context) {
	tx := framework.GetTx(c)
	var page pages.Page
	if err := c.ShouldBindJSON(&page); err != nil {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "page has invalid format"})
		return
	}

	if !assertTechnicalPageName(c, page.TechnicalName) {
		return
	}

	err := pages.InsertNewPage(tx, page)
	if err != nil {
		logs.Warn(c, "failed to insert new page: %v", err)
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to create page"})
		return
	}

	pageaccess.ClearCache()
	auth.RespondWithCookie(c, http.StatusOK, gin.H{"message": "created page"})
}

func putPage(c *gin.Context) {
	tx := framework.GetTx(c)
	var page pages.Page
	if err := c.ShouldBindJSON(&page); err != nil {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "page has invalid format"})
		return
	}

	if !assertTechnicalPageName(c, page.TechnicalName) {
		return
	}

	err := pages.UpdatePage(tx, page)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			auth.RespondWithCookie(c, http.StatusNotFound, gin.H{"message": "page not found"})
			return
		}

		logs.Warn(c, "failed to update page: %v", err)
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "failed to update page"})
		return
	}

	pageaccess.ClearCache()
	auth.RespondWithCookie(c, http.StatusOK, gin.H{"message": "updated page"})
}

func deletePage(c *gin.Context) {
	tx := framework.GetTx(c)
	pageId := c.Param("id")
	err := pages.DeletePage(tx, pageId)
	if err != nil {
		logs.Warn(c, "failed to delete page: %v", err)
		auth.RespondWithCookie(c, http.StatusInternalServerError, gin.H{"message": "page not deleted"})
		return
	}

	pageaccess.ClearCache()
	auth.RespondWithCookie(c, http.StatusOK, gin.H{"message": "deleted page"})
}

func assertTechnicalPageName(c *gin.Context, pageName string) bool {
	if len(pageName) < 3 {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "technical name must at least be 3 characters long"})
		return false
	}

	if invalidCharsForTechnicalNameRegex.MatchString(pageName) {
		auth.RespondWithCookie(c, http.StatusBadRequest, gin.H{"message": "technical name contains invalid characters"})
		return false
	}

	return true
}
