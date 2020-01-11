package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
)

// ShortenedURLResource holds information about URLs
type ShortenedURLResource struct {
	ID        int       `json:"id"`
	URL       string    `json:"url"`
	ShortURL  string    `json:"short_url"`
	CreatedAt time.Time `json:"created_at"`
}

// GetShortUrl returns the url detail
func GetShortUrl(c *gin.Context) {
	var shortenedURL ShortenedURLResource
	id := c.Param("url_id")
	db := c.MustGet("DB").(*sqlx.DB)
	err := db.Get(&shortenedURL, "SELECT * FROM url WHERE id=$1", id)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"result": shortenedURL,
		})
	}
}

// CreateShortUrl handles the POST
func CreateShortUrl(c *gin.Context) {
	var shortenedURL ShortenedURLResource
	db := c.MustGet("DB").(*sqlx.DB)
	if err := c.BindJSON(&shortenedURL); err == nil {
		tx := db.MustBegin()
		// TODO: add logic for the creation of shorturl
		result, err := tx.NamedExec("INSERT INTO url (url, shorturl) VALUES (:url, :shorturl)", &ShortenedURLResource{URL: shortenedURL.URL, ShortURL: shortenedURL.URL})
		err = tx.Commit()
		if err == nil {
			newID, _ := result.LastInsertId()
			shortenedURL.ID = int(newID)
			shortenedURL.ShortURL = shortenedURL.URL
			c.JSON(http.StatusOK, gin.H{
				"shortUrl": shortenedURL.ShortURL,
			})
		} else {
			c.String(http.StatusInternalServerError, err.Error())
		}
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}
