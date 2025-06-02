package third

import (
	"context"
	"os/exec"

	"github.com/whoisnian/virt-launcher/global"
)

var virtInstallBinary = "virt-install"

func setupVirtInstall(ctx context.Context) {
	resolveBinaryPath(ctx, &virtInstallBinary)
}

var archMap = map[string]string{
	"386":     "i386",
	"amd64":   "x86_64",
	"arm":     "arm",
	"arm64":   "aarch64",
	"loong64": "loongarch64",
	"riscv64": "riscv64",
}

func CreateVM(ctx context.Context, diskVolume, cdromVolume string) ([]byte, error) {
	arch, ok := archMap[global.CFG.Arch]
	if !ok {
		arch = global.CFG.Arch
	}

	cmd := exec.Command(virtInstallBinary, "--connect", global.CFG.Connect,
		"--import",
		"--name", global.CFG.Name,
		"--osinfo", global.CFG.Os,
		"--arch", arch,
		"--vcpus", global.CFG.Cpu,
		"--memory", global.CFG.Mem,
		"--boot", global.CFG.Boot,
		"--disk", "vol="+global.CFG.Storage+"/"+diskVolume,
		"--disk", "source.startupPolicy=optional,vol="+global.CFG.Storage+"/"+cdromVolume,
		"--network", "network="+global.CFG.Network,
		"--graphics", "none",
		"--video", "virtio",
		"--noautoconsole",
	)
	return prepareOrCombinedOutput(ctx, cmd)
}
