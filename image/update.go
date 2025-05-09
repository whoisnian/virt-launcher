package image

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/whoisnian/virt-launcher/global"
)

var versionRegexMap = map[string]*regexp.Regexp{
	"alpinelinux3.21": regexp.MustCompile(`href="generic_alpine-(\d+\.\d+\.\d+)-x86_64-bios-cloudinit-r0\.qcow2"`),
	"archlinux":       regexp.MustCompile(`href="v(\d+\.\d+)/"`),
	"centos7.0":       regexp.MustCompile(`href="CentOS-7-x86_64-GenericCloud-(\d+)\.qcow2"`),
	"centos-stream9":  regexp.MustCompile(`href="CentOS-Stream-GenericCloud-9-(\d+\.\d+)\.x86_64\.qcow2"`),
	"centos-stream10": regexp.MustCompile(`href="CentOS-Stream-GenericCloud-10-(\d+\.\d+)\.x86_64\.qcow2"`),
	"debian11":        regexp.MustCompile(`href="(\d+-\d+)/"`),
	"debian12":        regexp.MustCompile(`href="(\d+-\d+)/"`),
	"fedora41":        regexp.MustCompile(`href="Fedora-Cloud-Base-Generic-(\d+-\d+\.\d+)\.x86_64\.qcow2"`),
	"rocky8":          regexp.MustCompile(`href="Rocky-8-GenericCloud-Base-(\d+\.\d+-\d+\.\d+)\.x86_64\.qcow2"`),
	"rocky9":          regexp.MustCompile(`href="Rocky-9-GenericCloud-Base-(\d+\.\d+-\d+\.\d+)\.x86_64\.qcow2"`),
	"ubuntu18.04":     regexp.MustCompile(`href="(\d+)/"`),
	"ubuntu20.04":     regexp.MustCompile(`href="(\d+)/"`),
	"ubuntu22.04":     regexp.MustCompile(`href="(\d+)/"`),
	"ubuntu24.04":     regexp.MustCompile(`href="(\d+)/"`),
}

func fetchLatestVersion(os string, upstream string) (version string, err error) {
	if regex, ok := versionRegexMap[os]; ok {
		resp, err := http.Get(upstream)
		if err != nil {
			return "", fmt.Errorf("http.Get: %w", err)
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			if matches := regex.FindAllSubmatch(scanner.Bytes(), -1); len(matches) > 0 {
				for _, match := range matches {
					if len(match) > 1 && string(match[1]) > version {
						version = string(match[1])
					}
				}
			}
		}
		return version, nil
	} else {
		return "", errors.New("unknown version regex for " + os)
	}
}

func (distro *Distro) CheckAndUpdate(ctx context.Context) (bool, error) {
	global.LOG.Infof(ctx, "check for updates of %s...", distro.Name)
	latestV, err := fetchLatestVersion(distro.Name, distro.Upstream)
	if err != nil {
		return false, fmt.Errorf("fetchLatestVersion: %w", err)
	}
	if latestV > distro.Version {
		global.LOG.Infof(ctx, "found newer version: %s => %s", distro.Version, latestV)
		for i := range distro.Images {
			distro.Images[i].FileUrl = strings.ReplaceAll(distro.Images[i].FileUrl, distro.Version, latestV)
			distro.Images[i].HashUrl = strings.ReplaceAll(distro.Images[i].HashUrl, distro.Version, latestV)
			if err = distro.Images[i].UpdateHashVal(ctx); err != nil {
				return false, fmt.Errorf("image.UpdateHashVal: %w", err)
			}
		}
		distro.Version = latestV
		return true, nil
	}
	return false, nil
}
