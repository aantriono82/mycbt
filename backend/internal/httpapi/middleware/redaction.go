package middleware

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strings"
)

var redactedPlaceholder = "[REDACTED]"

// Keys are compared in a case-insensitive manner.
var sensitiveKeys = map[string]struct{}{
	"access_token":   {},
	"api_key":        {},
	"apikey":         {},
	"authorization":  {},
	"client_secret":  {},
	"password":       {},
	"refresh_token":  {},
	"secret":         {},
	"smtp_password":  {},
	"token":          {},
	"whatsapp_api_key": {},
}

func isSensitiveKey(key string) bool {
	k := strings.ToLower(strings.TrimSpace(key))
	_, ok := sensitiveKeys[k]
	return ok
}

func redactQueryString(rawQuery string) string {
	rawQuery = strings.TrimSpace(rawQuery)
	if rawQuery == "" {
		return ""
	}
	q, err := url.ParseQuery(rawQuery)
	if err != nil {
		// Best-effort: if parsing fails, don't log raw query.
		return redactedPlaceholder
	}
	changed := false
	for k := range q {
		if isSensitiveKey(k) {
			q[k] = []string{redactedPlaceholder}
			changed = true
		}
	}
	enc := q.Encode()
	if enc == "" && changed {
		return redactedPlaceholder
	}
	return enc
}

func redactBody(contentType string, raw []byte) string {
	if len(raw) == 0 {
		return ""
	}

	mediaType := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	switch mediaType {
	case "", "application/json", "text/json":
		// Try JSON redaction first
		var v any
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.UseNumber()
		if err := dec.Decode(&v); err == nil {
			redactAny(&v)
			b, err := json.Marshal(v)
			if err == nil {
				return string(b)
			}
			return redactedPlaceholder
		}
		// Fallthrough: if not valid JSON, do not log raw.
		return redactedPlaceholder

	case "application/x-www-form-urlencoded":
		vals, err := url.ParseQuery(string(raw))
		if err != nil {
			return redactedPlaceholder
		}
		for k := range vals {
			if isSensitiveKey(k) {
				vals[k] = []string{redactedPlaceholder}
			}
		}
		return vals.Encode()
	default:
		// Unknown/unsafe content types: don't log raw body.
		return redactedPlaceholder
	}
}

func redactAny(v *any) {
	switch t := (*v).(type) {
	case map[string]any:
		for k, vv := range t {
			if isSensitiveKey(k) {
				t[k] = redactedPlaceholder
				continue
			}
			child := vv
			redactAny(&child)
			t[k] = child
		}
	case []any:
		for i := range t {
			child := t[i]
			redactAny(&child)
			t[i] = child
		}
	default:
		// scalar: no-op
	}
}

