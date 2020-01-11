package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"

	hs "github.com/jessequinn/linkshortener/handlers"
	mdl "github.com/jessequinn/linkshortener/models"
)

// Database middleware
func Database(con string) gin.HandlerFunc {
	db, err := sqlx.Connect("sqlite3", con) // temporarily use sqlite3
	if err != nil {
		log.Fatalln(err)
	}
	db.MustExec(mdl.UrlSchema)

	return func(c *gin.Context) {
		c.Set("DB", db)
		c.Next()
	}
}

func main() {
	r := gin.Default()
	r.Use(Database("./linkshortener.db"))

	// Health endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"serverTime": time.Now().UTC(),
		})
	})
	//
	r.GET("/v1/urls/:short_url", hs.GetShortUrl)
	r.POST("/v1/urls", hs.CreateShortUrl)
	r.Run(":8000") // Listen and serve on 0.0.0.0:8080
}
