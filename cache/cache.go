package cache

import (
	"context"
	"os"
	"path/filepath"

	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/glb/util/osutil"
	"github.com/whoisnian/virt-launcher/global"
)

var appCacheDir = ""
var subCacheDir = []string{"images", "cloud-init", "boot"}

func Setup(ctx context.Context) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		global.LOG.Fatal(ctx, "os.UserCacheDir", logger.Error(err))
	}

	appCacheDir = filepath.Join(userCacheDir, global.AppName)
	global.LOG.Debugf(ctx, "use base cache dir %s", appCacheDir)

	for _, sub := range subCacheDir {
		err = os.MkdirAll(filepath.Join(appCacheDir, sub), osutil.DefaultDirMode)
		if err != nil {
			global.LOG.Fatal(ctx, "os.MkdirAll", logger.Error(err))
		}
	}
}

func JoinImagesDir(sub string) string {
	return filepath.Join(appCacheDir, "images", sub)
}
func JoinCloudInitDir(sub string) string {
	return filepath.Join(appCacheDir, "cloud-init", sub)
}
func JoinBootDir(sub string) string {
	return filepath.Join(appCacheDir, "boot", sub)
}
