package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/alikarimi999/shahboard/pkg/jwt"
	"github.com/alikarimi999/shahboard/types"
	"github.com/gin-gonic/gin"
)

func ParseQueryToken(v *jwt.Validator) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := ctx.Query("token")
		if token == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
			ctx.Abort()
			return
		}

		user, err := v.ValidateJWT(token)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			ctx.Abort()
			return
		}

		ctx.Set(userKey, user)
		ctx.Next()
	}
}

func parseToken(token string) (types.User, error) {
	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return types.User{}, err
	}

	var u types.User
	err = json.Unmarshal(b, &u)
	return u, err
}
