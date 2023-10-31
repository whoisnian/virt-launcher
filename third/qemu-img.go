package third

import (
	"os/exec"

	"github.com/whoisnian/virt-launcher/global"
)

var qemuImgBinary = "qemu-img"

func ResizeImage(imagePath string) ([]byte, error) {
	cmd := exec.Command(qemuImgBinary, "resize", imagePath, global.CFG.Size)
	if global.CFG.DryRun {
		global.LOG.Info("[DRY-RUN] " + cmd.String())
		return nil, nil
	} else {
		global.LOG.Debug(cmd.String())
		return cmd.CombinedOutput()
	}
}
