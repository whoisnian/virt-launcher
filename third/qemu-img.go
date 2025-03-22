package third

import (
	"context"
	"os/exec"

	"github.com/whoisnian/virt-launcher/global"
)

var qemuImgBinary = "qemu-img"

func ResizeImage(ctx context.Context, imagePath string) ([]byte, error) {
	cmd := exec.Command(qemuImgBinary, "resize", imagePath, global.CFG.Size)
	if global.CFG.DryRun {
		global.LOG.Infof(ctx, "[DRY-RUN] %s", cmd.String())
		return nil, nil
	} else {
		global.LOG.Debug(ctx, cmd.String())
		return cmd.CombinedOutput()
	}
}
