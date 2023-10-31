package image

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strconv"

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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func ListAll() {
	nameLen, archLen, versionLen := 4, 4, 7
	list := [][]string{{"name", "arch", "version"}, {"----", "----", "-------"}}
	for _, o := range osMap {
		for _, img := range o.Images {
			nameLen = max(nameLen, len(o.Name))
			archLen = max(archLen, len(img.Arch))
			versionLen = max(versionLen, len(o.Version))
			list = append(list, []string{o.Name, img.Arch, o.Version})
		}
	}
	sort.Slice(list[2:], func(i, j int) bool {
		if list[i+2][0] == list[j+2][0] {
			return list[i+2][1] < list[j+2][1]
		}
		return list[i+2][0] < list[j+2][0]
	})
	for _, item := range list {
		fmt.Printf("| %-*s | %-*s | %-*s |\n", nameLen, item[0], archLen, item[1], versionLen, item[2])
	}
}

func Setup() {
	files, err := data.FS.ReadDir(data.OsDir)
	if err != nil {
		global.LOG.Fatal(err.Error())
	}
	global.LOG.Debug("Found " + strconv.Itoa(len(files)) + " os files")

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		global.LOG.Debug("Read and parse '" + file.Name() + "'...")
		content, err := data.FS.ReadFile(filepath.Join(data.OsDir, file.Name()))
		if err != nil {
			global.LOG.Fatal(err.Error())
		}

		o := &Os{}
		err = json.Unmarshal(content, o)
		if err != nil {
			global.LOG.Fatal(err.Error())
		}
		if _, ok := osMap[o.Name]; ok {
			global.LOG.Fatal("Duplicated os " + o.Name)
		}
		osMap[o.Name] = o
	}
}
