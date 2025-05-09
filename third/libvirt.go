package third

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/glb/util/fsutil"
	"github.com/whoisnian/virt-launcher/global"
)

var virshBinary = "virsh"
var libvirtDataRootDir string

func setupLibvirt(ctx context.Context) {
	resolveBinaryPath(ctx, &virshBinary)

	// https://github.com/virt-manager/virt-manager/blob/d17731aea1fd8f7fa926253cc40a1c777264da07/virtinst/connection.py#L388
	isPrivileged := true
	if u, err := url.Parse(global.CFG.Connect); err == nil {
		if u.Path == "/session" {
			isPrivileged = false
		} else if u.Path == "/embed" {
			isPrivileged = os.Getuid() == 0
		}
	}
	// https://github.com/virt-manager/virt-manager/blob/d17731aea1fd8f7fa926253cc40a1c777264da07/virtinst/connection.py#L186
	if isPrivileged {
		libvirtDataRootDir = "/var/lib/libvirt"
	} else if dir, ok := os.LookupEnv("XDG_DATA_HOME"); ok {
		libvirtDataRootDir = dir
	} else if dir, err := fsutil.ExpandHomeDir("~/.local/share/libvirt"); err == nil {
		libvirtDataRootDir = dir
	} else {
		global.LOG.Fatal(ctx, "fsutil.ExpandHomeDir", logger.Error(err))
	}
}

func EnsureNetworkIsActive(ctx context.Context) ([]byte, error) {
	if global.CFG.Prepare {
		return nil, nil
	}

	cmd := exec.Command(virshBinary, "--connect", global.CFG.Connect, "net-info", "--network", global.CFG.Network)
	output, err := stdoutOrErrorWithStderr(ctx, cmd)
	if err != nil {
		return output, err
	}

	activeReg := regexp.MustCompile(`Active:\s+(\w+)`)
	if matches := activeReg.FindSubmatch(output); len(matches) < 2 {
		return nil, errors.New("invalid network state")
	} else if bytes.Equal(matches[1], []byte("yes")) {
		global.LOG.Debugf(ctx, "network %s is already active", global.CFG.Network)
		return nil, nil
	} else {
		global.LOG.Debugf(ctx, "network %s is not active (%s)", global.CFG.Network, matches[1])
		cmd = exec.Command(virshBinary, "--connect", global.CFG.Connect, "net-start", "--network", global.CFG.Network)
		return prepareOrCombinedOutput(ctx, cmd)
	}
}

func EnsureStoragePoolIsActive(ctx context.Context) ([]byte, error) {
	if global.CFG.Prepare {
		return nil, nil
	}

	cmd := exec.Command(virshBinary, "--connect", global.CFG.Connect, "pool-info", "--pool", global.CFG.Storage)
	output, err := stdoutOrErrorWithStderr(ctx, cmd)
	if err != nil && global.CFG.Storage == "default" {
		target := filepath.Join(libvirtDataRootDir, "images")
		global.LOG.Warnf(ctx, "storage pool %s not found, create with target %s", global.CFG.Storage, target)
		cmd = exec.Command(virshBinary, "--connect", global.CFG.Connect,
			"pool-define-as",
			"--name", global.CFG.Storage,
			"--type", "dir",
			"--target", target,
		)
		if output, err = prepareOrCombinedOutput(ctx, cmd); err != nil {
			return output, err
		}
		cmd = exec.Command(virshBinary, "--connect", global.CFG.Connect, "pool-start", "--pool", global.CFG.Storage)
		return prepareOrCombinedOutput(ctx, cmd)
	} else if err != nil {
		return output, err
	}

	stateReg := regexp.MustCompile(`State:\s+(\w+)`)
	if matches := stateReg.FindSubmatch(output); len(matches) < 2 {
		return nil, errors.New("invalid storage pool state")
	} else if bytes.Equal(matches[1], []byte("running")) {
		global.LOG.Debugf(ctx, "storage pool %s is already running", global.CFG.Storage)
		return nil, nil
	} else {
		global.LOG.Debugf(ctx, "storage pool %s is not running (%s)", global.CFG.Storage, matches[1])
		cmd = exec.Command(virshBinary, "--connect", global.CFG.Connect, "pool-start", "--pool", global.CFG.Storage)
		return prepareOrCombinedOutput(ctx, cmd)
	}
}

func UploadVolume(ctx context.Context, volume, file string) ([]byte, error) {
	cmd := exec.Command(virshBinary, "--connect", global.CFG.Connect, "vol-create-as", "--pool", global.CFG.Storage, "--name", volume, "--capacity", "0")
	if output, err := prepareOrCombinedOutput(ctx, cmd); err != nil {
		return output, err
	}
	cmd = exec.Command(virshBinary, "--connect", global.CFG.Connect, "vol-upload", "--pool", global.CFG.Storage, "--vol", volume, "--file", file)
	return prepareOrCombinedOutput(ctx, cmd)
}

func ResizeVolume(ctx context.Context, volume string) ([]byte, error) {
	cmd := exec.Command(virshBinary, "--connect", global.CFG.Connect, "vol-resize", "--pool", global.CFG.Storage, "--vol", volume, "--capacity", global.CFG.Size)
	return prepareOrCombinedOutput(ctx, cmd)
}

func DeleteVolume(ctx context.Context, volume string) ([]byte, error) {
	cmd := exec.Command(virshBinary, "--connect", global.CFG.Connect, "vol-delete", "--pool", global.CFG.Storage, "--vol", volume)
	return prepareOrCombinedOutput(ctx, cmd)
}

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
	if global.CFG.Prepare {
		global.LOG.Infof(ctx, "[NOTE] %s", exec.Command(virshBinary, args...).String())
		return nil, nil
	}

	spaceReg := regexp.MustCompile(`\s+`)
	stateReg := regexp.MustCompile(`state.state=(\d+)`)
	for range 40 {
		cmd := exec.Command(virshBinary, args...)
		output, err = stdoutOrErrorWithStderr(ctx, cmd)
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
			return output, nil
		}
		time.Sleep(time.Second * 3)
	}
	return nil, errors.New("timeout waiting for domain off")
}

func StartVM(ctx context.Context) ([]byte, error) {
	cmd := exec.Command(virshBinary, "--connect", global.CFG.Connect, "start", "--domain", global.CFG.Name)
	return prepareOrCombinedOutput(ctx, cmd)
}
