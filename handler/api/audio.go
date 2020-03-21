package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aimkiray/reosu-server/conf"
	"github.com/aimkiray/reosu-server/utils"
)

//获取歌单列表
func GetAllPlayList(c *gin.Context) {
	playlist := utils.Client.LRange("playlist", 0, -1)

	infoList := make([]map[string]string, len(playlist.Val()))

	for index, value := range playlist.Val() {
		res := utils.Client.HGetAll(value).Val()
		infoList[index] = res
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": infoList,
	})
}

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
	//name := audioInfo["name"].(string)
	audioInfo["create"] = time.Now().Format("2006/1/2 15:04:05")

	//audioList := utils.Client.LRange("audio-list", 0, -1)
	//exist := utils.InList(audioList.Val(), name)

	utils.Client.LPush("pla:"+audioInfo["playlist"].(string), "au:"+songID)

	utils.Client.HMSet("au:"+songID, audioInfo)
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "add audio success",
		"id":   songID,
	})
}

// 文件上传
func UploadFiles(c *gin.Context) {
	id := c.PostForm("id")
	name := c.PostForm("name")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 0,
			"msg":  "file upload failed",
		})
		return
	}

	fileName := header.Filename
	fileFrag := strings.Split(fileName, ".")
	fileSuffix := fileFrag[len(fileFrag)-1]

	if v, ok := conf.FileTypes[fileSuffix]; ok {
		localFileDir := conf.FileDIR + "/" + name

		os.MkdirAll(localFileDir, os.ModePerm)

		localFile, err := os.Create(localFileDir + "/" + fileName)

		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "create" + localFileDir + " file error. " + err.Error(),
			})
			return
		}

		defer localFile.Close()
		io.Copy(localFile, file)

		utils.Client.HSet("au:"+id, v, localFileDir+"/"+fileName)

		c.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "Upload success",
		})
	} else {
		doDeleteAudio(id)
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "File type error",
		})
	}
}

//文件下载
func DownloadFile(c *gin.Context) {
	id := c.Param("id")
	fileType := c.Param("type")[1:]

	info := utils.Client.HGetAll("au:" + id).Val()
	if filePath, ok := info[fileType]; ok {
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 0,
				"msg":  "open file " + filePath + " error. " + err.Error(),
			})
			return
		}

		// 获取文件名
		fileFrag := strings.Split(filePath, "/")
		//fileSuffix := fileFrag[len(fileFrag)-1]
		filename := fileFrag[len(fileFrag)-1]

		c.Writer.WriteHeader(http.StatusOK)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Type", "application/text/plain")
		c.Header("Accept-Length", fmt.Sprintf("%d", len(content)))
		c.Writer.Write([]byte(content))
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 0,
			"msg":  "no such file",
		})
	}
}
