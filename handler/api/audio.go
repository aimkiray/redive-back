package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"

	"github.com/aimkiray/reosu-server/conf"
	"github.com/aimkiray/reosu-server/utils"
)

var client *redis.Client

func init() {
	cfg := conf.Cfg
	sec, err := cfg.GetSection("database")
	if err != nil {
		log.Fatal("open section database error %v", err)
	}
	hostname := sec.Key("HOST").MustString("127.0.0.1:6379")
	password := sec.Key("PASSWORD").MustString("")
	db := sec.Key("DB").MustInt(0)

	client = redis.NewClient(&redis.Options{
		Addr:     hostname,
		Password: password,
		DB:       db,
	})
}

//获取音频列表
func GetAudioList(c *gin.Context) {
	audioList := client.LRange("audio-list", 0, -1)

	type info map[string]string
	infoList := make([]info, len(audioList.Val()))

	for idx, v := range audioList.Val() {
		res := client.HGetAll(v).Val()
		infoList[idx] = res
	}
	c.JSON(http.StatusOK, gin.H{
		"data": infoList,
	})
}

//删除音频信息
func DeleteAudio(c *gin.Context) {
	name := c.Param("name")
	client.LRem("audio-list", 0, name)
	client.Del(name)

	os.RemoveAll(conf.FilePath + "/" + utils.HashName(name))

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
	name := audioInfo["name"].(string)
	audioInfo["create"] = time.Now().Format("2006/1/2 15:04:05")

	audioList := client.LRange("audio-list", 0, -1)

	update := utils.InList(audioList.Val(), name)

	if !update {
		client.LPush("audio-list", name)
	}

	client.HMSet(name, audioInfo)
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
		localFilePath := conf.FilePath + "/" + utils.HashName(name)

		os.MkdirAll(localFilePath, os.ModePerm)

		localFile, err := os.Create(localFilePath + "/" + fileName)

		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "create" + localFilePath + " file error",
			})
			//log.Fatalf("create %s file error %v", localFilePath, err)
			//return
		}

		defer localFile.Close()
		io.Copy(localFile, file)

		client.HSet(name, v, localFilePath+"/"+fileName)

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

	info := client.HGetAll(name).Val()
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
