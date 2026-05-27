package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCleanRelativePath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "root", input: "/", want: ""},
		{name: "empty", input: "", want: ""},
		{name: "normal", input: "docs/report.txt", want: "docs/report.txt"},
		{name: "leading slash treated as root relative", input: "/docs/report.txt", want: "docs/report.txt"},
		{name: "backslash normalized", input: `docs\report.txt`, want: "docs/report.txt"},
		{name: "dot dot rejected", input: "../secret.txt", wantErr: true},
		{name: "nested dot dot rejected", input: "docs/../../secret.txt", wantErr: true},
		{name: "windows drive rejected", input: `C:\Users\secret.txt`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cleanRelativePath(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestShareFSResolveAndList(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "report.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "z.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}

	share, err := NewShareFS(root)
	if err != nil {
		t.Fatal(err)
	}
	abs, rel, err := share.Resolve("docs/report.txt")
	if err != nil {
		t.Fatalf("resolve failed: %v", err)
	}
	if rel != "docs/report.txt" || !strings.HasSuffix(filepath.ToSlash(abs), "/docs/report.txt") {
		t.Fatalf("unexpected resolve result abs=%q rel=%q", abs, rel)
	}
	if _, _, err := share.Resolve("../z.txt"); err == nil {
		t.Fatalf("expected traversal to fail")
	}
	_, legacyRel, err := share.Resolve(filepath.Join(root, "docs", "report.txt"))
	if err != nil {
		t.Fatalf("legacy absolute resolve failed: %v", err)
	}
	if legacyRel != "docs/report.txt" {
		t.Fatalf("legacy rel = %q", legacyRel)
	}

	files, rel, err := share.List("")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if rel != "" {
		t.Fatalf("root rel = %q", rel)
	}
	if len(files) != 2 {
		t.Fatalf("got %d files, want 2", len(files))
	}
	if !files[0].IsDir || files[0].FileName != "docs" {
		t.Fatalf("directories should be first and sorted, got %#v", files[0])
	}
}

func TestShareFSSkipsSymlinkOutsideRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation requires extra privileges on Windows")
	}

	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "outside")); err != nil {
		t.Fatal(err)
	}

	share, err := NewShareFS(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := share.Resolve("outside/secret.txt"); err == nil {
		t.Fatalf("expected symlink escape to fail")
	}
	files, _, err := share.List("")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("unsafe symlink should be hidden, got %#v", files)
	}
}
