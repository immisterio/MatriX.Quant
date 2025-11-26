package utils

import "github.com/gin-gonic/gin"

func UserID(c *gin.Context) string {
	if !c.GetBool("auth_required") {
		return "base"
	}

	user, ok := c.Get(gin.AuthUserKey)
	if !ok || user == "" {
		return "base"
	}

	if v, ok := user.(string); ok && v != "" {
		return v
	}

	return "base"
}
