package third

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/virt-launcher/global"
)

var genisoimageBinary = "xorrisofs|genisoimage"

func CreateCloudInitIso(cacheDir, isoPath string) ([]byte, error) {
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return nil, err
	}
	if !global.CFG.DryRun {
		defer os.RemoveAll(cacheDir)
	}

	logger.Debug("Start writing cloud-init data files to ", cacheDir)
	for _, params := range [][]string{
		{filepath.Join(cacheDir, "meta-data"), metaDataContent()},
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

func metaDataContent() string {
	return fmt.Sprintf(strings.TrimSpace(`
instance-id: i-%x
local-hostname: %s
`), time.Now().UnixMilli(), global.CFG.Name)
}

func userDataContent() string {
	if global.CFG.Key == "" {
		return "#cloud-config"
	} else {
		return fmt.Sprintf(strings.TrimSpace(`
#cloud-config
ssh_authorized_keys:
  - %s
`), global.CFG.Key)
	}
}
