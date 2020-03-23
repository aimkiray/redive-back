package api

import (
	"fmt"
	"github.com/aimkiray/reosu-server/conf"
	"github.com/aimkiray/reosu-server/utils"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

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
