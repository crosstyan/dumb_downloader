package description

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

type RequestRecord struct {
	Timestamp time.Time         `json:"timestamp"`
	Cookies   []http.Cookie     `json:"cookies"`
	Headers   map[string]string `json:"headers"`
}

func (r *RequestRecord) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}

	timeTemp := ""
	err = json.Unmarshal(raw["timestamp"], &timeTemp)
	if err != nil {
		return err
	}
	t, err := time.Parse(time.RFC3339, timeTemp)
	if err != nil {
		return err
	}
	r.Timestamp = t

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
