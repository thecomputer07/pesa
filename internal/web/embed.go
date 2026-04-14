// Package web exposes the embedded Vite SPA to the Go binary.
//
// The //go:embed directive requires the target to be a subdirectory of this
// source file's directory; the Makefile copies web/dist/ into ./dist/ before
// `go build`. At development time, dist/ may be empty aside from .gitkeep; in
// that case the server falls back gracefully.
package web

import "embed"

//go:embed all:dist
var Dist embed.FS
