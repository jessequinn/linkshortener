package main

import (
	"github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
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
	db.MustExec(mdl.UserSchema)
	return func(c *gin.Context) {
		c.Set("DB", db)
		c.Next()
	}
}

// JWT
type User struct {
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

var identityKey = "id"

func main() {
	port := os.Getenv("PORT")
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	if port == "" {
		port = "8000" // Listen and serve on 0.0.0.0:8080
	}
	r.Use(Database("./linkshortener.db"))
	// the jwt middleware
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "test zone",
		Key:         []byte("secret key"),
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*User); ok {
				return jwt.MapClaims{
					identityKey: v.Username,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &User{
				Username: claims[identityKey].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals login
			var user User
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			user.Username = loginVals.Username
			user.Password = loginVals.Password
			db := c.MustGet("DB").(*sqlx.DB)
			rows, err := db.NamedQuery(`SELECT * FROM user WHERE username=:username AND password=:password`, user)
			defer rows.Close()
			if err != nil {
				log.Println(err)
				c.String(http.StatusInternalServerError, err.Error())
			} else {
				if rows.Next() == false {
					return nil, jwt.ErrFailedAuthentication
				} else {
					return &User{
						Username: user.Username,
					}, nil
				}
			}
			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if v, ok := data.(*User); ok && v.Username == "admin" {
				return true
			}

			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	})
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}
	// Health endpoint
	r.GET("/health", hs.Health)
	// Login
	r.POST("/login", authMiddleware.LoginHandler)
	r.POST("/register", hs.RegisterUser)
	// No route
	r.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
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
