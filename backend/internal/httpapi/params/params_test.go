package params

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIntQuery(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/?limit=999&offset=-1&bad=x", nil)

	if got := IntQuery(c, "limit", 50, 1, 200); got != 200 {
		t.Fatalf("expected capped limit 200, got %d", got)
	}
	if got := IntQuery(c, "offset", 0, 0, 1000); got != 0 {
		t.Fatalf("expected clamped offset 0, got %d", got)
	}
	if got := IntQuery(c, "bad", 7, 1, 10); got != 7 {
		t.Fatalf("expected default for bad int, got %d", got)
	}
	if got := IntQuery(c, "missing", 9, 1, 10); got != 9 {
		t.Fatalf("expected default for missing int, got %d", got)
	}
}

func TestStringQueryTrimAndTrim(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/?q=%20%20hello%0A", nil)

	if got := StringQueryTrim(c, "q"); got != "hello" {
		t.Fatalf("expected trimmed query, got %q", got)
	}
	if got := trim("\t hello \r\n"); got != "hello" {
		t.Fatalf("expected trimmed helper, got %q", got)
	}
}
