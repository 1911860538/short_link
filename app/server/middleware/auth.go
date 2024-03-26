package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/1911860538/short_link/config"
)

var (
	confJwtSecretKeyBytes = []byte(config.Conf.Jwt.SecretKey)
	confJwtAlgo           = config.Conf.Jwt.Algo

	errTokenInvalid = errors.New("jwt is invalid")
	errTokenAlgo    = errors.New("jwt algo is invalid")
	errTokenExpired = errors.New("jwt is expired")
)

type JwtClaims struct {
	jwt.RegisteredClaims

	Id       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

func JwtMiddleware(c *gin.Context) {
	headerAuth := c.Request.Header.Get("Authorization")
	if headerAuth == "" {
		abort401(c, "Authorization header missing")
		return
	}

	splitToken := strings.Split(headerAuth, "Bearer ")
	if len(splitToken) != 2 {
		abort401(c, "Invalid Authorization header format")
		return
	}

	tokenString := splitToken[1]

	claims, err := validateJwt(tokenString)
	if err != nil {
		abort401(c, err.Error())
		return
	}

	c.Set("jwt", claims)
	c.Next()
}

func validateJwt(tokenString string) (*JwtClaims, error) {
	var myClaims JwtClaims
	token, err := jwt.ParseWithClaims(tokenString, &myClaims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != confJwtAlgo {
			return nil, errTokenAlgo
		}
		return confJwtSecretKeyBytes, nil
	})

	if err != nil {
		return nil, err
	}

	validClaims, ok := token.Claims.(*JwtClaims)
	if !ok || !token.Valid {
		return nil, errTokenInvalid
	}

	if validClaims.ExpiresAt.Before(time.Now()) {
		return nil, errTokenExpired
	}

	return validClaims, nil
}

func abort401(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"detail": msg,
	})
}
