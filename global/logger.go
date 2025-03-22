package global

import (
	"context"
	"os"

	"github.com/whoisnian/glb/ansi"
	"github.com/whoisnian/glb/logger"
)

var LOG *logger.Logger

func SetupLogger(_ context.Context) {
	if CFG.Debug {
		LOG = logger.New(logger.NewNanoHandler(os.Stderr, logger.Options{
			Level:     logger.LevelDebug,
			Colorful:  ansi.IsSupported(os.Stderr.Fd()),
			AddSource: true,
		}))
	} else {
		LOG = logger.New(logger.NewNanoHandler(os.Stderr, logger.Options{
			Level:     logger.LevelInfo,
			Colorful:  ansi.IsSupported(os.Stderr.Fd()),
			AddSource: false,
		}))
	}
}
