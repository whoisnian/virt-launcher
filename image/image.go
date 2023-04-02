package image

import (
	"encoding/json"
	"fmt"
	"path"
	"sort"

	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/virt-launcher/data"
)

var distroMap = make(map[string]*Distro)

type Distro struct {
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

func (d *Distro) LookupByArch(arch string) *Image {
	for i := range d.Images {
		if d.Images[i].Arch == arch {
			return &d.Images[i]
		}
	}
	return nil
}

func LookupDistro(name string) *Distro {
	return distroMap[name]
}

func ListAll() {
	nameLen, archLen, versionLen := 6, 4, 7
	list := [][]string{{"distro", "arch", "version"}, {"------", "----", "-------"}}
	for _, dt := range distroMap {
		for _, img := range dt.Images {
			if len(dt.Name) > nameLen {
				nameLen = len(dt.Name)
			}
			if len(img.Arch) > archLen {
				archLen = len(img.Arch)
			}
			if len(dt.Version) > versionLen {
				versionLen = len(dt.Version)
			}
			list = append(list, []string{dt.Name, img.Arch, dt.Version})
		}
	}
	sort.Slice(list[2:], func(i, j int) bool {
		if list[i+2][0] == list[j+2][0] {
			return list[i+2][1] < list[j+2][1]
		}
		return list[i+2][0] < list[j+2][0]
	})
	for _, it := range list {
		fmt.Printf("| %-*s | %-*s | %-*s |\n", nameLen, it[0], archLen, it[1], versionLen, it[2])
	}
}

func Init() {
	dirName := "distro"
	files, err := data.FS.ReadDir(dirName)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Debug("Found ", len(files), " distro files")

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		logger.Debug("Read and parse '", file.Name(), "'...")
		content, err := data.FS.ReadFile(path.Join(dirName, file.Name()))
		if err != nil {
			logger.Fatal(err)
		}

		distro := &Distro{}
		err = json.Unmarshal(content, distro)
		if err != nil {
			logger.Fatal(err)
		}
		if _, ok := distroMap[distro.Name]; ok {
			logger.Fatal("Duplicated distro ", distro.Name)
		}
		distroMap[distro.Name] = distro
	}
}
