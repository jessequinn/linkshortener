package main

import (
	"github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	//"time"

	hs "github.com/jessequinn/linkshortener/handlers"
	mw "github.com/jessequinn/linkshortener/middlewares"
	mdl "github.com/jessequinn/linkshortener/models"
)

// Database middlewares
func Database(con string) gin.HandlerFunc {
	db, err := sqlx.Connect("sqlite3", con) // temporarily use sqlite3
	if err != nil {
		log.Fatalln(err)
	}
	db.MustExec(mdl.UrlSchema)
	db.MustExec(mdl.UserSchema)
	return func(c *gin.Context) {
		c.Set("DB", db)
		c.Next()
	}
}

func main() {
	port := os.Getenv("PORT")
	// Production
	//gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	if port == "" {
		port = "8080"
	}
	r.Use(Database("./linkshortener.db"))
	authMiddleware, err := jwt.New(mw.JwtConfigGenerate())
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", hs.Health)
		v1.POST("/login", authMiddleware.LoginHandler)
		v1.POST("/register", hs.RegisterUser)
	}
	// No route
	r.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": http.StatusNotFound, "message": "Route not found"})
	})
	auth := r.Group("/auth")
	auth.GET("/refresh_token", authMiddleware.RefreshHandler)
	auth.Use(authMiddleware.MiddlewareFunc())
	{
		auth.GET("/v1/urls/:short_url", hs.GetShortUrl)
		auth.POST("/v1/urls", hs.CreateShortUrl)
		auth.DELETE("/v1/urls/:short_url", hs.RemoveShortUrl)
		auth.PATCH("/v1/urls/:short_url", hs.UpdateShortUrl)
	}
	r.Run(":" + port)
}
