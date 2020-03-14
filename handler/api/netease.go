package api

import (
	"encoding/json"
	"fmt"
	"github.com/aimkiray/reosu-server/models"
	"io"
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

type Song struct {
	Data []struct {
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

type Play struct {
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

func BatchDownload(c *gin.Context) {
	id := c.Query("id")
	s := c.DefaultQuery("s", "8")

	resPlay := GetPlayList(id, s)

	var play Play
	jsonErr := json.Unmarshal(resPlay, &play)
	if jsonErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    0,
			"message": "playlist json, " + jsonErr.Error(),
		})
		return
	}
	baseFileDir := conf.FileDIR + "/music/" + play.Playlist.Name + "/"
	createTime := time.Now().Format("2006/1/2 15:04:05")

	tracks := play.Playlist.Tracks

	utils.Client.LPush("file-process", "download")

	// File download progress
	fileProgress := make(map[string]interface{})
	fileProgress["current"] = 0
	fileProgress["total"] = len(tracks)
	utils.Client.HMSet("download", fileProgress)

	// Loop download
	for index := 0; index < len(tracks); index++ {
		songID := strconv.Itoa(tracks[index].ID)
		songName := tracks[index].Name

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
		songDir := baseFileDir + songName + "/"
		os.MkdirAll(songDir, os.ModePerm)
		// Save album cover
		alPath := songDir + songName + "." + alSuffix
		fileErr := DownloadURL(alURL, alPath)
		if fileErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    0,
				"message": "album cover, id=" + songID + fileErr.Error(),
			})
			return
		}

		// Get song lyric
		resLyric := GetLyric(songID)

		var lyric Lyric
		jsonErr := json.Unmarshal(resLyric, &lyric)
		if jsonErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    0,
				"message": "lyric json, id=" + songID + jsonErr.Error(),
			})
			return
		}
		// Save lyric
		lyricPath := songDir + songName + ".lrc"
		fileErr = DownloadText(lyric.Lrc.Lyric, lyricPath)
		if fileErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    0,
				"message": "lyric, id=" + songID + fileErr.Error(),
			})
			return
		}
		// Save translate lyric
		tlyricPath := songDir + songName + "_trans.lrc"
		fileErr = DownloadText(lyric.Tlyric.Lyric, tlyricPath)
		if fileErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    0,
				"message": "translate lyric, id=" + songID + fileErr.Error(),
			})
			return
		}

		// Get song url
		resSongURL := GetSongURL(songID, "999000")
		var song Song
		jsonErr = json.Unmarshal(resSongURL, &song)
		if jsonErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    0,
				"message": "mp3 json, id=" + songID + jsonErr.Error(),
			})
			return
		}
		// Save mp3 file
		songPath := songDir + songName + ".mp3"
		fileErr = DownloadURL(song.Data[0].Url, songPath)
		if fileErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    0,
				"message": "mp3, id=" + songID + fileErr.Error(),
			})
			return
		}

		// Redis insert
		songModel := &models.Audio{
			ID:     songID,
			Name:   songName,
			Artist: tracks[index].Ar[0].Name,
			Audio:  songPath,
			Cover:  alPath,
			Lrc:    lyricPath,
			Tlrc:   tlyricPath,
			Create: createTime,
			From:   "batch",
			Others: "",
		}
		var songInfo map[string]interface{}

		jsonSong, _ := json.Marshal(songModel)
		jsonErr = json.Unmarshal(jsonSong, &songInfo)
		if jsonErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    0,
				"message": "json song error, id=" + songID + jsonErr.Error(),
			})
			return
		}

		utils.Client.LPush("audio-list", songID)
		utils.Client.HMSet(songID, songInfo)
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "batch import success",
	})
}

func DownloadText(text string, filePath string) error {
	lrcFile, err := os.Create(filePath)
	defer lrcFile.Close()

	if err != nil {
		return err
	}

	_, err = lrcFile.Write([]byte(text))
	if err != nil {
		return err
	}
	return nil
}

func DownloadURL(fileURL string, filePath string) error {
	// Get data
	resp, err := http.Get(fileURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create a file
	outfile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer outfile.Close()

	// Response stream and file stream
	_, err = io.Copy(outfile, resp.Body)
	if err != nil {
		return err
	}
	return nil
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
