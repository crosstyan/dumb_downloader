package description

import (
	"net/http"
	"time"
)

// a temporary cookie struct
// deserialize from json
type TempCookie struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Domain string `json:"domain"`
	Path   string `json:"path"`
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Set-Cookie
	Expires   float64 `json:"expires"`
	Size      int     `json:"size"`
	HttpOnly  bool    `json:"httpOnly"`
	Secure    bool    `json:"secure"`
	Session   bool    `json:"session"`
	SameParty bool    `json:"sameParty"`
}

func (t TempCookie) ToNetCookie() http.Cookie {
	if t.Expires <= 0 {
		return http.Cookie{
			Name:     t.Name,
			Value:    t.Value,
			Domain:   t.Domain,
			Path:     t.Path,
			HttpOnly: t.HttpOnly,
			Secure:   t.Secure,
			SameSite: http.SameSiteDefaultMode,
		}
	}
	return http.Cookie{
		Name:     t.Name,
		Value:    t.Value,
		Domain:   t.Domain,
		Path:     t.Path,
		Expires:  time.Unix(int64(t.Expires), 0),
		HttpOnly: t.HttpOnly,
		Secure:   t.Secure,
		SameSite: http.SameSiteDefaultMode,
	}
}
