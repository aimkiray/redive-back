package api

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aimkiray/reosu-server/conf"
	"github.com/aimkiray/reosu-server/utils"
)

//获取音频列表
func GetAllAudio(c *gin.Context) {
	plID := c.Query("id")
	// 默认查询第一个歌单
	if plID == "" {
		plID = strings.Split(utils.Client.LRange("playlist", 0, 0).Val()[0], ":")[1]
	}
	audioList := utils.Client.LRange("pla:"+plID, 0, -1)

	infoList := make([]map[string]string, len(audioList.Val()))

	for index, value := range audioList.Val() {
		res := utils.Client.HGetAll(value).Val()
		infoList[index] = res
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": infoList,
		"id":   plID,
	})
}

//删除音频信息
func DeleteAudio(c *gin.Context) {
	id := c.Param("id")
	//plID := c.Param("plID")

	result := doDeleteAudio(id)

	c.JSON(http.StatusOK, gin.H{
		"code": result,
	})
}

func doDeleteAudio(id string) int8 {
	audioInfo := utils.Client.HGetAll("au:" + id).Val()
	audioName := audioInfo["name"]
	audioArtist := audioInfo["artist"]
	plID := audioInfo["playlist"]

	plName := utils.Client.HGet("pl:"+plID, "name").Val()

	audioPath := conf.FileDIR + "/" + strings.Replace(plName, "/", "*", -1) + "/" + audioName + " - " + audioArtist
	err := os.RemoveAll(audioPath)
	if err != nil {
		return 0
	}

	utils.Client.LRem("pla:"+plID, 0, "au:"+id)
	utils.Client.Del("au:" + id)
	return 1
}

// 新增音频信息
// TODO update
func AddAudio(c *gin.Context) {
	audioInfo := make(map[string]interface{})
	err := c.BindJSON(&audioInfo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "decode error",
		})
		return
	}

	// generate song ID
	songID := utils.GetRandom()
	audioInfo["id"] = songID
	audioInfo["create"] = time.Now().Format("2006/1/2 15:04:05")

	utils.Client.LPush("pla:"+audioInfo["playlist"].(string), "au:"+songID)

	utils.Client.HMSet("au:"+songID, audioInfo)
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "add audio success",
		"id":   songID,
	})
}
