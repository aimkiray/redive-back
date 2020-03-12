package api

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/aimkiray/reosu-server/utils"
)

type Song struct {
	Data struct {
		Url string `json:"url"`
	} `json:"data"`
	Code int `json:"code"`
}

type Lyric struct {
	Lrc struct {
		Lyric string `json:"lyric"`
	} `json:"lrc"`
	Tlyric struct {
		Lyric string `json:"lyric"`
	} `json:"tlyric"`
	Code int `json:"code"`
}

//获取歌单内容，返回歌单内歌曲的简略信息
//GET方式
//必选参数 id : 歌单 id
//可选参数 s : 歌单最近的 s 个收藏者
func GetPlayList(c *gin.Context) {
	id := c.Query("id")
	s := c.DefaultQuery("s", "8")
	save := c.DefaultQuery("save", "0")

	params, encSecKey, encErr := utils.Encrypt(fmt.Sprintf(`{"id":"%s","s":%s,"n":10000,"csrf_token":""}`, id, s))
	if encErr != nil {
		log.Println(encErr)
	}
	resBody, resErr := utils.DoPostRequest(c.Request.Cookies(), "https://music.163.com/weapi/v3/playlist/detail", params, encSecKey)
	if resErr != nil {
		log.Println(resErr)
	}
	if save == "1" {

	}
	c.String(200, string(resBody))
}

//获取歌曲详情，可返回歌曲专辑封面
//GET方式
//必选参数 ids : 歌曲 id，可传入多个，以英文逗号分隔
func GetSongDetail(c *gin.Context) {
	ids := c.Query("ids")
	idsList := strings.Split(ids, ",")
	save := c.DefaultQuery("save", "0")

	idJson := ""
	for id := 0; id < len(idsList); id++ {
		idJson += fmt.Sprintf(`{\"id\":\"%s\"},`, idsList[id])
	}
	idJson = idJson[0 : len(idJson)-1]

	params, encSecKey, encErr := utils.Encrypt(fmt.Sprintf(`{"id":"%s","c":"[%s]","csrf_token":""}`, ids, idJson))
	if encErr != nil {
		log.Println(encErr)
	}
	resBody, resErr := utils.DoPostRequest(c.Request.Cookies(), "https://music.163.com/weapi/v3/song/detail", params, encSecKey)
	if resErr != nil {
		log.Println(resErr)
	}
	if save == "1" {

	}
	c.String(200, string(resBody))
}

//获取歌曲直链
//GET方式
//必选参数 id : 歌曲 id
//可选参数 br : 比特率
func GetSongURL(c *gin.Context) {
	id := c.Query("id")
	br := c.DefaultQuery("br", "999000")
	save := c.DefaultQuery("save", "0")

	params, encSecKey, encErr := utils.Encrypt(fmt.Sprintf(`{"ids":[%s],"br":"%s","csrf_token":""}`,
		id, br))
	if encErr != nil {
		log.Println(encErr)
	}
	resBody, resErr := utils.DoPostRequest(c.Request.Cookies(), "https://music.163.com/weapi/song/enhance/player/url", params, encSecKey)
	if resErr != nil {
		log.Println(resErr)
	}
	if save == "1" {

	}
	c.String(200, string(resBody))
}

//获取歌词，返回歌词和翻译（如果有的话）
//GET方式
//必选参数 id : 歌曲 id
func GetLyric(c *gin.Context) {
	id := c.Query("id")
	c.String(200, doLyric(id, c, ""))
}

func doLyric(id string, c *gin.Context, filename string) string {
	params, encSecKey, encErr := utils.Encrypt(fmt.Sprintf(`{"id":"%s","lv":"-1","kv":"-1","tv":"-1","csrf_token":""}`, id))
	if encErr != nil {
		log.Println(encErr)
	}
	resBody, resErr := utils.DoPostRequest(c.Request.Cookies(), "https://music.163.com/weapi/song/lyric", params, encSecKey)
	if resErr != nil {
		log.Println(resErr)
	}
	if filename != "" {
		var lrc Lyric
		jsonErr := json.Unmarshal(resBody, &lrc)
		if jsonErr != nil {
			log.Println(jsonErr)
		}
	}
	return string(resBody)
}
