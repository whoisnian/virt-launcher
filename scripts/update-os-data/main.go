package main

import (
	"bufio"
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
	"strconv"
	"strings"

	"github.com/whoisnian/glb/ansi"
	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/virt-launcher/data"
	"github.com/whoisnian/virt-launcher/image"
)

var regMap = map[string]*regexp.Regexp{
	"archlinux":      regexp.MustCompile(`href="v(\d+.\d+)/"`),
	"centos7.0":      regexp.MustCompile(`href="CentOS-7-x86_64-GenericCloud-(\d+).qcow2"`),
	"centos-stream8": regexp.MustCompile(`href="CentOS-Stream-GenericCloud-8-(\d+.\d+).x86_64.qcow2"`),
	"centos-stream9": regexp.MustCompile(`href="CentOS-Stream-GenericCloud-9-(\d+.\d+).x86_64.qcow2"`),
	"debian10":       regexp.MustCompile(`href="(\d+-\d+)/"`),
	"debian11":       regexp.MustCompile(`href="(\d+-\d+)/"`),
	"debian12":       regexp.MustCompile(`href="(\d+-\d+)/"`),
	"ubuntu18.04":    regexp.MustCompile(`href="(\d+)/"`),
	"ubuntu20.04":    regexp.MustCompile(`href="(\d+)/"`),
	"ubuntu22.04":    regexp.MustCompile(`href="(\d+)/"`),
}

var LOG = logger.New(logger.NewNanoHandler(os.Stderr, logger.NewOptions(
	logger.LevelInfo, ansi.IsSupported(os.Stderr.Fd()), false,
)))

func main() {
	files, err := data.FS.ReadDir(data.OsDir)
	if err != nil {
		LOG.Fatal(err.Error())
	}
	LOG.Info("Found " + strconv.Itoa(len(files)) + " os files")

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		LOG.Info("Check for updates '" + file.Name() + "'...")
		content, err := data.FS.ReadFile(filepath.Join(data.OsDir, file.Name()))
		if err != nil {
			LOG.Fatal(err.Error())
		}

		o := &image.Os{}
		err = json.Unmarshal(content, o)
		if err != nil {
			LOG.Fatal(err.Error())
		}

		if updateOsData(o) {
			LOG.Info("Update os file: " + file.Name())
			fi, err := os.Create(filepath.Join("data", data.OsDir, file.Name()))
			if err != nil {
				LOG.Fatal(err.Error())
			}
			defer fi.Close()

			enc := json.NewEncoder(fi)
			enc.SetIndent("", "  ")
			if err = enc.Encode(o); err != nil {
				LOG.Fatal(err.Error())
			}
		}
	}
}

func updateOsData(o *image.Os) bool {
	reg, ok := regMap[o.Name]
	if !ok {
		LOG.Fatal("Unknown os name")
	}

	latestV := fetchLatestVersion(o.Upstream, reg)
	if latestV <= o.Version {
		return false
	}
	LOG.Info("Found newer version: " + o.Version + " => " + latestV)

	for i := range o.Images {
		o.Images[i].Url = strings.ReplaceAll(o.Images[i].Url, o.Version, latestV)
		if strings.HasPrefix(o.Images[i].Hash, "https://") || strings.HasPrefix(o.Images[i].Hash, "http://") {
			o.Images[i].Hash = strings.ReplaceAll(o.Images[i].Hash, o.Version, latestV)
		} else if strings.HasPrefix(o.Images[i].Hash, "sha256sum:") {
			o.Images[i].Hash = "sha256sum:" + remoteHashFrom(o.Images[i].Url, sha256.New())
		} else if strings.HasPrefix(o.Images[i].Hash, "sha512sum:") {
			o.Images[i].Hash = "sha512sum:" + remoteHashFrom(o.Images[i].Url, sha512.New())
		}
	}
	o.Version = latestV
	return true
}

func remoteHashFrom(url string, hasher hash.Hash) string {
	resp, err := http.Get(url)
	if err != nil {
		LOG.Fatal(err.Error())
	}
	defer resp.Body.Close()

	_, err = io.CopyBuffer(hasher, resp.Body, make([]byte, 4*1024*1024))
	if err != nil {
		LOG.Fatal(err.Error())
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func fetchLatestVersion(upstream string, reg *regexp.Regexp) (version string) {
	resp, err := http.Get(upstream)
	if err != nil {
		LOG.Fatal(err.Error())
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
