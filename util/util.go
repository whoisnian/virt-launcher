package util

import (
	"io"
	"os"

	"github.com/whoisnian/glb/logger"
)

func CopyToTemp(fromPath, pattern string) (string, error) {
	from, err := os.Open(fromPath)
	if err != nil {
		return "", err
	}
	defer from.Close()

	to, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	defer to.Close()

	logger.Debug("Start copying file from ", fromPath, " to ", to.Name())
	_, err = io.CopyBuffer(to, from, make([]byte, 4*1024*1024))
	return to.Name(), err
}
