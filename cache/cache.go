package cache

import (
	"os"
	"path/filepath"

	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/virt-launcher/global"
)

var appCacheDir = ""
var subCacheDir = []string{"images", "cloud-init", "boot"}

func Setup() {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		logger.Fatal(err)
	}

	appCacheDir = filepath.Join(userCacheDir, global.AppName)
	logger.Debug("Use base cache dir ", appCacheDir)

	for _, sub := range subCacheDir {
		err = os.MkdirAll(filepath.Join(appCacheDir, sub), os.ModePerm)
		if err != nil {
			logger.Fatal(err)
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
