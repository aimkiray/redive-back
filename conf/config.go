package conf

import (
	"log"
	"os"
	"time"

	"github.com/go-ini/ini"
)

var (
	Cfg *ini.File

	RunMode string

	FileDIR   string
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
	RunMode = os.Getenv("RUN_MODE")
	if RunMode == "" {
		RunMode = Cfg.Section("").Key("RUN_MODE").MustString("debug")
	}
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
	JwtSecret = os.Getenv("JWT_SECRET")
	if JwtSecret == "" {
		JwtSecret = Cfg.Section("").Key("JWT_SECRET").MustString("akari")
	}
	FileDIR = Cfg.Section("file").Key("DIR").MustString("static/")
	FileTypes = map[string]string{
		"mp3":  "audio",
		"lrc":  "lrc",
		"jpg":  "cover",
		"jpeg": "cover",
		"png":  "cover",
		"webp": "cover",
	}

	UserName = os.Getenv("GIN_USERNAME")
	if UserName == "" {
		UserName = Cfg.Section("user").Key("USERNAME").MustString("admin")
	}
	PassWord = os.Getenv("GIN_PASSWORD")
	if PassWord == "" {
		PassWord = Cfg.Section("user").Key("PASSWORD").MustString("admin")
	}
}
