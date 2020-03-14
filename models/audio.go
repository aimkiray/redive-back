package models

type Audio struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Artist string `json:"artist"`
	Audio  string `json:"audio"`
	Cover  string `json:"cover"`
	Lrc    string `json:"lrc"`
	Tlrc   string `json:"tlrc"`
	From   string `json:"from"`
	Create string `json:"create"`
	Others string `json:"others"`
}
