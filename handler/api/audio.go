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
	// 默认查询最新加入的歌单
	if plID == "" {
		plList := utils.Client.LRange("playlist", 0, 0).Val()
		if plList == nil || len(plList) != 1 {
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "no audio",
			})
			return
		}
		plID = strings.Split(plList[0], ":")[1]
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

	audioPath := conf.FileDIR + "/music/" + plID + "/" + audioArtist + " - " + audioName
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

	if audioInfo["playlist"] == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "未找到磁带",
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

func UpdateData(c *gin.Context) {
	info := make(map[string]interface{})
	err := c.BindJSON(&info)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "decode error",
		})
		return
	}

	if info["regions"] != nil {
		utils.Client.HSet("au:"+info["id"].(string), "regions", info["regions"].(string))
	}

	if info["peaks"] != nil {
		utils.Client.HSet("au:"+info["id"].(string), "peaks", info["peaks"].(string))
	}

	if info["duration"] != nil {
		utils.Client.HSet("au:"+info["id"].(string), "duration", info["duration"].(string))
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "update data success",
	})
}
