package data

import (
	"embed"
)

//go:embed os/*.json
var FS embed.FS
