package image

import (
	"bufio"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/whoisnian/glb/logger"
)

var hexReg = regexp.MustCompile(`[A-Za-z0-9]{64,128}`)

func (img *Image) Hasher() (hash.Hash, error) {
	url := strings.ToLower(img.Hash)
	if strings.Contains(url, "sha256") {
		return sha256.New(), nil
	} else if strings.Contains(url, "sha512") {
		return sha512.New(), nil
	} else {
		return nil, errors.New("unknown hash type")
	}
}

func (img *Image) RemoteHash() (string, error) {
	logger.Debug("Get remote hash from ", img.Hash)
	if strings.HasPrefix(img.Hash, "https://") || strings.HasPrefix(img.Hash, "http://") {
		resp, err := http.Get(img.Hash)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		fileName := filepath.Base(img.Url)
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.Contains(line, fileName) {
				continue
			}
			if res := hexReg.FindString(line); res != "" {
				return res, nil
			}
		}
	} else if strings.HasPrefix(img.Hash, "sha256sum:") || strings.HasPrefix(img.Hash, "sha512sum:") {
		return hexReg.FindString(img.Hash), nil
	}

	return "", errors.New("remote hash not found")
}

func (img *Image) LocalHashFrom(filePath string) (string, error) {
	hasher, err := img.Hasher()
	if err != nil {
		return "", err
	}

	logger.Debug("Calc local hash from ", filePath)
	fi, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer fi.Close()

	_, err = io.CopyBuffer(hasher, fi, make([]byte, 4*1024*1024))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
