package third

import (
	"bytes"
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

func WaitForVMOff() (output []byte, err error) {
	args := []string{"--connect", "qemu:///system", "domstats", "--state", "--domain", global.CFG.Name}
	if global.CFG.DryRun {
		global.LOG.Infof("[DRY-RUN] %s", exec.Command(virshBinary, args...).String())
		return nil, nil
	}

	for i := 0; i < 20; i++ {
		cmd := exec.Command(virshBinary, args...)
		global.LOG.Debug(cmd.String())

		output, err = cmd.CombinedOutput()
		if err != nil {
			return output, err
		}
		global.LOG.Debug(string(spaceReg.ReplaceAll(output, []byte{' '})))
		matches := stateReg.FindSubmatch(output)
		if len(matches) < 2 {
			return output, errors.New("invalid domain state")
		}
		global.LOG.Infof("Wait for domain off. Current state: %s", stateMap[string(matches[1])])
		if bytes.Equal(matches[1], []byte("5")) {
			return output, err
		}
		time.Sleep(time.Second * 3)
	}
	return nil, nil
}

func DetachCloudInitIso(isoPath string) ([]byte, error) {
	cmd := exec.Command(virshBinary, "--connect", "qemu:///system", "detach-disk", "--persistent", "--domain", global.CFG.Name, isoPath)
	if global.CFG.DryRun {
		global.LOG.Infof("[DRY-RUN] %s", cmd.String())
		return nil, nil
	} else {
		global.LOG.Debug(cmd.String())
		return cmd.CombinedOutput()
	}
}

func StartVM() ([]byte, error) {
	cmd := exec.Command(virshBinary, "--connect", "qemu:///system", "start", "--domain", global.CFG.Name)
	if global.CFG.DryRun {
		global.LOG.Infof("[DRY-RUN] %s", cmd.String())
		return nil, nil
	} else {
		global.LOG.Debug(cmd.String())
		return cmd.CombinedOutput()
	}
}
