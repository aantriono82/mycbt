package params

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func IntQuery(c *gin.Context, key string, def, min, max int) int {
	v := c.Query(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	if i < min {
		return min
	}
	if i > max {
		return max
	}
	return i
}

func StringQueryTrim(c *gin.Context, key string) string {
	v := c.Query(key)
	return trim(v)
}

func trim(s string) string {
	// Avoid importing strings everywhere (small helper).
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\n' || s[0] == '\r') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t' || s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}
