package third

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"regexp"
	"time"

	"github.com/whoisnian/virt-launcher/global"
)

var virshBinary = "virsh"

var spaceReg = regexp.MustCompile(`\s+`)
var stateReg = regexp.MustCompile(`state.state=(\d+)`)
var stateMap = map[string]string{
	"0": "none",
	"1": "running",
	"2": "blocked",
	"3": "paused",
	"4": "shutting down",
	"5": "shut off",
	"6": "crashed",
	"7": "suspended",
}

func WaitForVMOff(ctx context.Context) (output []byte, err error) {
	args := []string{"--connect", global.CFG.Connect, "domstats", "--state", "--domain", global.CFG.Name}
	if global.CFG.DryRun {
		global.LOG.Infof(ctx, "[DRY-RUN] %s", exec.Command(virshBinary, args...).String())
		return nil, nil
	}

	for range 20 {
		cmd := exec.Command(virshBinary, args...)
		global.LOG.Debug(ctx, cmd.String())

		output, err = cmd.CombinedOutput()
		if err != nil {
			return output, err
		}
		global.LOG.Debug(ctx, string(bytes.TrimSpace(spaceReg.ReplaceAll(output, []byte{' '}))))
		matches := stateReg.FindSubmatch(output)
		if len(matches) < 2 {
			return output, errors.New("invalid domain state")
		}
		global.LOG.Infof(ctx, "wait for domain off. current state: %s", stateMap[string(matches[1])])
		if bytes.Equal(matches[1], []byte("5")) {
			return output, err
		}
		time.Sleep(time.Second * 3)
	}
	return nil, nil
}

func DetachCloudInitIso(ctx context.Context, isoPath string) ([]byte, error) {
	cmd := exec.Command(virshBinary, "--connect", global.CFG.Connect, "detach-disk", "--persistent", "--domain", global.CFG.Name, isoPath)
	if global.CFG.DryRun {
		global.LOG.Infof(ctx, "[DRY-RUN] %s", cmd.String())
		return nil, nil
	} else {
		global.LOG.Debug(ctx, cmd.String())
		return cmd.CombinedOutput()
	}
}

func StartVM(ctx context.Context) ([]byte, error) {
	cmd := exec.Command(virshBinary, "--connect", global.CFG.Connect, "start", "--domain", global.CFG.Name)
	if global.CFG.DryRun {
		global.LOG.Infof(ctx, "[DRY-RUN] %s", cmd.String())
		return nil, nil
	} else {
		global.LOG.Debug(ctx, cmd.String())
		return cmd.CombinedOutput()
	}
}
