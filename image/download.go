package image

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/whoisnian/glb/util/ioutil"
	"github.com/whoisnian/virt-launcher/global"
)

func (img *Image) Download(filePath string) error {
	rHash, err := img.RemoteHash()
	if err != nil {
		return err
	}
	global.LOG.Debugf("File remote hash: %s", rHash)

	// check image file exists
	if _, err := os.Stat(filePath); err == nil {
		global.LOG.Info("File already exists, start verifying")
		lHash, err := img.LocalHashFrom(filePath)
		if err != nil {
			return err
		}
		global.LOG.Debugf("File local hash: %s", lHash)
		if lHash == rHash {
			global.LOG.Info("Hash verification ok, skip downloading")
			return nil
		} else {
			global.LOG.Warn("Hash verification failed, delete local file")
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
	global.LOG.Debugf("File local hash: %s", lHash)
	if lHash == rHash {
		global.LOG.Info("Hash verification ok")
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

	global.LOG.Debugf("Start downloading image from %s", img.Url)
	resp, err := http.Get(img.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	length, err := strconv.Atoi(resp.Header.Get("content-length"))
	if err != nil {
		global.LOG.Debugf("Content-Length atoi %v", err)
	} else {
		global.LOG.Debugf("Get Content-Length %d", length)
	}

	wg := &sync.WaitGroup{}
	pw := ioutil.NewProgressWriter(fi)

	wg.Add(1)
	go showProgress(wg, pw, length)

	_, err = io.Copy(pw, resp.Body)
	if err != nil {
		return err
	}
	pw.Close()
	wg.Wait()
	return nil
}

func showProgress(wg *sync.WaitGroup, pw *ioutil.ProgressWriter, total int) {
	defer wg.Done()

	var last time.Time
	for sum := range pw.Status() {
		if time.Since(last) < 2*time.Second && sum < total {
			continue
		}
		if total > 0 {
			global.LOG.Infof("%3d MiB of %3d MiB downloaded (%d%%)", sum/1024/1024, total/1024/1024, sum*100/total)
		} else {
			global.LOG.Infof("%3d MiB downloaded", sum/1024/1024)
		}
		last = time.Now()
	}
}
