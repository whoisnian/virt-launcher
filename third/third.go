package third

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/whoisnian/virt-launcher/global"
)

func Setup(ctx context.Context) {
	setupLibvirt(ctx)
	setupLibisoburn(ctx)
	setupVirtInstall(ctx)
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
		global.LOG.Warnf(ctx, "exec.LookPath(%s): %v, switch to prepare mode", *binary, err)
		global.CFG.Prepare = true
	} else {
		global.LOG.Debugf(ctx, "exec.LookPath(%s): %s", *binary, path)
		*binary = path
	}
}

func prepareOrCombinedOutput(ctx context.Context, cmd *exec.Cmd) ([]byte, error) {
	if global.CFG.Prepare {
		global.LOG.Infof(ctx, "[NOTE] %s", cmd.String())
		return nil, nil
	}

	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, "LC_ALL=C")
	global.LOG.Debug(ctx, cmd.String())
	return cmd.CombinedOutput()
}

func stdoutOrErrorWithStderr(ctx context.Context, cmd *exec.Cmd) ([]byte, error) {
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, "LC_ALL=C")
	global.LOG.Debug(ctx, cmd.String())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stderr.Bytes(), err
	}
	return stdout.Bytes(), nil
}
