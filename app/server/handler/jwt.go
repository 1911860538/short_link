package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/1911860538/short_link/app/server/middleware"
)

func getJwtClaims(c *gin.Context) (*middleware.JwtClaims, bool) {
	jwtClaimsItf, ok := c.Get("jwt")
	if !ok {
		return nil, false
	}
	jwtClaims, ok := jwtClaimsItf.(*middleware.JwtClaims)
	return jwtClaims, ok
}
