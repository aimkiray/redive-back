package api

import (
	"github.com/aimkiray/reosu-server/conf"
	"github.com/aimkiray/reosu-server/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strings"
	"time"
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

// 增加或修改playlist
func AddPlaylist(c *gin.Context) {
	id := c.Query("id")
	name := c.Query("name")

	//playlist := utils.Client.LRange("playlist", 0, -1)
	//exist := utils.InList(playlist.Val(), "pl:"+id)

	if id != "" {
		utils.Client.HSet("pl:"+id, "name", name)
	} else {
		id = utils.GetRandom()
		// 先记录播放列表的key，便于索引
		utils.Client.LPush("playlist", "pl:"+id)

		utils.Client.HSet("pl:"+id, "id", id)
		utils.Client.HSet("pl:"+id, "name", name)
		utils.Client.HSet("pl:"+id, "create", time.Now().Format("2006/1/2 15:04:05"))
		// 替换分隔符
		baseFileDir := conf.FileDIR + "/music/" + strings.Replace(name, "/", "*", -1)
		os.MkdirAll(baseFileDir, os.ModePerm)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
	})
}

func DeletePlaylist(c *gin.Context) {
	id := c.Param("id")

	audioList := utils.Client.LRange("pla:"+id, 0, -1).Val()

	for _, value := range audioList {
		// del audio key
		utils.Client.Del(value)
	}
	utils.Client.LRem("", 0, "pla:"+id)
	plName := utils.Client.HGet("pl:"+id, "name").Val()

	os.RemoveAll(conf.FileDIR + "/" + strings.Replace(plName, "/", "*", -1))

	utils.Client.Del("pl:" + id)
	utils.Client.LRem("playlist", 0, "pl:"+id)

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
	})
}
