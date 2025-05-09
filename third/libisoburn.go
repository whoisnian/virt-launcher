package third

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/whoisnian/glb/util/osutil"
	"github.com/whoisnian/virt-launcher/global"
)

// https://github.com/virt-manager/virt-manager/blob/d17731aea1fd8f7fa926253cc40a1c777264da07/virtinst/install/installerinject.py#L52
var xorrisofsBinary = "xorrisofs|genisoimage|mkisofs"

func setupLibisoburn(ctx context.Context) {
	resolveBinaryPath(ctx, &xorrisofsBinary)
}

func CreateCloudInitIso(ctx context.Context, cacheDir, isoPath, instanceID string) ([]byte, error) {
	if err := os.MkdirAll(cacheDir, osutil.DefaultDirMode); err != nil {
		return nil, err
	}
	if !global.CFG.Prepare {
		defer os.RemoveAll(cacheDir)
	}

	global.LOG.Debugf(ctx, "start writing cloud-init data files to %s", cacheDir)
	for _, param := range []struct {
		name string
		data []byte
	}{
		{filepath.Join(cacheDir, "meta-data"), metaDataContent(instanceID)},
		{filepath.Join(cacheDir, "user-data"), userDataContent()},
	} {
		if err := os.WriteFile(param.name, param.data, osutil.DefaultFileMode); err != nil {
			return nil, err
		}
	}

	global.LOG.Debugf(ctx, "start writing cloud-init iso file to %s", isoPath)
	cmd := exec.Command(xorrisofsBinary, "-output", isoPath, "-volid", "cidata", "-joliet", "-input-charset", "utf8", "-rational-rock", cacheDir)
	return prepareOrCombinedOutput(ctx, cmd)
}

func metaDataContent(instanceID string) []byte {
	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "instance-id: %s\n", instanceID)
	fmt.Fprintf(buf, "local-hostname: %s\n", global.CFG.Name)
	return buf.Bytes()
}

func userDataContent() []byte {
	buf := &bytes.Buffer{}
	buf.WriteString("#cloud-config\n")
	buf.WriteString("power_state:\n")
	buf.WriteString("  delay: now\n")
	buf.WriteString("  mode: poweroff\n")
	buf.WriteString("  message: Powering off\n")
	buf.WriteString("  timeout: 30\n")
	buf.WriteString("  condition: true\n")
	if global.CFG.Pass != "" {
		buf.WriteString("ssh_pwauth: true\n")
		fmt.Fprintf(buf, "password: %s\n", global.CFG.Pass)
	}
	if global.CFG.Key != "" {
		buf.WriteString("ssh_authorized_keys:\n")
		fmt.Fprintf(buf, "  - %s\n", global.CFG.Key)
	}
	return buf.Bytes()
}
