package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// AccessTokenParser описывает минимальный контракт JWT-парсера access токена.
type AccessTokenParser interface {
	Parse(tokenStr, expectedType string) (int64, error)
}

// RequireJWT извлекает user id из Bearer access token.
func RequireJWT(parser AccessTokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		if parser == nil {
			c.AbortWithStatusJSON(500, gin.H{
				"code":  "internal_error",
				"error": "jwt parser is not configured",
			})
			return
		}

		rawAuth := strings.TrimSpace(c.GetHeader("Authorization"))
		if rawAuth == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"code":  "unauthorized",
				"error": "missing Authorization header",
			})
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(rawAuth, bearerPrefix) {
			c.AbortWithStatusJSON(401, gin.H{
				"code":  "unauthorized",
				"error": "invalid Authorization header",
			})
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(rawAuth, bearerPrefix))
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"code":  "unauthorized",
				"error": "empty bearer token",
			})
			return
		}

		userID, err := parser.Parse(token, "access")
		if err != nil || userID <= 0 {
			c.AbortWithStatusJSON(401, gin.H{
				"code":  "unauthorized",
				"error": "invalid token",
			})
			return
		}

		setUserID(c, userID)
		c.Next()
	}
}
