package entity

import (
	"encoding/json"
	"net/http"
)

type DownloadRequest struct {
	Url     string        `json:"url"`
	Cookies []http.Cookie `json:"cookies"`
	// recommended to remove "User-Agent" from headers
	Headers      map[string]string `json:"headers"`
	IsSaveOutput bool              `json:"is_save_output"`
	OutPrefix    string            `json:"out_prefix"`
}

type DownloadResponse struct {
	Url        string            `json:"url"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	// if it's binary, it's base64 encoded
	// if it's text, it's utf-8 encoded
	Body []byte `json:"body"`
}

// UnmarshalJSON
//
// @see https://gist.github.com/miguelmota/904f0fdad34eaac09c5d53098f960c5c
func (r *DownloadRequest) UnmarshalJSON(data []byte) error {
	type Alias DownloadRequest
	aux := &struct {
		Cookies []TempCookie `json:"cookies"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	cookies := make([]http.Cookie, len(aux.Cookies))
	for i, c := range aux.Cookies {
		cookies[i] = c.ToNetCookie()
	}
	r.Cookies = cookies
	return nil
}
