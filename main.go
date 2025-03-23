package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/glb/util/osutil"
	"github.com/whoisnian/virt-launcher/cache"
	"github.com/whoisnian/virt-launcher/global"
	"github.com/whoisnian/virt-launcher/image"
	"github.com/whoisnian/virt-launcher/third"
)

func main() {
	ctx := context.Background()
	global.SetupConfig(ctx)
	global.SetupLogger(ctx)
	global.LOG.Debugf(ctx, "use config: %+v", global.CFG)

	if global.CFG.Version {
		fmt.Printf("%s version %s built with %s at %s\n", global.AppName, global.Version, runtime.Version(), global.BuildTime)
		return
	}

	cache.Setup(ctx)
	image.Setup(ctx)
	third.Setup(ctx)
	if global.CFG.ListAll {
		image.ListAll()
		return
	}

	if global.CFG.Arch == "" {
		global.LOG.Warnf(ctx, "automatically detect architecture: %s", runtime.GOARCH)
		global.CFG.Arch = runtime.GOARCH
	}
	img, err := image.LookupImage(global.CFG.Os, global.CFG.Arch)
	if err != nil {
		global.LOG.Errorf(ctx, "image.LookupImage(%s,%s): %v", global.CFG.Os, global.CFG.Arch, err)
		return
	}

	oriImagePath := cache.JoinImagesDir(img.BaseName())
	global.LOG.Infof(ctx, "start downloading image to %s", oriImagePath)
	if err = img.Download(ctx, oriImagePath); err != nil {
		global.LOG.Error(ctx, "image.Download", logger.Error(err))
		return
	}
	timeStr := strconv.FormatInt(time.Now().UnixMilli(), 36)
	finalImageName := fmt.Sprintf("%s.%s.qcow2", global.CFG.Name, timeStr)
	finalImagePath := cache.JoinBootDir(finalImageName)
	global.LOG.Debugf(ctx, "start copying file from %s to %s", oriImagePath, finalImagePath)
	if _, err = osutil.CopyFile(oriImagePath, finalImagePath); err != nil {
		global.LOG.Error(ctx, "osutil.CopyFile", logger.Error(err))
		return
	}

	output, err := third.ResizeImage(ctx, finalImagePath)
	if err != nil {
		global.LOG.Error(ctx, "third.ResizeImage "+string(output), logger.Error(err))
		return
	}

	cloudIsoCacheDir := cache.JoinCloudInitDir(fmt.Sprintf("%s.%s", global.CFG.Name, timeStr))
	cloudIsoName := fmt.Sprintf("%s.%s.iso", global.CFG.Name, timeStr)
	cloudIsoPath := cache.JoinBootDir(cloudIsoName)
	output, err = third.CreateCloudInitIso(ctx, cloudIsoCacheDir, cloudIsoPath, timeStr)
	if err != nil {
		global.LOG.Error(ctx, "third.CreateCloudInitIso "+string(output), logger.Error(err))
		return
	}

	disk := filepath.Join(global.CFG.Storage, finalImageName)
	cdrom := filepath.Join(global.CFG.Storage, cloudIsoName)
	for _, params := range [][]string{
		{finalImagePath, disk},
		{cloudIsoPath, cdrom},
	} {
		if global.CFG.DryRun {
			global.LOG.Infof(ctx, "[DRY-RUN] %s", exec.Command("mv", params[0], params[1]).String())
		} else {
			global.LOG.Debug(ctx, exec.Command("mv", params[0], params[1]).String())
			if err = osutil.MoveFile(params[0], params[1]); err != nil {
				global.LOG.Error(ctx, "osutil.MoveFile", logger.Error(err))
				return
			}
		}
	}

	output, err = third.CreateVM(ctx, disk, cdrom)
	if err != nil {
		global.LOG.Error(ctx, "third.CreateVM "+string(output), logger.Error(err))
		return
	}

	output, err = third.WaitForVMOff(ctx)
	if err != nil {
		global.LOG.Error(ctx, "third.WaitForVMOff "+string(output), logger.Error(err))
		return
	}

	output, err = third.DetachCloudInitIso(ctx, cdrom)
	if err != nil {
		global.LOG.Error(ctx, "third.DetachCloudInitIso "+string(output), logger.Error(err))
		return
	}
	defer os.Remove(cdrom)

	output, err = third.StartVM(ctx)
	if err != nil {
		global.LOG.Error(ctx, "third.StartVM "+string(output), logger.Error(err))
		return
	}

	global.LOG.Infof(ctx, "[NOTE] cloud image default user: %s", img.Account)
	global.LOG.Infof(ctx, "[NOTE] fetch vm ip addr: %s", exec.Command("virsh", "--connect", global.CFG.Connect, "domifaddr", "--domain", global.CFG.Name).String())
}
