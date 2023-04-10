package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
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
	timeStr := strconv.FormatInt(time.Now().UnixMilli(), 36)
	finalImageName := fmt.Sprintf("%s.%s.qcow2", global.CFG.Name, timeStr)
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

	cloudIsoCacheDir := cache.CloudInit(fmt.Sprintf("%s.%s", global.CFG.Name, timeStr))
	cloudIsoName := fmt.Sprintf("%s.%s.iso", global.CFG.Name, timeStr)
	cloudIsoPath := cache.Boot(cloudIsoName)
	output, err = third.CreateCloudInitIso(cloudIsoCacheDir, cloudIsoPath, timeStr)
	if err != nil {
		logger.Error(string(output))
		return
	}

	disk := filepath.Join(global.CFG.Storage, finalImageName)
	cdrom := filepath.Join(global.CFG.Storage, cloudIsoName)
	for _, params := range [][]string{
		{finalImagePath, disk},
		{cloudIsoPath, cdrom},
	} {
		if global.CFG.DryRun {
			logger.Info("[DRY-RUN] ", exec.Command("mv", params[0], params[1]).String())
		} else {
			logger.Debug(exec.Command("mv", params[0], params[1]).String())
			if err = os.Rename(params[0], params[1]); err != nil {
				logger.Error(err)
				return
			}
		}
	}

	output, err = third.CreateVM(disk, cdrom)
	if err != nil {
		logger.Error(string(output))
		return
	}

	logger.Info("[POST-INSTALL] ", exec.Command("virsh", "--connect", "qemu:///system", "detach-disk", "--persistent", "--domain", global.CFG.Name, cdrom).String())
	logger.Info("[POST-INSTALL] ", exec.Command("virsh", "--connect", "qemu:///system", "reboot", "--domain", global.CFG.Name).String())
	logger.Info("[POST-INSTALL] ", exec.Command("virsh", "--connect", "qemu:///system", "domifaddr", " --domain", global.CFG.Name).String())
	logger.Info("[POST-INSTALL] ", exec.Command("rm", cdrom).String())
}
