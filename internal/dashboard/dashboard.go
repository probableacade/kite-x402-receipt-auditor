package dashboard

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/*
var files embed.FS

func Handler() http.Handler {
	root, err := fs.Sub(files, "web")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(root))
}
