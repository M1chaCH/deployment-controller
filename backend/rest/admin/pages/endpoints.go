package pages

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func Init(router *gin.RouterGroup) {
	router.GET("/pages", getPages)
	router.POST("/pages", postPage)
	router.PUT("/pages", putPage)
	router.DELETE("/pages/:id", deletePage)

	tx := framework.DB().MustBegin()
	_, err := LoadPages(tx)
	if err != nil {
		panic(fmt.Sprintf("failed to init all pages: %v", err))
	}
	err = tx.Commit()
	if err != nil {
		panic(fmt.Sprintf("failed to init all pages, transaction not committed ?!?: %v", err))
	}
}

func getPages(c *gin.Context) {
	var pages, err = LoadPages(framework.GetTx(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, pages)
}

func postPage(c *gin.Context) {
	var page Page
	if err := c.ShouldBindJSON(&page); err != nil {
		log.Printf("failed to bind page from request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "provided page is not valid"})
		return
	}

	if SimilarPageExists(page.Id, page.Title, page.Description) {
		c.JSON(http.StatusConflict, gin.H{"message": "page already exists"})
		return
	}

	err := InsertPage(framework.GetTx(c), page)
	if err != nil {
		log.Panicf("failed to insert page %v: %v", page, err)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "page created"})
}

func putPage(c *gin.Context) {
	var page Page
	if err := c.ShouldBindJSON(&page); err != nil {
		log.Printf("failed to bind page from request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "provided page is not valid"})
		return
	}

	if !PageExists(page.Id) {
		c.JSON(http.StatusNotFound, gin.H{"message": "page does not exist"})
		return
	}

	err := UpdatePage(framework.GetTx(c), page)
	if err != nil {
		log.Panicf("failed to update page %v: %v", page, err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "page updated"})
}

func deletePage(c *gin.Context) {
	pageId := c.Param("id")
	err := DeletePage(framework.GetTx(c), pageId)

	if err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "page deleted"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
}
