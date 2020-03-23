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

	// User API
	r.GET("/api/login", api.Login)

	// Batch Import API
	r.GET("/api/batch/status", api.BatchStatus)

	// NetEase API
	r.GET("/api/song", api.PlayList)
	r.GET("/api/song/detail", api.SongDetail)
	r.GET("/api/song/url", api.SongURL)
	r.GET("/api/lyric", api.SongLyric)

	// Audio API
	r.GET("/api/playlist", api.GetAllPlayList)
	r.GET("/api/audio", api.GetAllAudio)
	r.GET("/api/audio/download/:id/*type", api.DownloadFile)

	// Require permissions
	handler := r.Group("/api")
	handler.Use(middleware.JWT())
	{
		// Check token
		handler.GET("/check", api.CheckToken)

		// Playlist API
		handler.POST("/playlist", api.AddPlaylist)
		handler.DELETE("/playlist/:id", api.DeletePlaylist)

		// Audio API
		handler.POST("/audio/upload", api.UploadFiles)
		handler.DELETE("/audio/:id", api.DeleteAudio)
		handler.POST("/audio", api.AddAudio)

		// Batch Import API
		r.GET("/api/batch", api.BatchDownload)
	}

	return r
}
