package main

import (
	"embed"
	"strings"
)

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
	var b strings.Builder
	for _, name := range webPartPaths {
		data, err := webFS.ReadFile(name)
		if err != nil {
			panic("embed " + name + ": " + err.Error())
		}
		b.Write(data)
	}
	indexHTML = b.String()
}
