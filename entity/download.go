package entity

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
)

type DownloadRequest struct {
	Url string `json:"url" example:"https://example.com/"`
	// Array of cookies. See also `entity.TempCookie`.
	//
	// https://chromedevtools.github.io/devtools-protocol/tot/Network/#type-Cookie
	Cookies []http.Cookie `json:"cookies" swaggertype:"array,object"`
	// recommended to remove "User-Agent" from headers
	Headers map[string]string `json:"headers"`
	// if it not exists it won't be saved.
	// if it's empty then it would be saved at root of output directory.
	// Otherwise, it would be saved at `output_dir/out_prefix`
	OutPrefix *string `json:"out_prefix,omitempty" example:"example"`
}

type DownloadResponse struct {
	Url        string            `json:"url" example:"https://example.com/"`
	StatusCode int               `json:"status_code" example:"200"`
	Headers    map[string]string `json:"headers"`
	MIMEType   string            `json:"mime_type" example:"text/html"`
	// if it's binary, it's base64 encoded. Otherwise,
	// it's text
	Body []byte `json:"body,omitempty" example:"<html>...</html>" swaggertype:"string"`
}

func (r *DownloadResponse) MarshalJSON() ([]byte, error) {
	type Alias DownloadResponse
	aux := &struct {
		Body *string `json:"body,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if r.Body == nil {
		return json.Marshal(aux)
	}
	if len(r.Body) == 0 {
		return json.Marshal(aux)
	}
	// https://github.com/edn-format/edn/issues/43
	validContains := []string{"text", "json", "edn", "xml", "html", "svg", "css", "javascript", "ecmascript", "x-www-form-urlencoded"}
	isText := func() bool {
		for _, v := range validContains {
			if strings.Contains(r.MIMEType, v) {
				return true
			}
		}
		return false
	}()
	if isText {
		s := string(r.Body)
		aux.Body = &s
	} else {
		s := base64.StdEncoding.EncodeToString(r.Body)
		aux.Body = &s
	}
	return json.Marshal(aux)
}

type ErrorResponse struct {
	Error string `json:"error" example:"error message"`
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
