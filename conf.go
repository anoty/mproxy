package mproxy

import (
	"os"
	"io/ioutil"
	"encoding/json"
	log "github.com/thinkboy/log4go"
)

type Conf struct {
	Servers              []string        `json:"servers"`
	MemcacheMaxIdleConns int             `json:"memcacheMaxIdleConns"`
	Port                 string          `json:"Port"`
}

func InitConf(conf *Conf) {
	wd, _ := os.Getwd()
	var (
		cp string
		lp string
	)

	cp = wd + "/conf/mpd.json"
	lp = wd + "/conf/log.xml"

	f, err := os.Open(cp)
	if err != nil {
		log.Error("conf file error", err)
		panic(err)
	}
	defer f.Close()

	fd, err := ioutil.ReadAll(f)
	if err != nil {
		log.Error("conf error", err)
		panic(err)
	}
	err = json.Unmarshal(fd, conf)
	if err != nil {
		log.Error("json error", err)
		panic(err)
	}
	log.LoadConfiguration(lp)
}
