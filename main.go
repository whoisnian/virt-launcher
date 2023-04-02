package main

import (
	"fmt"

	"github.com/whoisnian/glb/config"
	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/virt-launcher/global"
	"github.com/whoisnian/virt-launcher/image"
)

func main() {
	err := config.FromCommandLine(&global.CFG)
	if err != nil {
		logger.Fatal(err)
	}
	logger.SetDebug(global.CFG.Debug)

	if global.CFG.Version {
		fmt.Printf("virt-launcher %s(%s)\n", global.Version, global.BuildTime)
		return
	}

	image.Init()
	if global.CFG.List {
		image.ListAll()
	}
}
