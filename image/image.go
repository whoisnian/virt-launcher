package image

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/virt-launcher/data"
	"github.com/whoisnian/virt-launcher/global"
)

var osMap = make(map[string]*Os)

type Os struct {
	Name     string
	Version  string
	Upstream string
	Images   []Image
}

type Image struct {
	Arch    string
	Account string
	Url     string
	Hash    string
}

func (img *Image) BaseName() string {
	return path.Base(img.Url)
}

func LookupImage(os string, arch string) (*Image, error) {
	o, ok := osMap[os]
	if !ok {
		return nil, errors.New("os not found")
	}
	for i := range o.Images {
		if o.Images[i].Arch == arch {
			return &o.Images[i], nil
		}
	}
	return nil, errors.New("arch not found")
}

func ListAll() {
	nameLen, archLen, versionLen := 4, 4, 7
	list := [][]string{}
	for _, o := range osMap {
		for _, img := range o.Images {
			nameLen = max(nameLen, len(o.Name))
			archLen = max(archLen, len(img.Arch))
			versionLen = max(versionLen, len(o.Version))
			list = append(list, []string{o.Name, img.Arch, o.Version})
		}
	}
	slices.SortFunc(list, func(a, b []string) int {
		if a[0] == b[0] {
			return cmp.Compare(a[1], b[1])
		}
		return cmp.Compare(a[0], b[0])
	})
	fmt.Printf("| %-*s | %-*s | %-*s |\n", nameLen, "name", archLen, "arch", versionLen, "version")
	fmt.Printf("| %s | %s | %s |\n", strings.Repeat("-", nameLen), strings.Repeat("-", archLen), strings.Repeat("-", versionLen))
	for _, item := range list {
		fmt.Printf("| %-*s | %-*s | %-*s |\n", nameLen, item[0], archLen, item[1], versionLen, item[2])
	}
}

func Setup(ctx context.Context) {
	files, err := data.FS.ReadDir(data.OsDir)
	if err != nil {
		global.LOG.Fatal(ctx, "data.FS.ReadDir", logger.Error(err))
	}
	global.LOG.Debugf(ctx, "found %d os files", len(files))

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		global.LOG.Debugf(ctx, "read and parse '%s'...", file.Name())
		content, err := data.FS.ReadFile(filepath.Join(data.OsDir, file.Name()))
		if err != nil {
			global.LOG.Fatal(ctx, "data.FS.ReadFile", logger.Error(err))
		}

		o := &Os{}
		err = json.Unmarshal(content, o)
		if err != nil {
			global.LOG.Fatal(ctx, "json.Unmarshal", logger.Error(err))
		}
		if _, ok := osMap[o.Name]; ok {
			global.LOG.Fatalf(ctx, "duplicated os file %s", o.Name)
		}
		osMap[o.Name] = o
	}
}
