package main

import (
	"fmt"
	"runtime"

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
		fmt.Printf("%s %s(%s)\n", global.AppName, global.Version, global.BuildTime)
		return
	}

	image.Init()
	if global.CFG.List {
		image.ListAll()
		return
	}

	distro := image.LookupDistro(global.CFG.Distro)
	if distro == nil {
		logger.Error("Unknown distro ", global.CFG.Distro)
		return
	}

	arch := global.CFG.Arch
	if arch == "" {
		arch = runtime.GOARCH
		logger.Warn("Automatically detect GOARCH: ", arch)
	}
	img := distro.LookupByArch(arch)
	if img == nil {
		logger.Error("Unknown arch ", arch, " for distro ", global.CFG.Distro)
		return
	}

	if err := img.Download(); err != nil {
		logger.Error(err)
	}
}
