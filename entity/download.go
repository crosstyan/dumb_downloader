package entity

import (
	"encoding/json"
	"net/http"
)

type DownloadRequest struct {
	Url     string        `json:"url"`
	Cookies []http.Cookie `json:"cookies"`
	// recommended to remove "User-Agent" from headers
	Headers map[string]string `json:"headers"`
	// if it not exists it won't be saved
	// if it's empty then it would be saved at root of output directory
	// otherwise it would be saved at `output_dir/out_prefix`
	OutPrefix *string `json:"out_prefix,omitempty"`
}

type DownloadResponse struct {
	Url        string            `json:"url"`
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	// if it's binary, it's base64 encoded
	// if it's text, it's utf-8 encoded
	Body string `json:"body"`
}

type ErrorResponse struct {
	Error error `json:"error"`
}

func (e ErrorResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Error string `json:"error"`
	}{
		Error: e.Error.Error(),
	})
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
