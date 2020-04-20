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

// 批量导入歌单歌曲
// TODO 过长有待拆分（单曲导入）
func BatchDownload(c *gin.Context) {
	id := c.Query("id")
	s := c.DefaultQuery("s", "8")
	br := c.DefaultQuery("br", "999000")

	// record error message
	utils.Client.Del("error")
	var errSongList []map[string]interface{}
	//errSongList := make([]map[string]string, 0)

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

	createTime := time.Now().Format("2006/1/2 15:04:05")

	// 生成Playlist
	plName := playlist.Playlist.Name
	playlistInfo := make(map[string]interface{})
	playlistInfo["id"] = id
	playlistInfo["name"] = plName
	playlistInfo["create"] = createTime

	// 保存playlist元数据的key，便于提取
	// TODO 但是复杂度增加了。可改用SCAN
	utils.Client.LPush("playlist", "pl:"+id)
	// 记录playlist信息
	utils.Client.HMSet("pl:"+id, playlistInfo)

	// 保存playlist所包含歌曲的key，便于提取
	//utils.Client.LPush("audio", "pla:"+id)

	baseFileDir := conf.FileDIR + "/music/" + id + "/"

	tracks := playlist.Playlist.Tracks
	songTotal := len(tracks)

	// File download process
	SetStatus(0, songTotal, "", 0, nil)

	// Loop download
	for index, value := range tracks {
		songID := strconv.Itoa(value.ID)
		songName := value.Name
		songArtist := value.Ar[0].Name

		// Check song exists
		audioList := utils.Client.LRange("audio-list", 0, -1)
		exist := utils.InList(audioList.Val(), songID)
		if exist {
			continue
		}

		songDir := baseFileDir + songArtist + " - " + songName + "/"
		os.MkdirAll(songDir, os.ModePerm)

		// Get song url
		resSongURL := GetSongURL(songID, br)
		var song models.Song
		jsonErr = json.Unmarshal(resSongURL, &song)
		if jsonErr != nil {
			errSong := make(map[string]interface{})
			errSong["id"] = songID
			errSong["name"] = songName
			errSong["type"] = "song json"
			errSongList = append(errSongList, errSong)

			continue
		}
		// Save mp3 file
		songPath := songDir + songArtist + " - " + songName + ".mp3"
		songURL := song.Data[0].Url

		// Song is unavailable
		if songURL == "" {
			errSong := make(map[string]interface{})
			errSong["id"] = songID
			errSong["name"] = songName
			errSong["type"] = "song is unavailable"
			errSongList = append(errSongList, errSong)

			continue
		}
		fileErr := utils.DownloadURL(songURL, songPath)
		if fileErr != nil {
			errSong := make(map[string]interface{})
			errSong["id"] = songID
			errSong["name"] = songName
			errSong["type"] = "song file"
			errSongList = append(errSongList, errSong)

			continue
		}

		// Get album cover
		alURL := value.Al.PicUrl
		fileFrag := strings.Split(alURL, ".")
		alSuffix := fileFrag[len(fileFrag)-1]

		// Save album cover
		alPath := songDir + songArtist + " - " + songName + "." + alSuffix
		fileErr = utils.DownloadURL(alURL, alPath)
		if fileErr != nil {
			errSong := make(map[string]interface{})
			errSong["id"] = songID
			errSong["name"] = songName
			errSong["type"] = "cover file"
			errSongList = append(errSongList, errSong)

			continue
		}

		// Get song lyric
		resLyric := GetLyric(songID)

		var lyric models.Lyric
		jsonErr := json.Unmarshal(resLyric, &lyric)
		if jsonErr != nil {
			errSong := make(map[string]interface{})
			errSong["id"] = songID
			errSong["name"] = songName
			errSong["type"] = "lyric json"
			errSongList = append(errSongList, errSong)

			continue
		}
		// Save lyric
		lyricPath := songDir + songArtist + " - " + songName + ".lrc"
		fileErr = utils.DownloadText(lyric.Lrc.Lyric, lyricPath)
		if fileErr != nil {
			errSong := make(map[string]interface{})
			errSong["id"] = songID
			errSong["name"] = songName
			errSong["type"] = "lyric file"
			errSongList = append(errSongList, errSong)

			continue
		}
		// Save translate lyric
		tlyricPath := songDir + songArtist + " - " + songName + "_trans.lrc"
		fileErr = utils.DownloadText(lyric.Tlyric.Lyric, tlyricPath)
		if fileErr != nil {
			errSong := make(map[string]interface{})
			errSong["id"] = songID
			errSong["name"] = songName
			errSong["type"] = "translate lyric file"
			errSongList = append(errSongList, errSong)

			continue
		}

		// Redis insert
		songInfo := make(map[string]interface{})

		songInfo["id"] = songID
		songInfo["name"] = songName
		songInfo["playlist"] = id
		songInfo["artist"] = value.Ar[0].Name
		songInfo["audio"] = songPath
		songInfo["cover"] = alPath
		songInfo["lrc"] = lyricPath
		songInfo["tlrc"] = tlyricPath
		songInfo["create"] = createTime
		songInfo["from"] = "batch"
		songInfo["others"] = ""

		// 保存歌曲key，便于取用
		utils.Client.LPush("pla:"+id, "au:"+songID)
		// 记录歌曲信息
		utils.Client.HMSet("au:"+songID, songInfo)

		SetStatus(index, songTotal, songName, 0, nil)
	}

	// return status
	SetStatus(0, 0, "", 1, errSongList)
	c.JSON(http.StatusOK, gin.H{
		"code":    1,
		"message": "batch import success",
	})
}

func BatchStatus(c *gin.Context) {
	res := utils.Client.HGetAll("batch-import").Val()
	// read error list
	var errSongList []map[string]string
	if res["status"] == "1" {
		err := utils.Client.LRange("error", 0, -1).Val()
		errSongList = make([]map[string]string, len(err))
		for index, value := range err {
			errSong := utils.Client.HGetAll(value).Val()
			errSongList[index] = errSong
		}
		utils.Client.HSet("batch-import", "status", 0)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  1,
		"data":  res,
		"error": errSongList,
	})
}

func SetStatus(count int, total int, current string, status int, err []map[string]interface{}) {
	fileProcess := make(map[string]interface{})
	fileProcess["count"] = count
	fileProcess["total"] = total
	fileProcess["current"] = current
	fileProcess["status"] = status

	if err != nil {
		for _, value := range err {
			id := value["id"].(string)
			utils.Client.LPush("error", "err:"+id)
			utils.Client.HMSet("err:"+id, value)
		}
	}

	utils.Client.HMSet("batch-import", fileProcess)
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

	for _, id := range idsList {
		idJson += fmt.Sprintf(`{\"id\":\"%s\"},`, id)
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
