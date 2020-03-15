package handler

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/aimkiray/reosu-server/conf"
	"github.com/aimkiray/reosu-server/handler/api"
	"github.com/aimkiray/reosu-server/middleware"
)

func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.Use(cors.Default())
	gin.SetMode(conf.RunMode)

	r.GET("/api/login", api.Login)
	r.GET("/api/batch", api.BatchDownload)
	r.GET("/api/batch/status", api.BatchStatus)
	// NetEase API
	r.GET("/api/song", api.PlayList)
	r.GET("/api/song/detail", api.SongDetail)
	r.GET("/api/song/url", api.SongURL)
	r.GET("/api/lyric", api.SongLyric)

	handler := r.Group("/api")
	handler.Use(middleware.JWT())
	{
		handler.GET("/check", api.CheckToken)
		handler.GET("/audio", api.GetAudioList)
		handler.GET("/audio/download/:name/*type", api.DownloadFile)
		handler.POST("/audio/upload", api.UploadFiles)
		handler.DELETE("/audio/:name", api.DeleteAudio)
		handler.POST("/audio", api.AddAudio)
	}

	return r
}
