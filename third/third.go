package third

import (
	"context"
	"os/exec"
	"strings"

	"github.com/whoisnian/virt-launcher/global"
)

func enableDryRun(ctx context.Context) {
	if global.CFG.DryRun {
		return
	}
	global.LOG.Warn(ctx, "automatically enable dry-run mode")
	global.CFG.DryRun = true
}

func resolveBinaryPath(ctx context.Context, binary *string) {
	if strings.Contains(*binary, "|") {
		for bin := range strings.SplitSeq(*binary, "|") {
			if _, err := exec.LookPath(bin); err == nil {
				*binary = bin
			}
		}
	}

	path, err := exec.LookPath(*binary)
	if err != nil {
		global.LOG.Warnf(ctx, "exec.LookPath(%s): %v", *binary, err)
		enableDryRun(ctx)
	} else {
		global.LOG.Debugf(ctx, "exec.LookPath(%s): %s", *binary, path)
		*binary = path
	}
}

func Setup(ctx context.Context) {
	resolveBinaryPath(ctx, &qemuImgBinary)
	resolveBinaryPath(ctx, &virtInstallBinary)
	resolveBinaryPath(ctx, &genisoimageBinary)
	resolveBinaryPath(ctx, &virshBinary)
}
