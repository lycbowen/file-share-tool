package frontend

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist
var buildFS embed.FS

func FS() (http.FileSystem, error) {
	ret, err := fs.Sub(buildFS, "dist")
	if err != nil {
		return nil, err
	}
	return http.FS(ret), nil
}
