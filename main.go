package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

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
	if global.CFG.Update {
		if err := image.UpdateIndex(ctx); err != nil {
			global.LOG.Fatalf(ctx, "image.UpdateIndex error: %v", err)
		}
		return
	} else if global.CFG.ListAll {
		image.ListAll(os.Stdout)
		return
	}

	if global.CFG.Arch == "" {
		global.LOG.Warnf(ctx, "automatically detect architecture: %s", runtime.GOARCH)
		global.CFG.Arch = runtime.GOARCH
	}
	img, err := image.LookupImage(global.CFG.Os, global.CFG.Arch)
	if err != nil {
		global.LOG.Fatalf(ctx, "image.LookupImage(%s,%s) error: %v", global.CFG.Os, global.CFG.Arch, err)
	}

	rawImageFilePath := cache.Images.Join(img.BaseName())
	global.LOG.Infof(ctx, "start downloading image to %s", rawImageFilePath)
	if err = img.Download(ctx, rawImageFilePath); err != nil {
		global.LOG.Fatalf(ctx, "image.Download error: %v", err)
	}

	var output []byte
	if output, err = third.EnsureNetworkIsActive(ctx); err != nil {
		global.LOG.Fatalf(ctx, "third.EnsureNetworkIsActive error: %v %s", err, output)
	}
	if output, err = third.EnsureStoragePoolIsActive(ctx); err != nil {
		global.LOG.Fatalf(ctx, "third.EnsureStoragePoolIsActive error: %v %s", err, output)
	}

	timeStr := strconv.FormatInt(time.Now().UnixMilli(), 36)
	diskVolume := fmt.Sprintf("%s.%s.qcow2", global.CFG.Name, timeStr)
	if output, err = third.UploadVolume(ctx, diskVolume, rawImageFilePath); err != nil {
		global.LOG.Fatalf(ctx, "third.UploadVolume error: %v %s", err, output)
	}
	if output, err = third.ResizeVolume(ctx, diskVolume); err != nil {
		global.LOG.Fatalf(ctx, "third.ResizeVolume error: %v %s", err, output)
	}

	seedCacheDir := cache.CloudInit.Join(fmt.Sprintf("%s.%s", global.CFG.Name, timeStr))
	seedVolume := fmt.Sprintf("%s.%s.seed.iso", global.CFG.Name, timeStr)
	seedFilePath := cache.Boot.Join(seedVolume)
	if output, err = third.CreateCloudInitIso(ctx, seedCacheDir, seedFilePath, timeStr); err != nil {
		global.LOG.Fatalf(ctx, "third.CreateCloudInitIso error: %v %s", err, output)
	}
	if output, err = third.UploadVolume(ctx, seedVolume, seedFilePath); err != nil {
		global.LOG.Fatalf(ctx, "third.UploadVolume error: %v %s", err, output)
	}

	if output, err = third.CreateVM(ctx, diskVolume, seedVolume); err != nil {
		global.LOG.Fatalf(ctx, "third.CreateVM error: %v %s", err, output)
	}

	if output, err = third.WaitForVMOff(ctx); err != nil {
		global.LOG.Errorf(ctx, "third.WaitForVMOff error: %v %s", err, output)
	}

	if output, err = third.DeleteVolume(ctx, seedVolume); err != nil {
		global.LOG.Fatalf(ctx, "third.DeleteVolume error: %v %s", err, output)
	}

	if output, err = third.StartVM(ctx); err != nil {
		global.LOG.Fatalf(ctx, "third.StartVM error: %v %s", err, output)
	}

	global.LOG.Infof(ctx, "[NOTE] cloud image default user: %s", img.Account)
	global.LOG.Infof(ctx, "[NOTE] fetch vm ip addr: %s", exec.Command("virsh", "--connect", global.CFG.Connect, "domifaddr", "--domain", global.CFG.Name).String())
}
