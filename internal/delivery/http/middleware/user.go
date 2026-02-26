package middleware

import (
	"github.com/gin-gonic/gin"
)

const userIDContextKey = "user_id"

func setUserID(c *gin.Context, userID int64) {
	c.Set(userIDContextKey, userID)
}

// UserIDFromContext возвращает user id, ранее установленный middleware.
func UserIDFromContext(c *gin.Context) (int64, bool) {
	value, ok := c.Get(userIDContextKey)
	if !ok {
		return 0, false
	}
	id, ok := value.(int64)
	return id, ok
}
