package main

import (
	"embed"
	"strings"
)

var templateLog = fileLogger("template")

//go:embed web/*
var webFS embed.FS

var indexHTML string

var webPartPaths = []string{
	"web/head.html",
	"web/styles.css",
	"web/body.html",
	"web/app.js",
	"web/foot.html",
}

func init() {
	templateLog.Debug("building embedded index template", "parts", len(webPartPaths))
	var b strings.Builder
	for _, name := range webPartPaths {
		templateLog.Debug("reading embedded web asset", "path", name)
		data, err := webFS.ReadFile(name)
		if err != nil {
			templateLog.Error("failed reading embedded web asset", "path", name, "error", err)
			panic("embed " + name + ": " + err.Error())
		}
		b.Write(data)
	}
	indexHTML = b.String()
	templateLog.Debug("built index html payload", "bytes", len(indexHTML))
}
