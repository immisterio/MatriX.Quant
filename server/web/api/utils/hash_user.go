package utils

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"server/web/auth"
)

// SplitHashUser separates hash and user values from a combined "hash:user" string.
// If no user is provided, defaultUser is returned.
func SplitHashUser(hashWithUser, defaultUser string) (hash string, user string) {
	hash = hashWithUser
	user = defaultUser

	parts := strings.SplitN(hashWithUser, ":", 2)
	if len(parts) == 2 {
		hash = parts[0]
		if parts[1] != "" {
			user = parts[1]
		}
	}

	return hash, user
}

// JoinHashUser returns hash formatted together with user.
func JoinHashUser(hash, user string) string {
	if hash == "" {
		return ""
	}
	return fmt.Sprintf("%s:%s", hash, user)
}

// ResolveHashUser splits the hash and validates that the requested user is authorized.
// When authentication is required, the user is resolved in the following order:
//  1. Authenticated user from the request context (Authorization header).
//  2. User specified in the hash (format: "hash:user").
//  3. defaultUser (only when authentication is not required).
//
// Authorization errors are returned only when the user is missing both in the
// Authorization header and the hash.
func ResolveHashUser(c *gin.Context, hashWithUser, defaultUser string) (hash string, user string, ok bool) {
	hash, userFromHash := SplitHashUser(hashWithUser, "")
	authRequired := c.GetBool("auth_required")
	authUser := c.GetString(gin.AuthUserKey)

	switch {
	case authUser != "":
		user = authUser
	case userFromHash != "":
		user = userFromHash
	case !authRequired:
		user = defaultUser
	default:
		return "", "", false
	}

	if !authRequired {
		return hash, user, true
	}

	if user == "" || !auth.UserExists(user) {
		return "", "", false
	}

	return hash, user, true
}
