package third

import (
	"os/exec"
	"strings"

	"github.com/whoisnian/virt-launcher/global"
)

func enableDryRun() {
	if global.CFG.DryRun {
		return
	}
	global.LOG.Warn("Automatically enable dry-run mode")
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
		global.LOG.Warn("LookPath(" + *binary + "): " + err.Error())
		enableDryRun()
	} else {
		global.LOG.Debug("LookPath(" + *binary + "): " + path)
		*binary = path
	}
}

func Setup() {
	resolveBinaryPath(&qemuImgBinary)
	resolveBinaryPath(&virtInstallBinary)
	resolveBinaryPath(&genisoimageBinary)
	resolveBinaryPath(&virshBinary)
}
