package third

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/whoisnian/glb/logger"
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
	for _, params := range [][]string{
		{filepath.Join(cacheDir, "meta-data"), metaDataContent(timeStr)},
		{filepath.Join(cacheDir, "user-data"), userDataContent()},
	} {
		if _, err := createFileWith(params[0], params[1]); err != nil {
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

func createFileWith(filePath string, content string) (int, error) {
	fi, err := os.Create(filePath)
	if err != nil {
		return 0, err
	}
	defer fi.Close()

	return fi.WriteString(content)
}

func metaDataContent(timeStr string) string {
	sb := &strings.Builder{}
	sb.WriteString(fmt.Sprintf("instance-id: i-%s\n", timeStr))
	sb.WriteString(fmt.Sprintf("local-hostname: %s\n", global.CFG.Name))
	return sb.String()
}

func userDataContent() string {
	sb := &strings.Builder{}
	sb.WriteString("#cloud-config\n")
	sb.WriteString("power_state:\n")
	sb.WriteString("  delay: now\n")
	sb.WriteString("  mode: poweroff\n")
	sb.WriteString("  message: Powering off\n")
	sb.WriteString("  timeout: 30\n")
	sb.WriteString("  condition: true\n")
	if global.CFG.Pass != "" {
		sb.WriteString("ssh_pwauth: true\n")
		sb.WriteString(fmt.Sprintf("password: %s\n", global.CFG.Pass))
	}
	if global.CFG.Key != "" {
		sb.WriteString("ssh_authorized_keys:\n")
		sb.WriteString(fmt.Sprintf("  - %s\n", global.CFG.Key))
	}
	return sb.String()
}
