package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/whoisnian/glb/ansi"
	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/virt-launcher/data"
	"github.com/whoisnian/virt-launcher/image"
)

var regMap = map[string]*regexp.Regexp{
	"archlinux":       regexp.MustCompile(`href="v(\d+.\d+)/"`),
	"centos7.0":       regexp.MustCompile(`href="CentOS-7-x86_64-GenericCloud-(\d+).qcow2"`),
	"centos-stream9":  regexp.MustCompile(`href="CentOS-Stream-GenericCloud-9-(\d+.\d+).x86_64.qcow2"`),
	"centos-stream10": regexp.MustCompile(`href="CentOS-Stream-GenericCloud-10-(\d+.\d+).x86_64.qcow2"`),
	"debian11":        regexp.MustCompile(`href="(\d+-\d+)/"`),
	"debian12":        regexp.MustCompile(`href="(\d+-\d+)/"`),
	"fedora41":        regexp.MustCompile(`href="Fedora-Cloud-Base-Generic-41-(\d+.\d+).x86_64.qcow2"`),
	"rocky8":          regexp.MustCompile(`href="Rocky-8-GenericCloud-Base-(\d+.\d+-\d+.\d+).x86_64.qcow2"`),
	"rocky9":          regexp.MustCompile(`href="Rocky-9-GenericCloud-Base-(\d+.\d+-\d+.\d+).x86_64.qcow2"`),
	"ubuntu18.04":     regexp.MustCompile(`href="(\d+)/"`),
	"ubuntu20.04":     regexp.MustCompile(`href="(\d+)/"`),
	"ubuntu22.04":     regexp.MustCompile(`href="(\d+)/"`),
	"ubuntu24.04":     regexp.MustCompile(`href="(\d+)/"`),
}

var LOG = logger.New(logger.NewNanoHandler(os.Stderr, logger.Options{
	Level:     logger.LevelInfo,
	Colorful:  ansi.IsSupported(os.Stderr.Fd()),
	AddSource: false,
}))

func main() {
	ctx := context.Background()
	files, err := data.FS.ReadDir(data.OsDir)
	if err != nil {
		LOG.Fatal(ctx, "data.FS.ReadDir", logger.Error(err))
	}
	LOG.Infof(ctx, "found %d os files", len(files))

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		LOG.Infof(ctx, "check for updates '%s'...", file.Name())
		content, err := data.FS.ReadFile(filepath.Join(data.OsDir, file.Name()))
		if err != nil {
			LOG.Fatal(ctx, "data.FS.ReadFile", logger.Error(err))
		}

		o := &image.Os{}
		err = json.Unmarshal(content, o)
		if err != nil {
			LOG.Fatal(ctx, "json.Unmarshal", logger.Error(err))
		}

		if updateOsData(ctx, o) {
			LOG.Infof(ctx, "update os file: %s", file.Name())
			fi, err := os.Create(filepath.Join("data", data.OsDir, file.Name()))
			if err != nil {
				LOG.Fatal(ctx, "os.Create", logger.Error(err))
			}
			defer fi.Close()

			enc := json.NewEncoder(fi)
			enc.SetIndent("", "  ")
			if err = enc.Encode(o); err != nil {
				LOG.Fatal(ctx, "json.Encoder.Encode", logger.Error(err))
			}
		}
	}
}

func updateOsData(ctx context.Context, o *image.Os) bool {
	reg, ok := regMap[o.Name]
	if !ok {
		LOG.Fatal(ctx, "unknown os name")
	}

	latestV := fetchLatestVersion(ctx, o.Upstream, reg)
	if latestV <= o.Version {
		return false
	}
	LOG.Infof(ctx, "found newer version: %s => %s", o.Version, latestV)

	for i := range o.Images {
		o.Images[i].Url = strings.ReplaceAll(o.Images[i].Url, o.Version, latestV)
		if strings.HasPrefix(o.Images[i].Hash, "https://") || strings.HasPrefix(o.Images[i].Hash, "http://") {
			o.Images[i].Hash = strings.ReplaceAll(o.Images[i].Hash, o.Version, latestV)
		} else if strings.HasPrefix(o.Images[i].Hash, "sha256sum:") {
			o.Images[i].Hash = "sha256sum:" + remoteHashFrom(ctx, o.Images[i].Url, sha256.New())
		} else if strings.HasPrefix(o.Images[i].Hash, "sha512sum:") {
			o.Images[i].Hash = "sha512sum:" + remoteHashFrom(ctx, o.Images[i].Url, sha512.New())
		}
	}
	o.Version = latestV
	return true
}

func remoteHashFrom(ctx context.Context, url string, hasher hash.Hash) string {
	resp, err := http.Get(url)
	if err != nil {
		LOG.Fatal(ctx, "http.Get", logger.Error(err))
	}
	defer resp.Body.Close()

	_, err = io.CopyBuffer(hasher, resp.Body, make([]byte, 4*1024*1024))
	if err != nil {
		LOG.Fatal(ctx, "io.CopyBuffer", logger.Error(err))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func fetchLatestVersion(ctx context.Context, upstream string, reg *regexp.Regexp) (version string) {
	resp, err := http.Get(upstream)
	if err != nil {
		LOG.Fatal(ctx, "http.Get", logger.Error(err))
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		if matches := reg.FindAllSubmatch(scanner.Bytes(), -1); len(matches) > 0 {
			for _, match := range matches {
				if len(match) > 1 && string(match[1]) > version {
					version = string(match[1])
				}
			}
		}
	}
	return
}
