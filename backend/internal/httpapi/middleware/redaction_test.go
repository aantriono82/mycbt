package middleware

import "testing"

func TestRedactQueryString(t *testing.T) {
	got := redactQueryString("access_token=abc&x=1&token=ZZZ")
	if got != "access_token=%5BREDACTED%5D&token=%5BREDACTED%5D&x=1" && got != "access_token=%5BREDACTED%5D&x=1&token=%5BREDACTED%5D" {
		// url.Values.Encode sorts by key; accept either ordering for future changes.
		t.Fatalf("unexpected redacted query: %q", got)
	}
}

func TestRedactBodyJSON(t *testing.T) {
	raw := []byte(`{"token":"ABC123","nested":{"password":"p","keep":1},"arr":[{"api_key":"k"}]}`)
	got := redactBody("application/json", raw)
	want1 := `{"arr":[{"api_key":"[REDACTED]"}],"nested":{"keep":1,"password":"[REDACTED]"},"token":"[REDACTED]"}`
	if got != want1 {
		t.Fatalf("unexpected redacted body: %q", got)
	}
}

func TestRedactBodyForm(t *testing.T) {
	raw := []byte("token=ABC&x=1&password=secret")
	got := redactBody("application/x-www-form-urlencoded", raw)
	if got != "password=%5BREDACTED%5D&token=%5BREDACTED%5D&x=1" && got != "token=%5BREDACTED%5D&password=%5BREDACTED%5D&x=1" {
		t.Fatalf("unexpected redacted form: %q", got)
	}
}

