package main

import (
	"fmt"
	"runtime"

	"github.com/whoisnian/glb/config"
	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/virt-launcher/global"
	"github.com/whoisnian/virt-launcher/image"
	"github.com/whoisnian/virt-launcher/third"
)

func setupPackages() {
	image.SetupIndex()
	image.SetupCache()
	third.SetupThird()
}

func main() {
	err := config.FromCommandLine(&global.CFG)
	if err != nil {
		logger.Fatal(err)
	}
	logger.SetDebug(global.CFG.Debug)

	setupPackages()

	if global.CFG.Version {
		fmt.Printf("%s %s(%s)\n", global.AppName, global.Version, global.BuildTime)
		return
	}
	if global.CFG.ListAll {
		image.ListAll()
		return
	}

	if global.CFG.Arch == "" {
		logger.Warn("Automatically detect architecture: ", runtime.GOARCH)
		global.CFG.Arch = runtime.GOARCH
	}
	img, err := image.LookupImage(global.CFG.Os, global.CFG.Arch)
	if err != nil {
		logger.Error("LookupImage(", global.CFG.Os, ",", global.CFG.Arch, "): ", err)
		return
	}

	logger.Info("Start downloading image to ", img.CacheFilePath())
	if err := img.Download(); err != nil {
		logger.Error(err)
		return
	}

	third.ResizeImage(img.CacheFilePath(), "20G")
	third.CreateVM("testing", global.CFG.Os, img.CacheFilePath())
}
