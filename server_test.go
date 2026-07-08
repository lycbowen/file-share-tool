package fileshare

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "中文 file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	share, err := NewShareFS(root)
	if err != nil {
		t.Fatal(err)
	}
	app := NewAppServer(share, http.Dir(root), "http://192.168.1.8:9001")
	return httptest.NewServer(app.Routes()), root
}

func TestFilesHandler(t *testing.T) {
	ts, _ := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/files")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.StatusCode, body)
	}
	text := string(body)
	if !strings.Contains(text, `"public_url":"http://192.168.1.8:9001"`) {
		t.Fatalf("public url missing: %s", text)
	}
	if !strings.Contains(text, `"path":"docs"`) {
		t.Fatalf("child relative path missing: %s", text)
	}
}

func TestDownloadHandler(t *testing.T) {
	ts, _ := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/download?path=" + "docs%2F%E4%B8%AD%E6%96%87%20file.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.StatusCode, body)
	}
	if string(body) != "hello" {
		t.Fatalf("body=%q", body)
	}
	contentDisposition := resp.Header.Get("Content-Disposition")
	if !strings.Contains(contentDisposition, "filename*=") || !strings.Contains(contentDisposition, "%E4%B8%AD%E6%96%87%20file.txt") {
		t.Fatalf("bad content disposition: %s", contentDisposition)
	}
}

func TestDownloadRejectsDirectoryAndBadMethod(t *testing.T) {
	ts, _ := newTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/download?path=docs")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("directory download status=%d", resp.StatusCode)
	}

	req, err := http.NewRequest(http.MethodPost, ts.URL+"/api/files", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("method status=%d", resp.StatusCode)
	}
}
