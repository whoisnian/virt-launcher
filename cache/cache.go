package cache

import (
	"os"
	"path/filepath"

	"github.com/whoisnian/glb/util/osutil"
	"github.com/whoisnian/virt-launcher/global"
)

var appCacheDir = ""
var subCacheDir = []string{"images", "cloud-init", "boot"}

func Setup() {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		global.LOG.Fatal(err.Error())
	}

	appCacheDir = filepath.Join(userCacheDir, global.AppName)
	global.LOG.Debugf("Use base cache dir %s", appCacheDir)

	for _, sub := range subCacheDir {
		err = os.MkdirAll(filepath.Join(appCacheDir, sub), osutil.DefaultDirMode)
		if err != nil {
			global.LOG.Fatal(err.Error())
		}
	}
}

func Images(sub string) string {
	return filepath.Join(appCacheDir, "images", sub)
}
func CloudInit(sub string) string {
	return filepath.Join(appCacheDir, "cloud-init", sub)
}
func Boot(sub string) string {
	return filepath.Join(appCacheDir, "boot", sub)
}
