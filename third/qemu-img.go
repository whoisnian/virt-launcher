package third

import (
	"os/exec"

	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/virt-launcher/global"
)

var qemuImgBinary = "qemu-img"

func ResizeImage(imagePath string) ([]byte, error) {
	cmd := exec.Command(qemuImgBinary, "resize", imagePath, global.CFG.Size)
	if global.CFG.DryRun {
		logger.Info("[DRY-RUN] ", cmd.String())
		return nil, nil
	} else {
		logger.Debug(cmd.String())
		return cmd.CombinedOutput()
	}
}
