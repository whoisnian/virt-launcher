package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/whoisnian/glb/util/osutil"
	"github.com/whoisnian/virt-launcher/cache"
	"github.com/whoisnian/virt-launcher/global"
	"github.com/whoisnian/virt-launcher/image"
	"github.com/whoisnian/virt-launcher/third"
)

func main() {
	global.SetupConfig()
	global.SetupLogger()
	global.LOG.Debugf("Use config: %+v", global.CFG)

	if global.CFG.Version {
		fmt.Printf("%s %s(%s)\n", global.AppName, global.Version, global.BuildTime)
		return
	}

	cache.Setup()
	image.Setup()
	third.Setup()
	if global.CFG.ListAll {
		image.ListAll()
		return
	}

	if global.CFG.Arch == "" {
		global.LOG.Warnf("Automatically detect architecture: %s", runtime.GOARCH)
		global.CFG.Arch = runtime.GOARCH
	}
	img, err := image.LookupImage(global.CFG.Os, global.CFG.Arch)
	if err != nil {
		global.LOG.Errorf("LookupImage(%s,%s): %v", global.CFG.Os, global.CFG.Arch, err)
		return
	}

	oriImagePath := cache.Images(img.BaseName())
	global.LOG.Infof("Start downloading image to %s", oriImagePath)
	if err = img.Download(oriImagePath); err != nil {
		global.LOG.Error(err.Error())
		return
	}
	timeStr := strconv.FormatInt(time.Now().UnixMilli(), 36)
	finalImageName := fmt.Sprintf("%s.%s.qcow2", global.CFG.Name, timeStr)
	finalImagePath := cache.Boot(finalImageName)
	global.LOG.Debugf("Start copying file from %s to %s", oriImagePath, finalImagePath)
	if _, err = osutil.CopyFile(oriImagePath, finalImagePath); err != nil {
		global.LOG.Error(err.Error())
		return
	}

	output, err := third.ResizeImage(finalImagePath)
	if err != nil {
		global.LOG.Error(string(output))
		return
	}

	cloudIsoCacheDir := cache.CloudInit(fmt.Sprintf("%s.%s", global.CFG.Name, timeStr))
	cloudIsoName := fmt.Sprintf("%s.%s.iso", global.CFG.Name, timeStr)
	cloudIsoPath := cache.Boot(cloudIsoName)
	output, err = third.CreateCloudInitIso(cloudIsoCacheDir, cloudIsoPath, timeStr)
	if err != nil {
		global.LOG.Error(string(output))
		return
	}

	disk := filepath.Join(global.CFG.Storage, finalImageName)
	cdrom := filepath.Join(global.CFG.Storage, cloudIsoName)
	for _, params := range [][]string{
		{finalImagePath, disk},
		{cloudIsoPath, cdrom},
	} {
		if global.CFG.DryRun {
			global.LOG.Infof("[DRY-RUN] %s", exec.Command("mv", params[0], params[1]).String())
		} else {
			global.LOG.Debug(exec.Command("mv", params[0], params[1]).String())
			if err = osutil.MoveFile(params[0], params[1]); err != nil {
				global.LOG.Error(err.Error())
				return
			}
		}
	}

	output, err = third.CreateVM(disk, cdrom)
	if err != nil {
		global.LOG.Error(string(output))
		return
	}

	output, err = third.WaitForVMOff()
	if err != nil {
		global.LOG.Error(string(output))
		return
	}

	output, err = third.DetachCloudInitIso(cdrom)
	if err != nil {
		global.LOG.Error(string(output))
		return
	}
	defer os.Remove(cdrom)

	output, err = third.StartVM()
	if err != nil {
		global.LOG.Error(string(output))
		return
	}

	global.LOG.Infof("[NOTE] cloud image default user: %s", img.Account)
	global.LOG.Infof("[NOTE] fetch vm ip addr: %s", exec.Command("virsh", "--connect", "qemu:///system", "domifaddr", "--domain", global.CFG.Name).String())
}
