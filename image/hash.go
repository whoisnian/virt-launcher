package image

import (
	"bufio"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/whoisnian/virt-launcher/global"
)

func (img *Image) Hasher() (hash.Hash, error) {
	if strings.HasPrefix(img.HashFmt, "sha512sum:") {
		return sha512.New(), nil
	} else if strings.HasPrefix(img.HashFmt, "sha256sum:") {
		return sha256.New(), nil
	} else if strings.HasPrefix(img.HashFmt, "sha1sum:") {
		return sha1.New(), nil
	} else if strings.HasPrefix(img.HashFmt, "md5sum:") {
		return md5.New(), nil
	} else {
		return nil, errors.New("unknown hash type")
	}
}

func (img *Image) LocalHashFrom(ctx context.Context, filePath string) (string, error) {
	global.LOG.Debugf(ctx, "calc local hash from %s", filePath)
	hasher, err := img.Hasher()
	if err != nil {
		return "", err
	}

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

func (img *Image) HashMatcher() (matcher func(string) (string, bool), err error) {
	parts := strings.Split(img.HashFmt, ":")
	if len(parts) != 2 {
		return nil, errors.New("unknown hash format")
	}
	var regex *regexp.Regexp
	switch parts[0] {
	case "sha512sum":
		regex = regexp.MustCompile(`[A-Fa-f0-9]{128}`)
	case "sha256sum":
		regex = regexp.MustCompile(`[A-Fa-f0-9]{64}`)
	case "sha1sum":
		regex = regexp.MustCompile(`[A-Fa-f0-9]{40}`)
	case "md5sum":
		regex = regexp.MustCompile(`[A-Fa-f0-9]{32}`)
	default:
		return nil, errors.New("unknown hash type")
	}

	switch parts[1] {
	case "raw": // e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
		return func(text string) (string, bool) {
			if res := regex.FindAllString(text, -1); res != nil {
				return res[0], true
			}
			return "", false
		}, nil
	case "gnu": // e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  empty.txt
		return func(text string) (string, bool) {
			if strings.HasSuffix(text, img.BaseName()) {
				if res := regex.FindAllString(text, -1); res != nil {
					return res[0], true
				}
			}
			return "", false
		}, nil
	case "bsd": // SHA256 (empty.txt) = e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
		return func(text string) (string, bool) {
			if strings.Contains(text, "("+img.BaseName()+")") {
				if res := regex.FindAllString(text, -1); res != nil {
					return res[len(res)-1], true
				}
			}
			return "", false
		}, nil
	default:
		return nil, errors.New("unknown hash style")
	}
}

func (img *Image) UpdateHashVal(ctx context.Context) error {
	global.LOG.Debugf(ctx, "get remote hash from %s", img.HashUrl)
	matcher, err := img.HashMatcher()
	if err != nil {
		return err
	}

	resp, err := requestGet(ctx, img.HashUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		if res, ok := matcher(scanner.Text()); ok {
			img.HashVal = res
			return nil
		}
	}
	return errors.New("remote hash not found")
}
