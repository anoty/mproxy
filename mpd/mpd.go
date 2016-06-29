package main

import (
	log "github.com/thinkboy/log4go"
	"runtime"
	"mproxy"
	"mproxy/logic/proxy"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var conf mproxy.Conf
	mproxy.InitConf(&conf)

	log.Info("mpd start")

	proxy.Run(conf)
}
