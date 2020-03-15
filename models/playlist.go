package models

type Playlist struct {
	Playlist struct {
		Tracks []struct {
			Name string `json:"name"`
			ID   int    `json:"id"`
			Ar   []struct {
				Name string `json:"name"`
			} `json:"ar"`
			Al struct {
				PicUrl string `json:"picUrl"`
			} `json:"al"`
		} `json:"tracks"`
		Name string `json:"name"`
	} `json:"playlist"`
	Code int `json:"code"`
}
