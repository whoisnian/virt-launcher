package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/virt-launcher/cache"
	"github.com/whoisnian/virt-launcher/global"
	"github.com/whoisnian/virt-launcher/image"
	"github.com/whoisnian/virt-launcher/third"
)

func main() {
	global.SetupConfig()

	cache.Setup()
	image.Setup()
	third.Setup()

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

	oriImagePath := cache.Images(img.BaseName())
	logger.Info("Start downloading image to ", oriImagePath)
	if err = img.Download(oriImagePath); err != nil {
		logger.Error(err)
		return
	}
	finalImageName := fmt.Sprintf("%s.%x.qcow2", global.CFG.Name, time.Now().UnixMilli())
	finalImagePath := cache.Boot(finalImageName)
	if err = cache.CopyFile(oriImagePath, finalImagePath); err != nil {
		logger.Error(err)
		return
	}

	output, err := third.ResizeImage(finalImagePath)
	if err != nil {
		logger.Error(string(output))
		return
	}

	cloudIsoCacheDir := cache.CloudInit(fmt.Sprintf("%s.%x", global.CFG.Name, time.Now().UnixMilli()))
	cloudIsoName := fmt.Sprintf("%s.%x.iso", global.CFG.Name, time.Now().UnixMilli())
	cloudIsoPath := cache.Boot(cloudIsoName)
	output, err = third.CreateCloudInitIso(cloudIsoCacheDir, cloudIsoPath)
	if err != nil {
		logger.Error(string(output))
		return
	}

	disk := filepath.Join(global.CFG.Storage, finalImageName)
	cdrom := filepath.Join(global.CFG.Storage, cloudIsoName)
	if global.CFG.DryRun {
		logger.Info("[DRY-RUN] ", exec.Command("mv", finalImagePath, disk).String())
		logger.Info("[DRY-RUN] ", exec.Command("mv", cloudIsoPath, cdrom).String())
	} else {
		os.Rename(finalImagePath, disk)
		os.Rename(cloudIsoPath, cdrom)
	}

	output, err = third.CreateVM(disk, cdrom)
	if err != nil {
		logger.Error(string(output))
		return
	}
}
