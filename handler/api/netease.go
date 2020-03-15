package api

import (
	"encoding/json"
	"fmt"
	"github.com/aimkiray/reosu-server/models"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aimkiray/reosu-server/conf"
	"github.com/aimkiray/reosu-server/utils"
)

func BatchDownload(c *gin.Context) {
	id := c.Query("id")
	s := c.DefaultQuery("s", "8")
	br := c.DefaultQuery("br", "999000")

	resPlay := GetPlayList(id, s)

	var playlist models.Playlist
	jsonErr := json.Unmarshal(resPlay, &playlist)
	if jsonErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    0,
			"message": "playlist json, " + jsonErr.Error(),
		})
		return
	}
	alName := playlist.Playlist.Name
	baseFileDir := conf.FileDIR + "/music/" + strings.Replace(alName, "/", "*", -1) + "/"
	createTime := time.Now().Format("2006/1/2 15:04:05")

	tracks := playlist.Playlist.Tracks
	songTotal := len(tracks)

	utils.Client.LPush("file-process", "download")

	// File download process
	SetStatus(0, songTotal, "")

	// Loop download
	for index := 0; index < songTotal; index++ {
		songID := strconv.Itoa(tracks[index].ID)
		songName := tracks[index].Name
		songArtist := tracks[index].Ar[0].Name

		// Check song exists
		audioList := utils.Client.LRange("audio-list", 0, -1)
		exist := utils.InList(audioList.Val(), songID)
		if exist {
			continue
		}

		// Get album cover
		alURL := tracks[index].Al.PicUrl
		fileFrag := strings.Split(alURL, ".")
		alSuffix := fileFrag[len(fileFrag)-1]
		songDir := baseFileDir + songArtist + " - " + songName + "/"
		os.MkdirAll(songDir, os.ModePerm)
		// Save album cover
		alPath := songDir + songArtist + " - " + songName + "." + alSuffix
		fileErr := utils.DownloadURL(alURL, alPath)
		if fileErr != nil {
			SetStatus(0, 0, "")
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    0,
				"message": "album cover, id=" + songID + fileErr.Error(),
			})
			continue
		}

		// Get song lyric
		resLyric := GetLyric(songID)

		var lyric models.Lyric
		jsonErr := json.Unmarshal(resLyric, &lyric)
		if jsonErr != nil {
			SetStatus(0, 0, "")
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    0,
				"message": "lyric json, id=" + songID + jsonErr.Error(),
			})
			continue
		}
		// Save lyric
		lyricPath := songDir + songArtist + " - " + songName + ".lrc"
		fileErr = utils.DownloadText(lyric.Lrc.Lyric, lyricPath)
		if fileErr != nil {
			SetStatus(0, 0, "")
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    0,
				"message": "lyric, id=" + songID + fileErr.Error(),
			})
			continue
		}
		// Save translate lyric
		tlyricPath := songDir + songArtist + " - " + songName + "_trans.lrc"
		fileErr = utils.DownloadText(lyric.Tlyric.Lyric, tlyricPath)
		if fileErr != nil {
			SetStatus(0, 0, "")
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    0,
				"message": "translate lyric, id=" + songID + fileErr.Error(),
			})
			continue
		}

		// Get song url
		resSongURL := GetSongURL(songID, br)
		var song models.Song
		jsonErr = json.Unmarshal(resSongURL, &song)
		if jsonErr != nil {
			SetStatus(0, 0, "")
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    0,
				"message": "mp3 json, id=" + songID + jsonErr.Error(),
			})
			continue
		}
		// Save mp3 file
		songPath := songDir + songArtist + " - " + songName + ".mp3"
		fileErr = utils.DownloadURL(song.Data[0].Url, songPath)
		if fileErr != nil {
			SetStatus(0, 0, "")
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    0,
				"message": "mp3, id=" + songID + fileErr.Error(),
			})
			continue
		}

		// Redis insert
		songInfo := make(map[string]interface{})

		songInfo["id"] = songID
		songInfo["name"] = songName
		songInfo["artist"] = tracks[index].Ar[0].Name
		songInfo["audio"] = songPath
		songInfo["cover"] = alPath
		songInfo["lrc"] = lyricPath
		songInfo["tlrc"] = tlyricPath
		songInfo["create"] = createTime
		songInfo["from"] = "batch"
		songInfo["others"] = ""

		utils.Client.LPush("audio-list", songID)
		utils.Client.HMSet(songID, songInfo)

		SetStatus(index, songTotal, songName)
	}

	// reset status
	SetStatus(0, 0, "")
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "batch import success",
	})
}

func BatchStatus(c *gin.Context) {
	res := utils.Client.HGetAll("download").Val()
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": res,
	})
}

func SetStatus(count int, total int, current string) {
	fileProcess := make(map[string]interface{})
	fileProcess["count"] = count
	fileProcess["total"] = total
	fileProcess["current"] = current
	utils.Client.HMSet("download", fileProcess)
}

//获取歌单内容，返回歌单内歌曲的简略信息
//GET方式
//必选参数 id : 歌单 id
//可选参数 s : 歌单最近的 s 个收藏者
func PlayList(c *gin.Context) {
	id := c.Query("id")
	s := c.DefaultQuery("s", "8")

	c.String(200, string(GetPlayList(id, s)))
}

func GetPlayList(id string, s string) []byte {
	params, encSecKey, encErr := utils.Encrypt(fmt.Sprintf(`{"id":"%s","s":%s,"n":10000,"csrf_token":""}`, id, s))
	if encErr != nil {
		log.Println(encErr)
	}
	resBody, resErr := utils.DoPostRequest("https://music.163.com/weapi/v3/playlist/detail", params, encSecKey)
	if resErr != nil {
		log.Println(resErr)
	}
	return resBody
}

//获取歌曲详情，可返回歌曲专辑封面
//GET方式
//必选参数 ids : 歌曲 id，可传入多个，以英文逗号分隔
func SongDetail(c *gin.Context) {
	ids := c.Query("ids")

	c.String(200, string(GetSongDetail(ids)))
}

func GetSongDetail(ids string) []byte {
	idsList := strings.Split(ids, ",")
	idJson := ""

	for id := 0; id < len(idsList); id++ {
		idJson += fmt.Sprintf(`{\"id\":\"%s\"},`, idsList[id])
	}
	idJson = idJson[0 : len(idJson)-1]

	params, encSecKey, encErr := utils.Encrypt(fmt.Sprintf(`{"id":"%s","c":"[%s]","csrf_token":""}`, ids, idJson))
	if encErr != nil {
		log.Println(encErr)
	}
	resBody, resErr := utils.DoPostRequest("https://music.163.com/weapi/v3/song/detail", params, encSecKey)
	if resErr != nil {
		log.Println(resErr)
	}
	return resBody
}

//获取歌曲直链
//GET方式
//必选参数 id : 歌曲 id
//可选参数 br : 比特率
func SongURL(c *gin.Context) {
	id := c.Query("id")
	br := c.DefaultQuery("br", "999000")

	c.String(200, string(GetSongURL(id, br)))
}

func GetSongURL(id string, br string) []byte {
	params, encSecKey, encErr := utils.Encrypt(fmt.Sprintf(`{"ids":[%s],"br":"%s","csrf_token":""}`,
		id, br))
	if encErr != nil {
		log.Println(encErr)
	}
	resBody, resErr := utils.DoPostRequest("https://music.163.com/weapi/song/enhance/player/url", params, encSecKey)
	if resErr != nil {
		log.Println(resErr)
	}
	return resBody
}

//获取歌词，返回歌词和翻译（如果有的话）
//GET方式
//必选参数 id : 歌曲 id
func SongLyric(c *gin.Context) {
	id := c.Query("id")

	c.String(200, string(GetLyric(id)))
}

func GetLyric(id string) []byte {
	params, encSecKey, encErr := utils.Encrypt(fmt.Sprintf(`{"id":"%s","lv":"-1","kv":"-1","tv":"-1","csrf_token":""}`, id))
	if encErr != nil {
		log.Println(encErr)
	}
	resBody, resErr := utils.DoPostRequest("https://music.163.com/weapi/song/lyric", params, encSecKey)
	if resErr != nil {
		log.Println(resErr)
	}
	return resBody
}
