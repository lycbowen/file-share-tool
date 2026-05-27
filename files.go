package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const dateFormat = "2006/01/02 15:04:05"

var ignoreList = map[string]struct{}{
	".DS_Store":   {},
	".Trash":      {},
	".localized":  {},
	"Thumbs.db":   {},
	"desktop.ini": {},
}

type File struct {
	FileName    string `json:"file_name"`
	FileModtime string `json:"file_modtime"`
	IsDir       bool   `json:"is_dir"`
	FileSize    string `json:"file_size"`
	SubFileNum  int    `json:"sub_file_num"`
	SubDirNum   int    `json:"sub_dir_num"`
	Path        string `json:"path"`
}

type ShareFS struct {
	root string
}

func NewShareFS(rootDir string) (*ShareFS, error) {
	abs, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", rootDir)
	}
	realRoot, err := filepath.EvalSymlinks(abs)
	if err != nil {
		log.Printf("[WARN] Cannot evaluate shared root symlinks, using absolute path: %v", err)
		realRoot = abs
	}
	return &ShareFS{root: filepath.Clean(realRoot)}, nil
}

func (s *ShareFS) Root() string {
	return s.root
}

func (s *ShareFS) Resolve(raw string) (string, string, error) {
	if abs, rel, ok := s.resolveAbsoluteCandidate(raw); ok {
		return abs, rel, nil
	}
	rel, err := cleanRelativePath(raw)
	if err != nil {
		return "", "", err
	}
	if rel == "" {
		return s.root, "", nil
	}
	abs := s.root
	abs = filepath.Join(s.root, filepath.FromSlash(rel))
	realPath, err := filepath.EvalSymlinks(abs)
	if err != nil {
		hasSymlink, linkErr := hasSymlinkComponent(s.root, rel)
		if linkErr != nil {
			return "", "", linkErr
		}
		if hasSymlink {
			return "", "", err
		}
		realPath = filepath.Clean(abs)
	} else {
		realPath = filepath.Clean(realPath)
	}
	if !isWithin(s.root, realPath) {
		return "", "", errors.New("access denied")
	}
	return realPath, rel, nil
}

func (s *ShareFS) resolveAbsoluteCandidate(raw string) (string, string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "/" || raw == "." || raw == "undefined" {
		return "", "", false
	}
	candidate := filepath.Clean(raw)
	if !filepath.IsAbs(candidate) {
		return "", "", false
	}
	realPath, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", "", false
	}
	realPath = filepath.Clean(realPath)
	if !isWithin(s.root, realPath) {
		return "", "", false
	}
	rel, err := filepath.Rel(s.root, realPath)
	if err != nil || rel == "." {
		return realPath, "", true
	}
	return realPath, filepath.ToSlash(rel), true
}

func (s *ShareFS) List(raw string) ([]File, string, error) {
	abs, rel, err := s.Resolve(raw)
	if err != nil {
		return nil, "", err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, "", err
	}
	if !info.IsDir() {
		return nil, "", errors.New("path is not a directory")
	}

	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, "", err
	}

	dirs := make([]File, 0)
	files := make([]File, 0)
	for _, entry := range entries {
		if isIgnored(entry.Name()) {
			continue
		}
		childRel := joinRelative(rel, entry.Name())
		entryPath, _, err := s.Resolve(childRel)
		if err != nil {
			log.Printf("[WARN] Skipped unsafe entry %s: %v", childRel, err)
			continue
		}
		info, err := os.Stat(entryPath)
		if err != nil {
			log.Printf("[WARN] Cannot read file info: %v", err)
			continue
		}

		item := File{
			FileName:    info.Name(),
			FileModtime: info.ModTime().Local().Format(dateFormat),
			IsDir:       info.IsDir(),
			Path:        childRel,
		}
		if info.IsDir() {
			item.SubDirNum, item.SubFileNum, _ = countDirsAndFiles(entryPath)
			dirs = append(dirs, item)
		} else {
			item.FileSize = humanReadableSize(info.Size())
			files = append(files, item)
		}
	}

	sort.SliceStable(dirs, func(i, j int) bool { return strings.ToLower(dirs[i].FileName) < strings.ToLower(dirs[j].FileName) })
	sort.SliceStable(files, func(i, j int) bool { return strings.ToLower(files[i].FileName) < strings.ToLower(files[j].FileName) })
	return append(dirs, files...), rel, nil
}

func cleanRelativePath(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "/" || raw == "." || raw == "undefined" {
		return "", nil
	}
	raw = strings.ReplaceAll(raw, "\\", "/")
	raw = strings.TrimPrefix(raw, "/")
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(raw)))
	if clean == "." {
		return "", nil
	}
	if filepath.IsAbs(raw) || strings.HasPrefix(clean, "../") || clean == ".." || strings.Contains(clean, ":") {
		return "", errors.New("access denied")
	}
	return clean, nil
}

func isWithin(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func hasSymlinkComponent(root, rel string) (bool, error) {
	current := root
	for _, part := range strings.Split(filepath.ToSlash(rel), "/") {
		if part == "" {
			continue
		}
		current = filepath.Join(current, filepath.FromSlash(part))
		info, err := os.Lstat(current)
		if err != nil {
			return false, err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return true, nil
		}
	}
	return false, nil
}

func joinRelative(base, name string) string {
	if base == "" {
		return filepath.ToSlash(name)
	}
	return filepath.ToSlash(filepath.Join(filepath.FromSlash(base), name))
}

func isIgnored(name string) bool {
	_, ok := ignoreList[name]
	return ok
}

func countDirsAndFiles(path string) (dirs, files int, err error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, 0, err
	}
	for _, entry := range entries {
		if isIgnored(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			log.Printf("[WARN] Skipped unreadable entry: %v", err)
			continue
		}
		if info.IsDir() {
			dirs++
		} else {
			files++
		}
	}
	return dirs, files, nil
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
