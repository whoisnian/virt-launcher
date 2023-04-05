package image

import (
	"os"
	"path/filepath"

	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/virt-launcher/global"
)

var imagesCacheDirName = ""

func SetupCache() {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		logger.Fatal(err)
	}

	imagesCacheDirName = filepath.Join(userCacheDir, global.AppName, "images")
	logger.Debug("Use images cache dir ", imagesCacheDirName)
	err = os.MkdirAll(imagesCacheDirName, os.ModePerm)
	if err != nil {
		logger.Fatal(err)
	}
}

func (img *Image) CacheFilePath() string {
	fileName := filepath.Base(img.Url)
	return filepath.Join(imagesCacheDirName, fileName)
}
