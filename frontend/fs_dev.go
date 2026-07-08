//go:build dev_frontend

package frontend

import (
	"net/http"
	"os"
)

func FS() http.FileSystem {
	if _, err := os.Stat("frontend/dist/index.html"); err == nil {
		return http.Dir("frontend/dist")
	}
	return http.Dir("frontend/public")
}
