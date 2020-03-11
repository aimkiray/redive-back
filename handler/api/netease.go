package api

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aimkiray/reosu-server/conf"
)

const lyricURL = "http://music.163.com/api/song/media?id="
var second = flag.Int("timeout", 10, "timeout second")

type Lyric struct {
	SongStatus   int    `json:"songStatus"`
	LyricVersion int    `json:"lyricVersion"`
	Lyric        string `json:"lyric"`
	Code         int    `json:"code"`
	Msg          string `json:"msg"`
}

func GetLyric(c *gin.Context) {
	id := c.Query("id")
	localFilePath := conf.FilePath + "/lyric/test.lrc"
	timeout := time.Second * time.Duration(*second)
	transport := &http.Transport{}
	err := downloadFile(localFilePath, lyricURL + id, transport, timeout)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 0,
			"msg":  "get a bug" + err.Error(),
		})
	}
	content, err := ioutil.ReadFile(localFilePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 0,
			"msg":  "open file " + localFilePath + " error",
		})
	}
	c.Writer.WriteHeader(http.StatusOK)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=test.lrc"))
	c.Header("Content-Type", "application/text/plain")
	c.Header("Accept-Length", fmt.Sprintf("%d", len(content)))
	c.Writer.Write([]byte(content))
}

func downloadFile(fileName string, url string, transport *http.Transport, timeout time.Duration) error {
	client := &http.Client{Transport: transport, Timeout: timeout}

	newReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	newReq.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) "+
		"Chrome/74.0.3729.131 Safari/537.36")
	newReq.Header.Set("Host", "music.163.com")
	newReq.Header.Set("Referer", "http://music.163.com/search/")
	newReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var resp *http.Response
	for resp, err = client.Do(newReq); err != nil; {
		resp, err = client.Do(newReq)
	}
	defer Close(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return errors.New(url + " " + resp.Status)
	}

	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var lyc Lyric
	err = json.Unmarshal(res, &lyc)
	if err != nil {
		panic(err)
	}

	if lyc.Code != 200 {
		return errors.New(lyc.Msg)
	}

	output, err := os.Create(fileName)
	defer Close(output)

	if err != nil {
		return err
	}

	_, err = output.Write([]byte(lyc.Lyric))
	if err != nil {
		return err
	}

	return nil
}

func Close(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatalln(err)
	}
}