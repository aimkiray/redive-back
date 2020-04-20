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

	publicHandler := r.Group("/api")

	// User API
	publicHandler.GET("/login", api.Login)

	// Batch Import API
	publicHandler.GET("/batch/status", api.BatchStatus)

	// NetEase API
	publicHandler.GET("/song", api.PlayList)
	publicHandler.GET("/song/detail", api.SongDetail)
	publicHandler.GET("/song/url", api.SongURL)
	publicHandler.GET("/lyric", api.SongLyric)

	// Audio API
	publicHandler.GET("/playlist", api.GetAllPlayList)
	publicHandler.GET("/audio", api.GetAllAudio)
	publicHandler.GET("/audio/download/:id/*type", api.DownloadFile)
	publicHandler.GET("/audio/region", api.GetRegion)

	// Require permissions
	privateHandler := r.Group("/api")
	privateHandler.Use(middleware.JWT())
	{
		// Check token
		privateHandler.GET("/check", api.CheckToken)

		// Playlist API
		privateHandler.POST("/playlist", api.AddPlaylist)
		privateHandler.DELETE("/playlist/:id", api.DeletePlaylist)

		// Audio API
		privateHandler.POST("/audio/upload", api.UploadFiles)
		privateHandler.DELETE("/audio/:id", api.DeleteAudio)
		privateHandler.POST("/audio", api.AddAudio)
		privateHandler.PUT("/audio/region", api.UpdateRegion)

		// Batch Import API
		privateHandler.GET("/batch", api.BatchDownload)
	}

	return r
}
