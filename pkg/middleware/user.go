package middleware

import (
	"net/http"

	"github.com/alikarimi999/shahboard/pkg/jwt"
	"github.com/alikarimi999/shahboard/types"
	"github.com/gin-gonic/gin"
)

const userKey = "user"

func ParsUserHeader(v *jwt.Validator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.GetHeader("Authorization")
		if token == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		token = token[len("Bearer "):]
		if token == "" {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		user, err := v.ValidateJWT(token)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set(userKey, user)
		ctx.Next()
	}
}

func ExtractUser(ctx *gin.Context) (types.User, bool) {
	u, ok := ctx.Get(userKey)
	if !ok {
		return types.User{}, false
	}
	return u.(types.User), true
}
