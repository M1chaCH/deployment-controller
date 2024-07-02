package pages

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type Page struct {
	ID          string `json:"id" db:"id"`
	Url         string `json:"url" db:"url"`
	Title       string `json:"title" db:"title"`
	Description string `json:"description" db:"description"`
	PrivatePage bool   `json:"private_page" db:"private_page"`
}

func Init(router *gin.Engine) {
	router.GET("/pages", getPages)
	router.POST("/pages", postPage)
	router.PUT("/pages", putPage)
	router.DELETE("/pages/:id", deletePage)

	_, err := LoadPages()
	if err != nil {
		panic(fmt.Sprintf("failed to init all pages: %v", err))
	}
}

func getPages(c *gin.Context) {
	var pages, err = LoadPages()
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

	if SimilarPageExists(page.ID, page.Title, page.Description) {
		c.JSON(http.StatusConflict, gin.H{"message": "page already exists"})
		return
	}

	err := InsertPage(page)
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

	if !PageExists(page.ID) {
		c.JSON(http.StatusNotFound, gin.H{"message": "page does not exist"})
		return
	}

	err := UpdatePage(page)
	if err != nil {
		log.Panicf("failed to update page %v: %v", page, err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "page updated"})
}

func deletePage(c *gin.Context) {
	pageId := c.Param("id")
	err := DeletePage(pageId)

	if err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "page deleted"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
}
