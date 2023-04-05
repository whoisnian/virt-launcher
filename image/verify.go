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
	"strings"

	"github.com/whoisnian/glb/logger"
)

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
	resp, err := http.Get(img.Hash)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fileName := filepath.Base(img.Url)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, fileName) {
			return strings.TrimSpace(strings.ReplaceAll(line, fileName, "")), nil
		}
	}
	return "", errors.New("remote hash not found")
}

func (img *Image) LocalHash() (string, error) {
	hasher, err := img.Hasher()
	if err != nil {
		return "", err
	}

	logger.Debug("Calc local hash from ", img.CacheFilePath())
	fi, err := os.Open(img.CacheFilePath())
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
