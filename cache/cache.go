package cache

import (
	"context"
	"os"
	"path/filepath"

	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/glb/util/osutil"
	"github.com/whoisnian/virt-launcher/global"
)

var (
	appCacheDir string

	Index     = Dir("index")
	Images    = Dir("images")
	CloudInit = Dir("cloud-init")
	Boot      = Dir("boot")
)

type Dir string

func (dir Dir) FullPath() string {
	return filepath.Join(appCacheDir, string(dir))
}

func (dir Dir) Join(elem ...string) string {
	return filepath.Join(appCacheDir, string(dir), filepath.Join(elem...))
}

func (dir Dir) Reset() error {
	if err := os.RemoveAll(dir.FullPath()); err != nil {
		return err
	}
	return os.MkdirAll(dir.FullPath(), osutil.DefaultDirMode)
}

func Setup(ctx context.Context) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		global.LOG.Fatal(ctx, "os.UserCacheDir", logger.Error(err))
	}

	appCacheDir = filepath.Join(userCacheDir, global.AppName)
	global.LOG.Debugf(ctx, "use app cache dir %s", appCacheDir)

	for _, dir := range []Dir{Index, Images, CloudInit, Boot} {
		if err = os.MkdirAll(dir.FullPath(), osutil.DefaultDirMode); err != nil {
			global.LOG.Fatal(ctx, "os.MkdirAll", logger.Error(err))
		}
	}
}
