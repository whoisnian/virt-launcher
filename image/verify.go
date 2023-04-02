package image

import (
	"bufio"
	"crypto"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/whoisnian/glb/logger"
)

func (img *Image) HashType() crypto.Hash {
	url := strings.ToLower(img.Hash)
	if strings.Contains(url, "sha256") {
		return crypto.SHA256
	} else if strings.Contains(url, "sha512") {
		return crypto.SHA512
	}
	return 0
}

func (img *Image) RemoteHash() (string, error) {
	logger.Debug("Get remote hash from ", img.Hash)
	resp, err := http.Get(img.Hash)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fileName := path.Base(img.Url)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, fileName) {
			return strings.TrimSpace(strings.ReplaceAll(line, fileName, "")), nil
		}
	}
	return "", errors.New("remote hash not found")
}

func (img *Image) LocalHash(filePath string) (string, error) {
	hashType := img.HashType()

	var hasher hash.Hash
	if hashType == crypto.SHA256 {
		hasher = sha256.New()
	} else if hashType == crypto.SHA512 {
		hasher = sha512.New()
	} else {
		return "", errors.New("unknown hash type")
	}

	logger.Debug("Calc local hash from ", filePath)
	fi, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	_, err = io.CopyBuffer(hasher, fi, make([]byte, 4*1024*1024))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
