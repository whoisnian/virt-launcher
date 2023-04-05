package third

import (
	"os/exec"

	"github.com/whoisnian/glb/logger"
)

var qemuImgBinary = "qemu-img"

func ResizeImage(imagePath string, size string) {
	cmd := exec.Command(qemuImgBinary, "resize", imagePath, size)
	logger.Info(cmd.String())
}
