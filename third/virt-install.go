package third

import (
	"os/exec"

	"github.com/whoisnian/glb/logger"
)

var virtInstallBinary = "virt-install"

func CreateVM(name string, os string, imagePath string) {
	cmd := exec.Command(virtInstallBinary,
		"--name", name,
		"--install", os,
		"--disk", imagePath,
	)
	logger.Info(cmd.String())
}
