package api

import (
	"github.com/aimkiray/reosu-server/conf"
	"github.com/aimkiray/reosu-server/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"time"
)

//获取歌单列表
func GetAllPlayList(c *gin.Context) {
	playlist := utils.Client.LRange("playlist", 0, -1)

	if playlist.Val() == nil || len(playlist.Val()) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "no playlist",
		})
		return
	}

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
// 是否需要禁止重名
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
		baseFileDir := conf.FileDIR + "/music/" + id
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

	filePath := conf.FileDIR + "/music/" + id

	os.RemoveAll(filePath)

	utils.Client.Del("pl:" + id)
	utils.Client.LRem("playlist", 0, "pl:"+id)

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
	})
}
