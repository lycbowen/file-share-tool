//go:build embed_frontend

package frontend

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist
var buildFS embed.FS

func FS() http.FileSystem {
	ret, err := fs.Sub(buildFS, "dist")
	if err != nil {
		return http.FS(buildFS)
	}
	return http.FS(ret)
}
