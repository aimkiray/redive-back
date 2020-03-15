package models

type Song struct {
	Data []struct {
		Url string `json:"url"`
	} `json:"data"`
	Code int `json:"code"`
}
