package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type AppServer struct {
	share     *ShareFS
	staticFS  http.FileSystem
	publicURL string
}

type FilesResponse struct {
	Message    string `json:"message,omitempty"`
	LocalIP    string `json:"local_ip"`
	PublicURL  string `json:"public_url"`
	TargetPath string `json:"path"`
	FileList   []File `json:"files"`
}

func NewAppServer(share *ShareFS, staticFS http.FileSystem, publicURL string) *AppServer {
	return &AppServer{share: share, staticFS: staticFS, publicURL: publicURL}
}

func (s *AppServer) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/files", s.filesHandler)
	mux.HandleFunc("/api/download", s.downloadHandler)
	mux.Handle("/", http.FileServer(s.staticFS))
	return mux
}

func (s *AppServer) filesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	files, rel, err := s.share.List(r.URL.Query().Get("path"))
	resp := FilesResponse{
		LocalIP:    strings.TrimPrefix(s.publicURL, "http://"),
		PublicURL:  s.publicURL,
		TargetPath: rel,
		FileList:   files,
	}
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, os.ErrNotExist) {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "access denied") {
			status = http.StatusForbidden
		} else if strings.Contains(err.Error(), "not a directory") {
			status = http.StatusBadRequest
		}
		resp.Message = err.Error()
		writeJSON(w, status, resp)
		return
	}

	log.Printf("[INFO] Request dir: %s", rel)
	writeJSON(w, http.StatusOK, resp)
}

func (s *AppServer) downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rawPath := r.URL.Query().Get("path")
	if rawPath == "" {
		rawPath = r.URL.Query().Get("fname")
	}
	if rawPath == "" {
		http.Error(w, "Missing path", http.StatusBadRequest)
		return
	}

	abs, rel, err := s.share.Resolve(rawPath)
	if err != nil {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}
	info, err := os.Stat(abs)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error opening file: %s", err), http.StatusNotFound)
		return
	}
	if info.IsDir() {
		http.Error(w, "Cannot download a directory", http.StatusBadRequest)
		return
	}

	file, err := os.Open(abs)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error opening file: %s", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	name := filepath.Base(abs)
	w.Header().Set("Content-Disposition", contentDisposition(name))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	if _, err := io.Copy(w, file); err != nil {
		log.Printf("[ERROR] Error sending file %s: %v", rel, err)
		return
	}
	log.Printf("[INFO] File downloaded: %s", rel)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("[ERROR] Encoding response: %v", err)
	}
}

func contentDisposition(filename string) string {
	ascii := strings.Map(func(r rune) rune {
		if r < 32 || r == '"' || r == '\\' || r > 126 {
			return -1
		}
		return r
	}, filename)
	if ascii == "" {
		ascii = "download"
	}
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, strings.ReplaceAll(ascii, `"`, ""), url.PathEscape(filename))
}
