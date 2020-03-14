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

//获取音频列表
func GetAudioList(c *gin.Context) {
	audioList := utils.Client.LRange("audio-list", 0, -1)

	type info map[string]string
	infoList := make([]info, len(audioList.Val()))

	for idx, v := range audioList.Val() {
		res := utils.Client.HGetAll(v).Val()
		infoList[idx] = res
	}
	c.JSON(http.StatusOK, gin.H{
		"data": infoList,
	})
}

//删除音频信息
func DeleteAudio(c *gin.Context) {
	name := c.Param("name")
	utils.Client.LRem("audio-list", 0, name)
	utils.Client.Del(name)

	os.RemoveAll(conf.FileDIR + "/" + utils.HashName(name))

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "delete success",
	})
}

//新增音频信息
func AddAudio(c *gin.Context) {
	audioInfo := make(map[string]interface{})
	err := c.BindJSON(&audioInfo)
	if err != nil {
		//log.Fatalf("decode param error :%v", err)
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "decode error",
		})
	}

	// generate song ID
	// TODO update
	songID := utils.GetRandom()
	audioInfo["ID"] = songID
	//name := audioInfo["name"].(string)
	audioInfo["create"] = time.Now().Format("2006/1/2 15:04:05")

	//audioList := utils.Client.LRange("audio-list", 0, -1)
	//exist := utils.InList(audioList.Val(), name)

	utils.Client.LPush("audio-list", songID)

	utils.Client.HMSet(songID, audioInfo)
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "add audio success",
	})
}

//文件上传
func UploadFiles(c *gin.Context) {
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
		localFileDir := conf.FileDIR + "/" + utils.HashName(name)

		os.MkdirAll(localFileDir, os.ModePerm)

		localFile, err := os.Create(localFileDir + "/" + fileName)

		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "create" + localFileDir + " file error",
			})
			//log.Fatalf("create %s file error %v", localFileDir, err)
			//return
		}

		defer localFile.Close()
		io.Copy(localFile, file)

		utils.Client.HSet(name, v, localFileDir+"/"+fileName)

		c.JSON(http.StatusOK, gin.H{
			"code": 1,
			"msg":  "Upload success",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "File type error",
		})
	}
}

//文件下载
func DownloadFile(c *gin.Context) {
	name := c.Param("name")
	fileType := c.Param("type")[1:]

	info := utils.Client.HGetAll(name).Val()
	if filePath, ok := info[fileType]; ok {
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			//log.Fatalf("open file %s error : %v", filePath, err)
			//return
			c.JSON(http.StatusBadRequest, gin.H{
				"code": 0,
				"msg":  "open file " + filePath + " error",
			})
		}
		fileFrag := strings.Split(filePath, ".")
		fileSuffix := fileFrag[len(fileFrag)-1]
		c.Writer.WriteHeader(http.StatusOK)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", name+"."+fileSuffix))
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
