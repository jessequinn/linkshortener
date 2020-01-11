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
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserResource holds information about users
type UserResource struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"password" db:"password"`
	Token     string    `json:"token" db:"token"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
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

// RemoveShortUrl handles the removing of resource
func RemoveShortUrl(c *gin.Context) {
	db := c.MustGet("DB").(*sqlx.DB)
	shortUrl := c.Param("short_url")
	tx := db.MustBegin()
	tx.MustExec("DELETE FROM url WHERE short_url=$1", shortUrl)
	err := tx.Commit()
	if err != nil {
		log.Println(err)
		c.String(http.StatusInternalServerError, err.Error())
	} else {
		c.String(http.StatusOK, "")
	}
}

// UpdateShortUrl returns 200 only
func UpdateShortUrl(c *gin.Context) {
	var shortenedURL ShortenedURLResource
	shortUrl := c.Param("short_url")
	db := c.MustGet("DB").(*sqlx.DB)
	if err := c.BindJSON(&shortenedURL); err == nil {
		tx := db.MustBegin()
		id, err := ss.Decode(shortUrl)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
		tx.MustExec("UPDATE url SET url=$1, updated_at=$2 WHERE id=$3", shortenedURL.URL, time.Now().UTC().Format("2006-01-02 15:04:05"), id)
		err = tx.Commit()
		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.String(http.StatusOK, "")
		}
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

// RegisterUser handles the POST
func RegisterUser(c *gin.Context) {
	var userResource UserResource
	db := c.MustGet("DB").(*sqlx.DB)
	if err := c.BindJSON(&userResource); err == nil {
		rows, err := db.NamedQuery(`SELECT * FROM user WHERE username=:username`, userResource)
		defer rows.Close()
		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			if rows.Next() == false {
				tx := db.MustBegin()
				tx.MustExec("INSERT INTO user (username, password) VALUES ($1, $2)", userResource.Username, userResource.Password)
				err = tx.Commit()
				if err != nil {
					log.Println(err)
					c.String(http.StatusInternalServerError, err.Error())
				} else {
					c.JSON(http.StatusOK, gin.H{
						"code":    http.StatusOK,
						"message": "User registered",
					})
				}
			} else {
				c.JSON(http.StatusOK, gin.H{
					"code":    http.StatusOK,
					"message": "User already registered",
				})
			}
		}
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

// Health check endpoint
func Health(c *gin.Context) {
	c.JSON(200, gin.H{
		"serverTime": time.Now().UTC(),
	})
}
