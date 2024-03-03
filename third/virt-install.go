package third

import (
	"os/exec"

	"github.com/whoisnian/virt-launcher/global"
)

var virtInstallBinary = "virt-install"

func CreateVM(disk, cdrom string) ([]byte, error) {
	cmd := exec.Command(virtInstallBinary,
		"--import",
		"--name", global.CFG.Name,
		"--osinfo", global.CFG.Os,
		"--disk", disk,
		"--disk", cdrom,
		"--vcpus", global.CFG.Cpu,
		"--memory", global.CFG.Mem,
		"--virt-type", "kvm",
		"--graphics", "none",
		"--noautoconsole",
		"--connect", global.CFG.Connect,
	)
	if global.CFG.DryRun {
		global.LOG.Infof("[DRY-RUN] %s", cmd.String())
		return nil, nil
	} else {
		global.LOG.Debug(cmd.String())
		return cmd.CombinedOutput()
	}
}
