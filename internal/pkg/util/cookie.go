package util

import "net/http"

// CookiesToMap convert cookies to map
func CookiesToMap(cookies []*http.Cookie) map[string]string {
	cookieMap := make(map[string]string, len(cookies))
	for _, c := range cookies {
		cookieMap[c.Name] = c.Value
	}
	return cookieMap
}
