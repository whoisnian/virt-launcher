package third

import (
	"bytes"
	"errors"
	"os/exec"
	"regexp"
	"time"

	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/virt-launcher/global"
)

var virshBinary = "virsh"

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
		logger.Info("[DRY-RUN] ", exec.Command(virshBinary, args...).String())
		return nil, nil
	}

	for i := 0; i < 30; i++ {
		cmd := exec.Command(virshBinary, args...)
		logger.Debug(cmd.String())

		output, err = cmd.CombinedOutput()
		if err != nil {
			return output, err
		}
		logger.Debug(string(output))
		matches := stateReg.FindSubmatch(output)
		if len(matches) < 2 {
			return output, errors.New("invalid domain state")
		}
		logger.Info("Wait for domain off. Current state: ", stateMap[string(matches[1])])
		if bytes.Equal(matches[1], []byte("5")) {
			return output, err
		}
		time.Sleep(time.Second * 2)
	}
	return nil, nil
}

func DetachCloudInitIso(isoPath string) ([]byte, error) {
	cmd := exec.Command(virshBinary, "--connect", "qemu:///system", "detach-disk", "--persistent", "--domain", global.CFG.Name, isoPath)
	if global.CFG.DryRun {
		logger.Info("[DRY-RUN] ", cmd.String())
		return nil, nil
	} else {
		logger.Debug(cmd.String())
		return cmd.CombinedOutput()
	}
}

func StartVM() ([]byte, error) {
	cmd := exec.Command(virshBinary, "--connect", "qemu:///system", "start", "--domain", global.CFG.Name)
	if global.CFG.DryRun {
		logger.Info("[DRY-RUN] ", cmd.String())
		return nil, nil
	} else {
		logger.Debug(cmd.String())
		return cmd.CombinedOutput()
	}
}
