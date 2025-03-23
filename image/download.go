package image

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/whoisnian/glb/util/ioutil"
	"github.com/whoisnian/virt-launcher/global"
)

func (img *Image) Download(ctx context.Context, filePath string) error {
	rHash, err := img.RemoteHash(ctx)
	if err != nil {
		return err
	}
	global.LOG.Debugf(ctx, "file remote hash: %s", rHash)

	// check image file exists
	if _, err := os.Stat(filePath); err == nil {
		global.LOG.Info(ctx, "file already exists, start verifying")
		lHash, err := img.LocalHashFrom(ctx, filePath)
		if err != nil {
			return err
		}
		global.LOG.Debugf(ctx, "file local hash: %s", lHash)
		if lHash == rHash {
			global.LOG.Info(ctx, "hash verification ok, skip downloading")
			return nil
		} else {
			global.LOG.Warn(ctx, "hash verification failed, delete local file")
			if err := os.Remove(filePath); err != nil {
				return err
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	// download image file
	if err = img.download(ctx, filePath); err != nil {
		return err
	}

	// verify image file
	lHash, err := img.LocalHashFrom(ctx, filePath)
	if err != nil {
		return err
	}
	global.LOG.Debugf(ctx, "file local hash: %s", lHash)
	if lHash == rHash {
		global.LOG.Info(ctx, "hash verification ok")
		return nil
	} else {
		return errors.New("hash verification failed")
	}
}

func (img *Image) download(ctx context.Context, filePath string) error {
	fi, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer fi.Close()

	global.LOG.Debugf(ctx, "start downloading image from %s", img.Url)
	resp, err := http.Get(img.Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	total, err := strconv.ParseInt(resp.Header.Get("content-length"), 10, 64)
	if err != nil {
		global.LOG.Debugf(ctx, "content-length atoi %v", err)
	} else {
		global.LOG.Debugf(ctx, "got content-length %d", total)
	}

	pw := ioutil.NewProgressWriter(fi)
	go func() {
		defer pw.Close()
		_, err = io.Copy(pw, resp.Body)
	}()

	var last time.Time
	for sum := range pw.Status() {
		if time.Since(last) < 2*time.Second && sum < total {
			continue
		}
		if total > 0 {
			global.LOG.Infof(ctx, "%3d MiB of %3d MiB downloaded (%d%%)", sum/1024/1024, total/1024/1024, sum*100/total)
		} else {
			global.LOG.Infof(ctx, "%3d MiB downloaded", sum/1024/1024)
		}
		last = time.Now()
	}
	return err
}
