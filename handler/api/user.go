package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aimkiray/reosu-server/conf"
	"github.com/aimkiray/reosu-server/utils"
)

func Login(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")
	if username == conf.UserName && password == conf.PassWord {
		token, err := utils.GenerateToken(username, password)
		if err != nil {
			//log.Fatalf("GenerateToken Error :%v", err)
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "Generate Token Error",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":  1,
			"msg":   "login success",
			"token": token,
		})

	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "username or password wrong",
		})
	}
}

func CheckToken(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "token is valid",
	})
}
