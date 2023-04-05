package third

import (
	"os/exec"

	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/virt-launcher/global"
)

func enableDryRun() {
	if global.CFG.DryRun {
		return
	}
	logger.Warn("Automatically enable dry-run mode")
	global.CFG.DryRun = true
}

func updateBinaryPath(binary *string) {
	path, err := exec.LookPath(*binary)
	if err != nil {
		logger.Warn("LookPath(", *binary, "): ", err)
		enableDryRun()
	} else {
		logger.Debug("LookPath(", *binary, "): ", path)
		*binary = path
	}
}

func SetupThird() {
	updateBinaryPath(&qemuImgBinary)
	updateBinaryPath(&virtInstallBinary)
}
