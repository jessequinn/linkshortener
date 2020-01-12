package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
)

type User struct {
	Username string `json:"username" db:"username"`
	Password string `json:"password" db:"password"`
}

type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

var identityKey = "id"

func JwtConfigGenerate() *jwt.GinJWTMiddleware {
	authMiddleware := &jwt.GinJWTMiddleware{
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
			user.Password = loginVals.Password // TODO: use bcrypt/argon2i
			db := c.MustGet("DB").(*sqlx.DB)
			rows, err := db.NamedQuery(`SELECT * FROM user WHERE username=:username AND password=:password`, user)
			defer rows.Close()
			if err != nil {
				return nil, jwt.ErrFailedAuthentication // TODO: change to proper error
			} else {
				if rows.Next() == false {
					return nil, jwt.ErrFailedAuthentication
				} else {
					c.Set("user", user.Username)
					c.Next()
					return &User{
						Username: user.Username,
					}, nil
				}
			}
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
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			user := c.MustGet("user")
			db := c.MustGet("DB").(*sqlx.DB)
			tx := db.MustBegin()
			tx.MustExec("UPDATE user SET token=$1, updated_at=$2 WHERE username=$3", token, time.Now().UTC().Format("2006-01-02 15:04:05"), user)
			err := tx.Commit()
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
			} else {
				c.JSON(http.StatusOK, gin.H{
					"code":   http.StatusOK,
					"token":  token,
					"expire": expire.Format(time.RFC3339),
				})
			}
		},
	}
	return authMiddleware
}
