package main

import (
	"embed"
	"io/fs"
	"vidvault/src/logging"
)

var templateLog = logging.FileLogger("template")

//go:embed web/dist/*
var webDist embed.FS

var webRoot fs.FS

func init() {
	var err error
	webRoot, err = fs.Sub(webDist, "web/dist")
	if err != nil {
		templateLog.Error("failed to sub web dist fs", "error", err)
		panic("embed web/dist: " + err.Error())
	}
	templateLog.Debug("embedded web dist ready")
}
