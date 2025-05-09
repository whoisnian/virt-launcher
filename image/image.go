package image

import (
	"cmp"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/whoisnian/glb/logger"
	"github.com/whoisnian/glb/util/osutil"
	"github.com/whoisnian/virt-launcher/cache"
	"github.com/whoisnian/virt-launcher/global"
)

//go:embed index/*.json
var dataFS embed.FS

type Distro struct {
	Name     string // short id from `osinfo-query os`
	Version  string
	Upstream string
	Images   []Image
}

type Image struct {
	Arch    string // current supported: amd64 arm64
	Account string
	FileUrl string
	HashUrl string
	HashFmt string // example: sha512sum:raw sha256sum:gnu sha256sum:bsd
	HashVal string // hex string
}

func (distro *Distro) WriteJsonFile(name string) error {
	fi, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("os.Create: %w", err)
	}
	defer fi.Close()

	enc := json.NewEncoder(fi)
	enc.SetIndent("", "  ")
	if err = enc.Encode(distro); err != nil {
		return fmt.Errorf("json.Encode: %w", err)
	}
	return nil
}

func (img *Image) BaseName() string {
	return path.Base(img.FileUrl)
}

var distroMap = make(map[string]*Distro)

func Setup(ctx context.Context) {
	subFS, err := fs.Sub(dataFS, "index")
	if err != nil {
		global.LOG.Fatal(ctx, "fs.Sub", logger.Error(err))
	}

	content, err := os.ReadFile(cache.Index.Join("version"))
	if err != nil || strings.TrimSpace(string(content)) != global.Version {
		global.LOG.Debugf(ctx, "index version missing or mismatched, reset index cache")
		if err = cache.Index.Reset(); err != nil {
			global.LOG.Fatal(ctx, "cache.Index.Reset", logger.Error(err))
		}
		if err = loadDistroMapFromFS(ctx, subFS); err != nil {
			global.LOG.Fatal(ctx, "image.loadDistroMapFromFS", logger.Error(err))
		}
		if err = saveDistroMapToDir(ctx, cache.Index.FullPath()); err != nil {
			global.LOG.Fatal(ctx, "image.saveDistroMapToDir", logger.Error(err))
		}
	} else {
		if err = loadDistroMapFromFS(ctx, os.DirFS(cache.Index.FullPath())); err != nil {
			global.LOG.Fatal(ctx, "image.loadDistroMapFromFS", logger.Error(err))
		}
	}
}

func loadDistroMapFromFS(ctx context.Context, f fs.FS) error {
	files, err := fs.ReadDir(f, ".")
	if err != nil {
		return fmt.Errorf("fs.ReadDir: %w", err)
	}
	global.LOG.Debugf(ctx, "found %d distro files", len(files))
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		global.LOG.Debugf(ctx, "read and parse '%s'...", file.Name())
		content, err := fs.ReadFile(f, file.Name())
		if err != nil {
			return fmt.Errorf("fs.ReadFile: %w", err)
		}

		distro := &Distro{}
		if err = json.Unmarshal(content, distro); err != nil {
			return fmt.Errorf("json.Unmarshal: %w", err)
		}
		if _, ok := distroMap[distro.Name]; ok {
			return errors.New("duplicated distro file for " + distro.Name)
		}
		distroMap[distro.Name] = distro
	}
	return nil
}

func saveDistroMapToDir(ctx context.Context, dir string) error {
	for _, distro := range distroMap {
		filePath := filepath.Join(dir, distro.Name+".json")
		global.LOG.Debugf(ctx, "save distro file to %s", filePath)
		if err := distro.WriteJsonFile(filePath); err != nil {
			return fmt.Errorf("distro.WriteJsonFile: %w", err)
		}
	}
	filePath := filepath.Join(dir, "version")
	if err := os.WriteFile(filePath, []byte(global.Version), osutil.DefaultFileMode); err != nil {
		return fmt.Errorf("os.WriteFile: %w", err)
	}
	return nil
}

func LookupImage(os string, arch string) (*Image, error) {
	distro, ok := distroMap[os]
	if !ok {
		return nil, errors.New("os not found")
	}
	for i := range distro.Images {
		if distro.Images[i].Arch == arch {
			return &distro.Images[i], nil
		}
	}
	return nil, errors.New("arch not found")
}

func UpdateIndex(ctx context.Context) error {
	for _, distro := range distroMap {
		if ok, err := distro.CheckAndUpdate(ctx); err != nil {
			return fmt.Errorf("distro.CheckAndUpdate(%s): %w", distro.Name, err)
		} else if ok {
			filePath := cache.Index.Join(distro.Name + ".json")
			global.LOG.Debugf(ctx, "update distro file at %s", filePath)
			if err = distro.WriteJsonFile(filePath); err != nil {
				return fmt.Errorf("distro.WriteJsonFile(%s): %w", distro.Name, err)
			}
		}
	}
	return nil
}

func ListAll(w io.Writer) {
	nameLen, archLen, versionLen := 4, 4, 7
	list := [][]string{}
	for _, distro := range distroMap {
		nameLen = max(nameLen, len(distro.Name))
		versionLen = max(versionLen, len(distro.Version))
		for _, img := range distro.Images {
			archLen = max(archLen, len(img.Arch))
			list = append(list, []string{distro.Name, img.Arch, distro.Version})
		}
	}
	slices.SortFunc(list, func(a, b []string) int {
		if a[0] == b[0] {
			return cmp.Compare(a[1], b[1])
		}
		return cmp.Compare(a[0], b[0])
	})
	fmt.Fprintf(w, "| %-*s | %-*s | %-*s |\n", nameLen, "name", archLen, "arch", versionLen, "version")
	fmt.Fprintf(w, "| %s | %s | %s |\n", strings.Repeat("-", nameLen), strings.Repeat("-", archLen), strings.Repeat("-", versionLen))
	for _, item := range list {
		fmt.Fprintf(w, "| %-*s | %-*s | %-*s |\n", nameLen, item[0], archLen, item[1], versionLen, item[2])
	}
}
