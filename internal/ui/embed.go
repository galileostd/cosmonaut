package ui

import (
	"embed"
	"io/fs"
)

//go:embed all:build/*
var BuildFS embed.FS

func GetFS() (fs.FS, error) {
	return fs.Sub(BuildFS, "build")
}