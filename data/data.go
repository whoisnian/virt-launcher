package data

import (
	"embed"
)

//go:embed distro/*.json
var FS embed.FS
