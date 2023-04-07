package image

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/whoisnian/glb/logger"
)

type progressWriter struct {
	w    io.Writer
	cur  int
	sum  int
	last time.Time
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.w.Write(p)
	pw.cur += n
	pw.showProgress()
	return n, err
}

func (pw *progressWriter) showProgress() {
	if time.Since(pw.last) < 2*time.Second && pw.cur < pw.sum {
		return
	}
	if pw.sum > 0 {
		logger.Info(fmt.Sprintf("%3d MiB of %3d MiB downloaded (%d%%)", pw.cur/1024/1024, pw.sum/1024/1024, pw.cur*100/pw.sum))
	} else {
		logger.Info(fmt.Sprintf("%3d MiB downloaded", pw.cur/1024/1024))
	}
	pw.last = time.Now()
}

func (img *Image) Download(filePath string) error {
	rHash, err := img.RemoteHash()
	if err != nil {
		return err
	}
	logger.Debug("File remote hash: ", rHash)

	// check image file exists
	if _, err := os.Stat(filePath); err == nil {
		logger.Info("File already exists, start verifying")
		lHash, err := img.LocalHashFrom(filePath)
		if err != nil {
			return err
		}
		logger.Debug("File local hash: ", lHash)
		if lHash == rHash {
			logger.Info("Hash verification ok, skip downloading")
			return nil
		} else {
			logger.Warn("Hash verification failed, delete local file")
			if err := os.Remove(filePath); err != nil {
				return err
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	// download image file
	if err = img.download(filePath); err != nil {
		return err
	}

	// verify image file
	lHash, err := img.LocalHashFrom(filePath)
	if err != nil {
		return err
	}
	logger.Debug("File local hash: ", lHash)
	if lHash == rHash {
		logger.Info("Hash verification ok")
		return nil
	} else {
		return errors.New("hash verification failed")
	}
}

func (img *Image) download(filePath string) error {
	fi, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fi.Close()

	logger.Debug("Start downloading image from ", img.Url)
	resp, err := http.Get(img.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	length, err := strconv.Atoi(resp.Header.Get("content-length"))
	if err != nil {
		logger.Debug("Content-Length atoi ", err)
	} else {
		logger.Debug("Get Content-Length ", length)
	}

	_, err = io.CopyBuffer(&progressWriter{w: fi, sum: length}, resp.Body, make([]byte, 4*1024*1024))
	if err != nil {
		return err
	}
	return nil
}
