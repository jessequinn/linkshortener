package main

import (
	"fmt"
	"github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
	"os"

	hs "github.com/jessequinn/linkshortener/handlers"
	mw "github.com/jessequinn/linkshortener/middlewares"
	mdl "github.com/jessequinn/linkshortener/models"
)

// Database middlewares
func Database(con string) gin.HandlerFunc {
	db, err := sqlx.Connect("postgres", con)
	if err != nil {
		log.Fatalln(err)
	}
	db.MustExec(mdl.UserSchema)
	db.MustExec(mdl.URLSchema)
	return func(c *gin.Context) {
		c.Set("DB", db)
		c.Next()
	}
}

func main() {
	port := os.Getenv("PORT")
	dbHost := os.Getenv("DB_HOST")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbPort := os.Getenv("DB_PORT")
	// Production
	//gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	if port == "" {
		port = "8080"
	}
	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbName == "" {
		dbName = "postgres"
	}
	if dbUser == "" {
		dbUser = "postgres"
	}
	if dbPass == "" {
		dbPass = "Ceihohch8ait5"
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)
	r.Use(Database(psqlInfo))
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
	//r.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
	//	claims := jwt.ExtractClaims(c)
	//	log.Printf("NoRoute claims: %#v\n", claims)
	//	c.JSON(404, gin.H{"code": http.StatusNotFound, "message": "Route not found"})
	//})
	r.NoRoute(hs.Redirect)

	auth := r.Group("/auth")
	auth.GET("/refresh_token", authMiddleware.RefreshHandler)
	auth.Use(authMiddleware.MiddlewareFunc())
	{
		auth.GET("/v1/urls/:short_url", hs.GetShortURL)
		auth.POST("/v1/urls", hs.CreateShortURL)
		auth.DELETE("/v1/urls/:short_url", hs.RemoveShortURL)
		auth.PATCH("/v1/urls/:short_url", hs.UpdateShortURL)
	}
	r.Run(":" + port)
}
