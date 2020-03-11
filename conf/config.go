package conf

import (
	"log"
	"time"

	"github.com/go-ini/ini"
)

var (
	Cfg *ini.File

	RunMode string

	FilePath  string
	FileTypes map[string]string

	UserName string
	PassWord string

	HTTPPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	JwtSecret    string
)

func init() {
	var err error
	Cfg, err = ini.Load("./conf/config.ini")
	if err != nil {
		log.Fatalf("Fail to parse 'config.ini': %v", err)
	}
	loadBase()
	loadServer()
	loadApp()
}

func loadBase() {
	RunMode = Cfg.Section("").Key("RUN_MODE").MustString("debug")
}

func loadServer() {
	sec, err := Cfg.GetSection("server")
	if err != nil {
		log.Fatalf("Fail to get section 'server': %v", err)
	}

	HTTPPort = sec.Key("HTTP_PORT").MustInt(2333)
	ReadTimeout = time.Duration(sec.Key("READ_TIMEOUT").MustInt(60)) * time.Second
	WriteTimeout = time.Duration(sec.Key("WRITE_TIMEOUT").MustInt(60)) * time.Second
}

func loadApp() {
	JwtSecret = Cfg.Section("").Key("JWT_SECRET").MustString("akari")
	FilePath = Cfg.Section("file").Key("PATH").MustString("static/")
	FileTypes = map[string]string{
		"mp3":  "audio",
		"lrc":  "lrc",
		"jpg":  "cover",
		"jpeg": "cover",
		"png":  "cover",
		"webp": "cover",
	}

	UserName = Cfg.Section("admin").Key("USERNAME").MustString("admin")
	PassWord = Cfg.Section("admin").Key("PASSWORD").MustString("admin")
}
