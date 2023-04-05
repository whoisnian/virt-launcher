package third

import (
	"os/exec"

	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/virt-launcher/global"
)

var virtInstallBinary = "virt-install"

func CreateVM(disk string) {
	cmd := exec.Command(virtInstallBinary,
		"--import",
		"--name", global.CFG.Name,
		"--os-variant", global.CFG.Os,
		"--disk", disk,
		"--vcpus", global.CFG.Cpu,
		"--memory", global.CFG.Mem,
		"--virt-type", "kvm",
		"--graphics", "none",
		"--noautoconsole",
		"--connect", global.CFG.Connect,
		// --cloud-init
	)
	if global.CFG.DryRun {
		logger.Info("[DRY-RUN] ", cmd.String())
	} else {
		cmd.Run()
	}
}
