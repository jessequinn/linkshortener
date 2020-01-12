package handlers

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"time"

	ss "github.com/catinello/base62"
)

// ShortenedURLResource holds information about URLs
type ShortenedURLResource struct {
	ID        int       `json:"id" db:"id"`
	UserId    int       `json:"user_id" db:"user_id"`
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

// RegisterUser handles the user registration
func RegisterUser(c *gin.Context) {
	var userResource UserResource
	// Database connection
	db := c.MustGet("DB").(*sqlx.DB)
	if err := c.BindJSON(&userResource); err == nil {
		// Test if the username exists
		rows, err := db.NamedQuery(`SELECT * FROM appuser WHERE username=:username`, userResource)
		log.Println(err)
		defer rows.Close()
		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			if rows.Next() == false {
				tx := db.MustBegin()
				tx.MustExec("INSERT INTO appuser (username, password) VALUES ($1, $2)", userResource.Username, userResource.Password)
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

// CreateShortUrl handles the creation of short url
func CreateShortUrl(c *gin.Context) {
	var shortenedURLResource ShortenedURLResource
	// User - take from JWT claim
	claims := jwt.ExtractClaims(c)
	user := claims["id"]
	// Database connection
	db := c.MustGet("DB").(*sqlx.DB)
	if err := c.BindJSON(&shortenedURLResource); err == nil {
		// Fetch user that is logged in
		userResource := UserResource{}
		err = db.Get(&userResource, "SELECT * FROM appuser WHERE username=$1", user)
		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			// Check if user has this url registered already
			rows, err := db.NamedQuery(`SELECT * FROM appurl WHERE user_id=:user_id AND url=:url`, map[string]interface{}{"user_id": userResource.ID, "url": shortenedURLResource.URL})
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
			}
			defer rows.Close()
			if rows.Next() == false {
				// Insert new record
				lastInsertId := 0
				err := db.QueryRow("INSERT INTO appurl (user_id, url) VALUES ($1, $2) RETURNING id", userResource.ID, shortenedURLResource.URL).Scan(&lastInsertId)
				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
				}
				log.Println(lastInsertId)
				log.Println("**************************************")
				if err == nil {
					// Create a short url based on the index using base62
					log.Println(lastInsertId)
					shortenedURLResource.ShortURL = ss.Encode(lastInsertId)
					tx := db.MustBegin()
					tx.MustExec("UPDATE appurl SET short_url=$1 WHERE id=$2", shortenedURLResource.ShortURL, lastInsertId)
					err = tx.Commit()
					if err != nil {
						log.Println(err)
						c.String(http.StatusInternalServerError, err.Error())
					} else {
						c.JSON(http.StatusOK, gin.H{
							"shortUrl": shortenedURLResource.ShortURL,
						})
					}
				} else {
					c.String(http.StatusInternalServerError, err.Error())
				}
			} else {
				shortenedURLResource2 := ShortenedURLResource{}
				err = db.Get(&shortenedURLResource2, "SELECT * FROM appurl WHERE user_id=$1 AND url=$2", userResource.ID, shortenedURLResource.URL)
				if err != nil {
					log.Println(err)
					c.String(http.StatusInternalServerError, err.Error())
				} else {
					c.JSON(http.StatusOK, gin.H{
						"code":     http.StatusOK,
						"shortUrl": shortenedURLResource2.ShortURL,
					})
				}
			}
		}
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
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
	var shortenedURLResource ShortenedURLResource
	shortUrl := c.Param("short_url")
	db := c.MustGet("DB").(*sqlx.DB)
	if err := c.BindJSON(&shortenedURLResource); err == nil {
		tx := db.MustBegin()
		id, err := ss.Decode(shortUrl)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
		tx.MustExec("UPDATE url SET url=$1, updated_at=$2 WHERE id=$3", shortenedURLResource.URL, time.Now().UTC().Format("2006-01-02 15:04:05"), id)
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

// Health check endpoint
func Health(c *gin.Context) {
	c.JSON(200, gin.H{
		"serverTime": time.Now().UTC(),
	})
}
