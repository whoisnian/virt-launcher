package image

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
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
	pw.Show()
	return n, err
}

func (pw *progressWriter) Show() {
	if time.Since(pw.last) < 3*time.Second && pw.cur < pw.sum {
		return
	}
	if pw.sum > 0 {
		logger.Info(fmt.Sprintf("%3d MiB of %3d MiB downloaded (%d%%)", pw.cur/1024/1024, pw.sum/1024/1024, pw.cur*100/pw.sum))
	} else {
		logger.Info(fmt.Sprintf("%3d MiB downloaded", pw.cur/1024/1024))
	}
	pw.last = time.Now()
}

func (img *Image) Download() error {
	fileName := path.Base(img.Url)
	filePath := path.Join(cacheDirName, fileName)

	logger.Info("Start downloading image to ", filePath)
	rHash, err := img.RemoteHash()
	if err != nil {
		return err
	}
	logger.Debug("File remote hash: ", rHash)

	if _, err := os.Stat(filePath); err == nil {
		logger.Info("File already exists, start verifying")
		lHash, err := img.LocalHash(filePath)
		if err != nil {
			return err
		}
		logger.Debug("File local hash: ", rHash)
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

	fi, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fi.Close()

	resp, err := http.Get(img.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	length, err := strconv.Atoi(resp.Header.Get("content-length"))
	if err != nil {
		logger.Debug("Content-Length atoi ", err)
	}

	_, err = io.CopyBuffer(&progressWriter{w: fi, sum: length}, resp.Body, make([]byte, 4*1024*1024))
	if err != nil {
		return err
	}
	fi.Close()

	lHash, err := img.LocalHash(filePath)
	if err != nil {
		return err
	}
	logger.Debug("File local hash: ", rHash)
	if lHash == rHash {
		logger.Info("Hash verification ok")
		return nil
	} else {
		return errors.New("hash verification failed")
	}
}
