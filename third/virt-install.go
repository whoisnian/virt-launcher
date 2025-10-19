package third

import (
	"bytes"
	"context"
	"os/exec"
	"regexp"

	"github.com/whoisnian/virt-launcher/global"
)

var virtInstallBinary = "virt-install"

func setupVirtInstall(ctx context.Context) {
	resolveBinaryPath(ctx, &virtInstallBinary)
}

var genericLinuxReg = regexp.MustCompile(`^linux\d{4}$`) // linux2024/linux2022/linux2020/linux2018/linux2016

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

	os := []byte(global.CFG.Os)
	if !global.CFG.Prepare {
		listCmd := exec.Command(virtInstallBinary, "--osinfo", "list")
		if output, err := stdoutOrErrorWithStderr(ctx, listCmd); err == nil {
			// https://github.com/virt-manager/virt-manager/blob/1ead880b2e51ae3fab5e103c05fd9cb1c921ec89/virtinst/cli.py#L5373
			//   print(", ".join(osobj.all_names))
			// So we split the output by comma and newline, and check if the OS is known
			known := false
			genericLinux := []byte{}
			for l, r := 0, 0; r < len(output); r++ {
				if output[r] == ',' || output[r] == '\n' {
					if item := bytes.TrimSpace(output[l:r]); bytes.Equal(item, os) {
						known = true
						break
					} else if genericLinuxReg.Match(item) && bytes.Compare(item, genericLinux) > 0 {
						genericLinux = item
					}
					l = r + 1
				}
			}
			if len(genericLinux) == 0 {
				genericLinux = []byte("generic")
			}
			if !known {
				global.LOG.Warnf(ctx, "unknown os name '%s' for virt-install, fallback to '%s'", global.CFG.Os, genericLinux)
				os = genericLinux
			}
		}
	}

	args := []string{
		"--connect", global.CFG.Connect,
		"--import",
		"--name", global.CFG.Name,
		"--osinfo", string(os),
		"--arch", arch,
		"--vcpus", global.CFG.Cpu,
		"--memory", global.CFG.Mem,
		"--disk", "vol=" + global.CFG.Storage + "/" + diskVolume,
		"--disk", "source.startupPolicy=optional,vol=" + global.CFG.Storage + "/" + cdromVolume,
		"--network", "network=" + global.CFG.Network,
		"--graphics", "none",
		"--video", "virtio",
		"--noautoconsole",
	}
	if global.CFG.Boot != "" {
		args = append(args, "--boot", global.CFG.Boot)
	}
	if global.CFG.Cpum != "" {
		args = append(args, "--cpu", global.CFG.Cpum)
	}

	cmd := exec.Command(virtInstallBinary, args...)
	return prepareOrCombinedOutput(ctx, cmd)
}
