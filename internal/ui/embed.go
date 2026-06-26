package ui

import (
	"embed"
	"io/fs"
)

//go:embed all:build
var BuildFS embed.FS

func GetFS() (fs.FS, error) {
    sub, err := fs.Sub(BuildFS, "build")
    if err != nil {
        return nil, err
    }
    return sub, nil
}