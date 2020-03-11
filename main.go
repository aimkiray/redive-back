package main

import (
	"fmt"
	"log"
	"syscall"

	"github.com/fvbock/endless"

	"github.com/aimkiray/reosu-server/conf"
	"github.com/aimkiray/reosu-server/handler"
)

func main() {
	endless.DefaultReadTimeOut = conf.ReadTimeout
	endless.DefaultWriteTimeOut = conf.WriteTimeout
	endless.DefaultMaxHeaderBytes = 1 << 20
	endPoint := fmt.Sprintf(":%d", conf.HTTPPort)

	server := endless.NewServer(endPoint, handler.InitRouter())
	server.BeforeBegin = func(add string) {
		log.Printf("Actual pid is %d", syscall.Getpid())
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("Server err: %v", err)
	}
}
