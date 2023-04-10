package third

import (
	"os/exec"
	"strings"

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

func resolveBinaryPath(binary *string) {
	if strings.Contains(*binary, "|") {
		for _, bin := range strings.Split(*binary, "|") {
			if _, err := exec.LookPath(bin); err == nil {
				*binary = bin
			}
		}
	}

	path, err := exec.LookPath(*binary)
	if err != nil {
		logger.Warn("LookPath(", *binary, "): ", err)
		enableDryRun()
	} else {
		logger.Debug("LookPath(", *binary, "): ", path)
		*binary = path
	}
}

func Setup() {
	resolveBinaryPath(&qemuImgBinary)
	resolveBinaryPath(&virtInstallBinary)
	resolveBinaryPath(&genisoimageBinary)
	resolveBinaryPath(&virshBinary)
}
