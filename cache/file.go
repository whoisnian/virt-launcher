package cache

import (
	"io"
	"os"

	"github.com/whoisnian/glb/logger"
)

func CopyFile(fromPath, toPath string) error {
	from, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer from.Close()

	to, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer to.Close()

	logger.Debug("Start copying file from ", fromPath, " to ", toPath)
	_, err = io.CopyBuffer(to, from, make([]byte, 4*1024*1024))
	return err
}
