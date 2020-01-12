package handlers

import (
	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"strings"
	"time"

	ss "github.com/catinello/base62"
)

// ShortenedURLResource - holds information about URLs
type ShortenedURLResource struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	URL       string    `json:"url" db:"url"`
	ShortURL  string    `json:"short_url" db:"short_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserResource - holds information about users
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

// CreateShortURL handles the creation of short url
func CreateShortURL(c *gin.Context) {
	var shortenedURLResource ShortenedURLResource
	// User - take from JWT claim
	claims := jwt.ExtractClaims(c)
	user := claims["id"]
	// Database connection
	db := c.MustGet("DB").(*sqlx.DB)
	if err := c.BindJSON(&shortenedURLResource); err == nil {
		// Fetch user that is logged in
		userResource := UserResource{}
		err := db.Get(&userResource, "SELECT * FROM appuser WHERE username=$1", user)
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
				lastInsertID := 0
				err := db.QueryRow("INSERT INTO appurl (user_id, url) VALUES ($1, $2) RETURNING id", userResource.ID, shortenedURLResource.URL).Scan(&lastInsertID)
				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
				}
				log.Println(lastInsertID)
				log.Println("**************************************")
				if err == nil {
					// Create a short url based on the index using base62
					log.Println(lastInsertID)
					shortenedURLResource.ShortURL = ss.Encode(lastInsertID)
					tx := db.MustBegin()
					tx.MustExec("UPDATE appurl SET short_url=$1 WHERE id=$2", shortenedURLResource.ShortURL, lastInsertID)
					err = tx.Commit()
					if err != nil {
						log.Println(err)
						c.String(http.StatusInternalServerError, err.Error())
					} else {
						c.JSON(http.StatusOK, gin.H{
							"code":     http.StatusOK,
							"shortURL": shortenedURLResource.ShortURL,
						})
					}
				} else {
					c.String(http.StatusInternalServerError, err.Error())
				}
			} else {
				shortenedURLResource2 := ShortenedURLResource{}
				err := db.Get(&shortenedURLResource2, "SELECT * FROM appurl WHERE user_id=$1 AND url=$2", userResource.ID, shortenedURLResource.URL)
				if err != nil {
					log.Println(err)
					c.String(http.StatusInternalServerError, err.Error())
				} else {
					c.JSON(http.StatusOK, gin.H{
						"code":     http.StatusOK,
						"shortURL": shortenedURLResource2.ShortURL,
					})
				}
			}
		}
	} else {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

// GetShortURL returns the url detail
func GetShortURL(c *gin.Context) {
	var shortenedURLResource ShortenedURLResource
	// Param
	shortURL := c.Param("short_url")
	// Database connection
	db := c.MustGet("DB").(*sqlx.DB)
	// User - take from JWT claim
	claims := jwt.ExtractClaims(c)
	user := claims["id"]
	// Fetch user that is logged in
	userResource := UserResource{}
	err := db.Get(&userResource, "SELECT * FROM appuser WHERE username=$1", user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    http.StatusNotFound,
			"message": "Not records found.",
		})
	} else {
		err := db.Get(&shortenedURLResource, "SELECT * FROM appurl WHERE short_url=$1 AND user_id=$2", shortURL, userResource.ID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    http.StatusNotFound,
				"message": "Not records found.",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": http.StatusOK,
				"url":  shortenedURLResource.URL,
			})
		}
	}
}

// RemoveShortURL handles the removing of resource
func RemoveShortURL(c *gin.Context) {
	// Param
	shortURL := c.Param("short_url")
	// Database connection
	db := c.MustGet("DB").(*sqlx.DB)
	// User - take from JWT claim
	claims := jwt.ExtractClaims(c)
	user := claims["id"]
	// Fetch user that is logged in
	userResource := UserResource{}
	err := db.Get(&userResource, "SELECT * FROM appuser WHERE username=$1", user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    http.StatusNotFound,
			"message": "Not records found.",
		})
	} else {
		tx := db.MustBegin()
		result := tx.MustExec("DELETE FROM appurl WHERE short_url=$1 AND user_id=$2", shortURL, userResource.ID)
		err := tx.Commit()
		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			rows, err := result.RowsAffected()
			if err != nil {
				log.Println(err)
				c.String(http.StatusInternalServerError, err.Error())
			}
			if rows < 1 {
				c.JSON(http.StatusOK, gin.H{
					"code":    http.StatusOK,
					"message": "Record does not exist.",
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"code":    http.StatusOK,
					"message": "Record removed.",
				})
			}
		}
	}
}

// UpdateShortURL - handles updating a resource
func UpdateShortURL(c *gin.Context) {
	var shortenedURLResource ShortenedURLResource
	// Param
	shortURL := c.Param("short_url")
	// Database connection
	db := c.MustGet("DB").(*sqlx.DB)
	// User - take from JWT claim
	claims := jwt.ExtractClaims(c)
	user := claims["id"]
	// Fetch user that is logged in
	userResource := UserResource{}
	err := db.Get(&userResource, "SELECT * FROM appuser WHERE username=$1", user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    http.StatusNotFound,
			"message": "Not records found.",
		})
	} else {
		if err := c.BindJSON(&shortenedURLResource); err == nil {
			// Fetch any matching URL for user to avoid duplicates even on update
			shortenedURLResource2 := ShortenedURLResource{}
			err := db.Get(&shortenedURLResource2, "SELECT * FROM appurl WHERE url=$1 AND user_id=$2", shortenedURLResource.URL, userResource.ID)
			if err != nil {
				tx := db.MustBegin()
				id, err := ss.Decode(shortURL)
				if err != nil {
					c.String(http.StatusInternalServerError, err.Error())
				}
				result := tx.MustExec("UPDATE appurl SET url=$1, updated_at=$2 WHERE id=$3 AND user_id=$4", shortenedURLResource.URL, time.Now().UTC().Format("2006-01-02 15:04:05"), id, userResource.ID)
				err = tx.Commit()
				if err != nil {
					log.Println(err)
					c.String(http.StatusInternalServerError, err.Error())
				} else {
					rows, err := result.RowsAffected()
					if err != nil {
						log.Println(err)
						c.String(http.StatusInternalServerError, err.Error())
					}
					if rows < 1 {
						c.JSON(http.StatusOK, gin.H{
							"code":    http.StatusOK,
							"message": "Record was not updated.",
						})
					} else {
						c.JSON(http.StatusOK, gin.H{
							"code":    http.StatusOK,
							"message": "Record updated.",
						})
					}
				}
			} else {
				c.JSON(http.StatusOK, gin.H{
					"code":     http.StatusOK,
					"message":  "The URL already has a short version.",
					"shortURL": shortenedURLResource2.ShortURL,
				})
			}
		} else {
			c.String(http.StatusInternalServerError, err.Error())
		}
	}
}

// Health - check endpoint
func Health(c *gin.Context) {
	c.JSON(200, gin.H{
		"code":       http.StatusOK,
		"serverTime": time.Now().UTC(),
	})
}

// Redirect -
func Redirect(c *gin.Context) {
	var shortenedURLResource ShortenedURLResource
	// Param
	shortURL := strings.Replace(c.Request.URL.String(), "/", "", -1)
	// Database connection
	db := c.MustGet("DB").(*sqlx.DB)
	err := db.Get(&shortenedURLResource, "SELECT * FROM appurl WHERE short_url=$1", shortURL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    http.StatusNotFound,
			"message": "Not records found.",
		})
	} else {
		// TODO: add/update click events here
		// Table requires url - date - ip
		origin := c.Request.Header.Get("Origin")
		log.Printf("debug: header origin ip: %v", origin)
		//ip, _ := cs.GetClientIPHelper(c.Request)
		//log.Printf("Here" + ip)
		c.Redirect(http.StatusMovedPermanently, shortenedURLResource.URL)
	}
}
