package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/alikarimi999/shahboard/types"
	"github.com/gin-gonic/gin"
)

const userKey = "user"

func ParsUserHeader() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		b64 := ctx.GetHeader("X-User-Data")
		if b64 == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "X-User-Data header is required"})
			ctx.Abort()
			return
		}

		data, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "X-User-Data header is invalid"})
			ctx.Abort()
			return
		}

		var u types.User
		err = json.Unmarshal(data, &u)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "X-User-Data header is invalid"})
			ctx.Abort()
			return
		}

		ctx.Set(userKey, u)
		ctx.Next()
	}
}
