package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"

	ss "github.com/catinello/base62"
)

// ShortenedURLResource holds information about URLs
type ShortenedURLResource struct {
	ID        int       `json:"id" db:"id"`
	URL       string    `json:"url" db:"url"`
	ShortURL  string    `json:"short_url" db:"short_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// GetShortUrl returns the url detail
func GetShortUrl(c *gin.Context) {
	var shortenedURL ShortenedURLResource
	shortUrl := c.Param("short_url")
	db := c.MustGet("DB").(*sqlx.DB)
	err := db.Get(&shortenedURL, "SELECT * FROM url WHERE short_url=$1", shortUrl)
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
		result := tx.MustExec("INSERT INTO url (url) VALUES ($1)", shortenedURL.URL)
		err = tx.Commit()
		if err == nil {
			newID, _ := result.LastInsertId()
			shortenedURL.ShortURL = ss.Encode(int(newID))
			tx := db.MustBegin()
			tx.MustExec("UPDATE url SET short_url=$1 WHERE id=$2", shortenedURL.ShortURL, int(newID))
			err = tx.Commit()
			if err != nil {
				log.Println(err)
				c.String(http.StatusInternalServerError, err.Error())
			} else {
				c.JSON(http.StatusOK, gin.H{
					"shortUrl": shortenedURL.ShortURL,
				})
			}
		} else {
			c.String(http.StatusInternalServerError, err.Error())
		}
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}
