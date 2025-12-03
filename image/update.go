package image

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/whoisnian/virt-launcher/global"
)

var versionRegexMap = map[string]*regexp.Regexp{
	"alpinelinux3.21 (amd64)": regexp.MustCompile(`href="generic_alpine-(\d+\.\d+\.\d+)-x86_64-bios-cloudinit-r0\.qcow2"`),
	"alpinelinux3.21 (arm64)": regexp.MustCompile(`href="generic_alpine-(\d+\.\d+\.\d+)-aarch64-uefi-cloudinit-r0\.qcow2"`),
	"alpinelinux3.22 (amd64)": regexp.MustCompile(`href="generic_alpine-(\d+\.\d+\.\d+)-x86_64-bios-cloudinit-r0\.qcow2"`),
	"alpinelinux3.22 (arm64)": regexp.MustCompile(`href="generic_alpine-(\d+\.\d+\.\d+)-aarch64-uefi-cloudinit-r0\.qcow2"`),
	"archlinux (amd64)":       regexp.MustCompile(`href="v(\d+\.\d+)/"`),
	"centos7.0 (amd64)":       regexp.MustCompile(`href="CentOS-7-x86_64-GenericCloud-(\d+)\.qcow2"`),
	"centos7.0 (arm64)":       regexp.MustCompile(`href="CentOS-7-aarch64-GenericCloud-(\d+)\.qcow2"`),
	"centos-stream9 (amd64)":  regexp.MustCompile(`href="CentOS-Stream-GenericCloud-9-(\d+\.\d+)\.x86_64\.qcow2"`),
	"centos-stream9 (arm64)":  regexp.MustCompile(`href="CentOS-Stream-GenericCloud-9-(\d+\.\d+)\.aarch64\.qcow2"`),
	"centos-stream10 (amd64)": regexp.MustCompile(`href="CentOS-Stream-GenericCloud-10-(\d+\.\d+)\.x86_64\.qcow2"`),
	"centos-stream10 (arm64)": regexp.MustCompile(`href="CentOS-Stream-GenericCloud-10-(\d+\.\d+)\.aarch64\.qcow2"`),
	"debian11 (amd64)":        regexp.MustCompile(`href="(\d+-\d+)/"`),
	"debian11 (arm64)":        regexp.MustCompile(`href="(\d+-\d+)/"`),
	"debian12 (amd64)":        regexp.MustCompile(`href="(\d+-\d+)/"`),
	"debian12 (arm64)":        regexp.MustCompile(`href="(\d+-\d+)/"`),
	"debian13 (amd64)":        regexp.MustCompile(`href="(\d+-\d+)/"`),
	"debian13 (arm64)":        regexp.MustCompile(`href="(\d+-\d+)/"`),
	"fedora41 (amd64)":        regexp.MustCompile(`href="Fedora-Cloud-Base-Generic-(\d+-\d+\.\d+)\.x86_64\.qcow2"`),
	"fedora41 (arm64)":        regexp.MustCompile(`href="Fedora-Cloud-Base-Generic-(\d+-\d+\.\d+)\.aarch64\.qcow2"`),
	"fedora42 (amd64)":        regexp.MustCompile(`href="Fedora-Cloud-Base-Generic-(\d+-\d+\.\d+)\.x86_64\.qcow2"`),
	"fedora42 (arm64)":        regexp.MustCompile(`href="Fedora-Cloud-Base-Generic-(\d+-\d+\.\d+)\.aarch64\.qcow2"`),
	"fedora43 (amd64)":        regexp.MustCompile(`href="Fedora-Cloud-Base-Generic-(\d+-\d+\.\d+)\.x86_64\.qcow2"`),
	"fedora43 (arm64)":        regexp.MustCompile(`href="Fedora-Cloud-Base-Generic-(\d+-\d+\.\d+)\.aarch64\.qcow2"`),
	"rocky8 (amd64)":          regexp.MustCompile(`href="Rocky-8-GenericCloud-Base-(\d+\.\d+-\d+\.\d+)\.x86_64\.qcow2"`),
	"rocky8 (arm64)":          regexp.MustCompile(`href="Rocky-8-GenericCloud-Base-(\d+\.\d+-\d+\.\d+)\.aarch64\.qcow2"`),
	"rocky9 (amd64)":          regexp.MustCompile(`href="Rocky-9-GenericCloud-Base-(\d+\.\d+-\d+\.\d+)\.x86_64\.qcow2"`),
	"rocky9 (arm64)":          regexp.MustCompile(`href="Rocky-9-GenericCloud-Base-(\d+\.\d+-\d+\.\d+)\.aarch64\.qcow2"`),
	"ubuntu18.04 (amd64)":     regexp.MustCompile(`href="(\d+)/"`),
	"ubuntu18.04 (arm64)":     regexp.MustCompile(`href="(\d+)/"`),
	"ubuntu20.04 (amd64)":     regexp.MustCompile(`href="(\d+)/"`),
	"ubuntu20.04 (arm64)":     regexp.MustCompile(`href="(\d+)/"`),
	"ubuntu22.04 (amd64)":     regexp.MustCompile(`href="(\d+)/"`),
	"ubuntu22.04 (arm64)":     regexp.MustCompile(`href="(\d+)/"`),
	"ubuntu24.04 (amd64)":     regexp.MustCompile(`href="(\d+)/"`),
	"ubuntu24.04 (arm64)":     regexp.MustCompile(`href="(\d+)/"`),
}

func fetchLatestVersion(ctx context.Context, os string, arch string, source string) (version string, err error) {
	if regex, ok := versionRegexMap[os+" ("+arch+")"]; ok {
		resp, err := requestGet(ctx, source)
		if err != nil {
			return "", fmt.Errorf("requestGet: %w", err)
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
	updated := false
	for i := range distro.Images {
		global.LOG.Infof(ctx, "check for updates of %s (%s)...", distro.Name, distro.Images[i].Arch)
		latestV, err := fetchLatestVersion(ctx, distro.Name, distro.Images[i].Arch, distro.Images[i].Source)
		if err != nil {
			return false, fmt.Errorf("fetchLatestVersion: %w", err)
		}
		if latestV > distro.Images[i].Version {
			global.LOG.Infof(ctx, "found newer version: %s => %s", distro.Images[i].Version, latestV)
			distro.Images[i].FileUrl = strings.ReplaceAll(distro.Images[i].FileUrl, distro.Images[i].Version, latestV)
			distro.Images[i].HashUrl = strings.ReplaceAll(distro.Images[i].HashUrl, distro.Images[i].Version, latestV)
			if err = distro.Images[i].UpdateHashVal(ctx); err != nil {
				return false, fmt.Errorf("image.UpdateHashVal: %w", err)
			}
			distro.Images[i].Version = latestV
			updated = true
		}
	}
	return updated, nil
}
