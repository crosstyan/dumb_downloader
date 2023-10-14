package entity

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
	type Alias RequestRecord
	aux := &struct {
		// this is the temporary field that json.Unmarshal know how to parse
		Timestamp string       `json:"timestamp"`
		Cookies   []TempCookie `json:"cookies"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	t, err := time.Parse(time.RFC3339, aux.Timestamp)
	if err != nil {
		return err
	}
	r.Timestamp = t
	cookies := make([]http.Cookie, len(aux.Cookies))
	for i, c := range aux.Cookies {
		cookies[i] = c.ToNetCookie()
	}
	r.Cookies = cookies
	return nil
}

type Description struct {
	Links    []url.URL                `json:"links"`
	Requests map[string]RequestRecord `json:"requests"`
}

func (d *Description) UnmarshalJSON(data []byte) error {
	type Alias Description
	aux := &struct {
		Links []string `json:"links"`
		*Alias
	}{
		Alias: (*Alias)(d),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	links := make([]url.URL, len(aux.Links))
	for i, l := range aux.Links {
		u, err := url.Parse(l)
		if err != nil {
			return err
		}
		links[i] = *u
	}
	d.Links = links
	return nil
}
