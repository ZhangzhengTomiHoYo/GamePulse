package controllers

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

const CtxUserIDKey = "userID"

var ErrorUserNotLogin = errors.New("鐢ㄦ埛鏈櫥褰?")

// getCurrentUserID gets the authenticated user ID from gin.Context.
func getCurrentUserID(c *gin.Context) (userID int64, err error) {
	uid, ok := c.Get(CtxUserIDKey)
	if !ok {
		return 0, ErrorUserNotLogin
	}

	switch v := uid.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("invalid user id type: %T", uid)
	}
}
