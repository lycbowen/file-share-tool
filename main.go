package main

import (
	"encoding/json"
	"errors"
	"file-share-tool/frontend"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

type File struct {
	FileName    string `json:"file_name"`
	FileModtime string `json:"file_modtime"`
	IsDir       bool   `json:"is_dir"`
	FileSize    string `json:"file_size"`
	SubFileNum  int    `json:"sub_file_num"`
	SubDirNum   int    `json:"sub_dir_num"`
}

type Resp struct {
	Message    string `json:"message"`
	LocalIP    string `json:"local_ip"`
	TargetPath string `json:"path"`
	FileList   []File `json:"files"`
}

var IgnoreList = []string{
	".DS_Store",
	".Trash",
	".localized",
}

const (
	defaultPort = 8000
	dateFormat  = "2006/01/02 15:04:05"
)

func GetInternalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", errors.New("internal IP fetch failed, detail:" + err.Error())
	}
	defer conn.Close()
	ip := strings.Split(conn.LocalAddr().String(), ":")[0]
	return ip, nil
}

func humanReadableSize(size int64) string {
	const (
		KB = 1 << 10
		MB = 1 << 20
		GB = 1 << 30
		TB = 1 << 40
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/TB)
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func CountDirsAndFiles(path string) (dirs, files int, err error) {
	fs, err := os.ReadDir(path)
	if err != nil {
		return 0, 0, err
	}
	for _, v := range fs {
		info, err := v.Info()
		if err != nil {
			log.Printf("[WARN] Skipped unreadable entry: %v", err)
			continue
		}
		if slices.Contains(IgnoreList, info.Name()) {
			continue
		}
		if info.IsDir() {
			dirs++
		} else {
			files++
		}
	}
	return
}

func buildFileList(targetDir string) ([]File, error) {
	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return nil, err
	}

	var result []File
	var dirs []File
	var files []File

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			log.Printf("[WARN] Cannot read file info: %v", err)
			continue
		}
		if slices.Contains(IgnoreList, info.Name()) {
			continue
		}
		file := File{
			FileName:    info.Name(),
			FileModtime: info.ModTime().Local().Format(dateFormat),
			IsDir:       info.IsDir(),
		}

		if info.IsDir() {
			dNum, fNum, _ := CountDirsAndFiles(filepath.Join(targetDir, info.Name()))
			file.SubDirNum = dNum
			file.SubFileNum = fNum
			dirs = append(dirs, file)
		} else {
			file.FileSize = humanReadableSize(info.Size())
			files = append(files, file)
		}
	}

	result = append(result, dirs...)
	result = append(result, files...)
	return result, nil
}

// 路径校验：确保 target 是 base 的子路径
func isSubPath(base, target string) bool {
	baseAbs, err1 := filepath.Abs(base)
	targetAbs, err2 := filepath.Abs(target)
	if err1 != nil || err2 != nil {
		return false
	}
	rel, err := filepath.Rel(baseAbs, targetAbs)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}

func getFileListHandler(rootDir, localIP string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rawPath := r.URL.Query().Get("path")
		if rawPath == "" || rawPath == "undefined" {
			rawPath = rootDir
		}
		rawPath = strings.TrimPrefix(rawPath, "/")
		cleanPath := filepath.Clean(rawPath)
		absPath, err := filepath.Abs(cleanPath)
		if err != nil || !isSubPath(rootDir, absPath) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		absPath = strings.ReplaceAll(absPath, "\\", "/")
		resp := Resp{
			LocalIP:    fmt.Sprintf("%s:%d", localIP, defaultPort),
			TargetPath: absPath,
		}

		fileList, err := buildFileList(absPath)
		if err != nil {
			resp.Message = err.Error()
		} else {
			resp.FileList = fileList
		}

		w.Header().Set("Content-Type", "application/json")
		log.Println("[INFO] Request dir:", absPath)
		if err := json.NewEncoder(w).Encode(&resp); err != nil {
			log.Printf("[ERROR] Encoding response: %v", err)
		}
	}
}

func downloadHandler(rootDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetFile := r.URL.Query().Get("fname")
		if targetFile == "" {
			http.Error(w, "Missing file name", http.StatusBadRequest)
			return
		}
		targetFile = strings.TrimPrefix(targetFile, "/")
		absPath, err := filepath.Abs(filepath.Clean(targetFile))
		if err != nil || !isSubPath(rootDir, absPath) {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		file, err := os.Open(absPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error opening file: %s", err), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(absPath)))
		w.Header().Set("Content-Type", "application/octet-stream")

		if _, err := io.Copy(w, file); err != nil {
			http.Error(w, fmt.Sprintf("Error sending file: %s", err), http.StatusInternalServerError)
			return
		}
		log.Println("[INFO] File downloaded:", absPath)
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	case "linux": // Linux
		cmd = exec.Command("xdg-open", url)
	case "windows": // Windows
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		fmt.Println("Unsupported platform")
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[ERROR] Failed to open browser: %v", err)
	}
}

func main() {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal("[FATAL] Failed to get current user:", err)
	}
	homeDir := currentUser.HomeDir

	defaultDir, err := os.Getwd()
	if err != nil {
		log.Fatal("[FATAL] Failed to get working directory:", err)
	}

	var targetDir string
	flag.StringVar(&targetDir, "t", defaultDir, "Directory to share (default: current dir, use 'home' for home directory)")
	flag.Parse()

	if targetDir == "home" {
		targetDir = homeDir
	}

	localIP, err := GetInternalIP()
	if err != nil {
		log.Printf("[WARN] Failed to get local IP: %v", err)
		localIP = "localhost"
	} else {
		log.Println("[INFO] Local IP:", localIP)
	}

	http.HandleFunc("/api/files", getFileListHandler(targetDir, localIP))
	http.HandleFunc("/api/download", downloadHandler(targetDir))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fs, err := frontend.FS()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.FileServer(fs).ServeHTTP(w, r)
	})

	log.Printf("[INFO] Server is running at http://localhost:%d", defaultPort)
	log.Printf("[INFO] LAN access: http://%s:%d", localIP, defaultPort)

	go openBrowser(fmt.Sprintf("http://%s:%d", localIP, defaultPort))

	if err := http.ListenAndServe(fmt.Sprintf(":%d", defaultPort), nil); err != nil {
		log.Fatalf("[FATAL] Server error: %v", err)
	}
}
