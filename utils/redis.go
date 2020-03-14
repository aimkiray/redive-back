package utils

import (
	"github.com/aimkiray/reosu-server/conf"
	"github.com/go-redis/redis"
	"log"
)

var Client *redis.Client

func init() {
	cfg := conf.Cfg
	sec, err := cfg.GetSection("database")
	if err != nil {
		log.Fatal("Open section database error, " + err.Error())
	}
	hostname := sec.Key("HOST").MustString("127.0.0.1:6379")
	password := sec.Key("PASSWORD").MustString("")
	db := sec.Key("DB").MustInt(0)

	Client = redis.NewClient(&redis.Options{
		Addr:     hostname,
		Password: password,
		DB:       db,
	})
}
