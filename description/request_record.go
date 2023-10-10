package description

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type RequestRecord struct {
	Cookies []http.Cookie     `json:"cookies"`
	Headers map[string]string `json:"headers"`
}

func (r *RequestRecord) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	temp := make([]TempCookie, 0)
	err = json.Unmarshal(raw["cookies"], &temp)
	if err != nil {
		return err
	}
	for _, v := range temp {
		r.Cookies = append(r.Cookies, v.ToNetCookie())
	}

	err = json.Unmarshal(raw["headers"], &r.Headers)
	if err != nil {
		return err
	}
	return nil
}

type Description struct {
	Images   []url.URL                `json:"images"`
	Requests map[string]RequestRecord `json:"requests"`
}

func (d *Description) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	temp := make([]string, 0)
	err = json.Unmarshal(raw["images"], &temp)
	if err != nil {
		return err
	}
	for _, v := range temp {
		u, err := url.Parse(v)
		if err != nil {
			return err
		}
		d.Images = append(d.Images, *u)
	}

	err = json.Unmarshal(raw["requests"], &d.Requests)
	if err != nil {
		return err
	}
	return nil
}
