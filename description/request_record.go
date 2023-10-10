package description

import (
	"encoding/json"
	"net/url"
)

type Url struct {
	Url url.URL
}

type RequestRecord struct {
	Cookies []TempCookie      `json:"cookies"`
	Headers map[string]string `json:"headers"`
}

type Description struct {
	Images   []Url                    `json:"images"`
	Requests map[string]RequestRecord `json:"requests"`
}

func (u *Url) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	rawUrl, err := url.Parse(s)
	u.Url = *rawUrl
	return err
}
