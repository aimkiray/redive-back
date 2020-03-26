package api

import (
	"fmt"
	"github.com/aimkiray/reosu-server/conf"
	"github.com/aimkiray/reosu-server/utils"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"strings"
)

// 文件上传
func UploadFiles(c *gin.Context) {
	id := c.PostForm("id")
	name := c.PostForm("name")
	artist := c.PostForm("artist")
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 0,
			"msg":  "file upload failed",
		})
		return
	}

	uploadName := header.Filename
	fileFrag := strings.Split(uploadName, ".")
	fileSuffix := fileFrag[len(fileFrag)-1]

	// 文件名，默认 artist - name
	filename := ""
	if artist == "" {
		filename = name
	} else {
		filename = artist + " - " + name
	}

	if v, ok := conf.FileTypes[fileSuffix]; ok {
		plID := utils.Client.HGet("au:"+id, "playlist").Val()
		localFileDir := conf.FileDIR + "/music/" + plID + "/" + filename + "/"

		os.MkdirAll(localFileDir, os.ModePerm)

		// 歌词翻译
		if v == "lrc" && strings.Contains(uploadName, "trans") {
			filename = filename + "_trans"
			v = "tlrc"
		}

		localFile, err := os.Create(localFileDir + filename + "." + fileSuffix)

		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "create" + localFileDir + " file error. " + err.Error(),
			})
			return
		}

		defer localFile.Close()
		io.Copy(localFile, file)

		utils.Client.HSet("au:"+id, v, localFileDir+filename+"."+fileSuffix)

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

// 文件下载
func DownloadFile(c *gin.Context) {
	id := c.Param("id")
	fileType := c.Param("type")[1:]

	info := utils.Client.HGetAll("au:" + id).Val()
	if filePath, ok := info[fileType]; ok {
		// 生成文件名
		fileFrag := strings.Split(filePath, "/")
		filename := fileFrag[len(fileFrag)-1]

		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Expires", "0")
		c.Header("Cache-Control", "must-revalidate")
		c.Header("Pragma", "public")
		c.File(filePath)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 0,
			"msg":  "no such file",
		})
	}
}
