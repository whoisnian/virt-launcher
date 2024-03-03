package global

import (
	"os"

	"github.com/whoisnian/glb/ansi"
	"github.com/whoisnian/glb/logger"
)

var LOG *logger.Logger

func SetupLogger() {
	if CFG.Debug {
		LOG = logger.New(logger.NewNanoHandler(os.Stderr, logger.NewOptions(
			logger.LevelDebug, ansi.IsSupported(os.Stderr.Fd()), true,
		)))
	} else {
		LOG = logger.New(logger.NewNanoHandler(os.Stderr, logger.NewOptions(
			logger.LevelInfo, ansi.IsSupported(os.Stderr.Fd()), false,
		)))
	}
}
