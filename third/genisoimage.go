package third

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/glb/util/osutil"
	"github.com/whoisnian/virt-launcher/global"
)

var genisoimageBinary = "xorrisofs|genisoimage"

func CreateCloudInitIso(cacheDir, isoPath, timeStr string) ([]byte, error) {
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return nil, err
	}
	if !global.CFG.DryRun {
		defer os.RemoveAll(cacheDir)
	}

	logger.Debug("Start writing cloud-init data files to ", cacheDir)
	for _, param := range []struct {
		name string
		data []byte
	}{
		{filepath.Join(cacheDir, "meta-data"), metaDataContent(timeStr)},
		{filepath.Join(cacheDir, "user-data"), userDataContent()},
	} {
		if err := os.WriteFile(param.name, param.data, osutil.DefaultFileMode); err != nil {
			return nil, err
		}
	}

	logger.Debug("Start creating cloud-init iso file to ", isoPath)
	cmd := exec.Command(genisoimageBinary, "-output", isoPath, "-volid", "cidata", "-joliet", "-input-charset", "utf8", "-rational-rock", cacheDir)
	if global.CFG.DryRun {
		logger.Info("[DRY-RUN] ", cmd.String())
		return nil, nil
	} else {
		logger.Debug(cmd.String())
		return cmd.CombinedOutput()
	}
}

func metaDataContent(timeStr string) []byte {
	buf := &bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("instance-id: i-%s\n", timeStr))
	buf.WriteString(fmt.Sprintf("local-hostname: %s\n", global.CFG.Name))
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
		buf.WriteString(fmt.Sprintf("password: %s\n", global.CFG.Pass))
	}
	if global.CFG.Key != "" {
		buf.WriteString("ssh_authorized_keys:\n")
		buf.WriteString(fmt.Sprintf("  - %s\n", global.CFG.Key))
	}
	return buf.Bytes()
}
