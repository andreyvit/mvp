package mvpstatics

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreyvit/mvp/mvphttp"
	"github.com/uptrace/bunrouter"
)

func TestSetupRoute_does_not_serve_directory_indexes(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "root.txt"), "root-body")
	writeFile(t, filepath.Join(dir, "index.html"), "<html>idx</html>")
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dir, "sub", "nested.txt"), "nested-body")

	r := bunrouter.New()
	SetupRoute(&r.Group, "/files", os.DirFS(dir), mvphttp.Uncached, nil)

	get := func(path string) *httptest.ResponseRecorder {
		t.Helper()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		r.ServeHTTP(rec, req)
		return rec
	}

	rec := get("/files/root.txt")
	if rec.Code != http.StatusOK || rec.Body.String() != "root-body" {
		t.Fatalf("root.txt: code=%d body=%q", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Cache-Control") == "" {
		t.Fatal("missing cache header on file")
	}

	rec = get("/files/sub/nested.txt")
	if rec.Code != http.StatusOK || rec.Body.String() != "nested-body" {
		t.Fatalf("nested.txt: code=%d body=%q", rec.Code, rec.Body.String())
	}

	for _, path := range []string{"/files/", "/files/sub/", "/files/sub"} {
		rec := get(path)
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: code=%d want 404 body=%q", path, rec.Code, rec.Body.String())
		}
		body := rec.Body.String()
		if strings.Contains(body, "root.txt") || strings.Contains(body, "nested.txt") || strings.Contains(body, "idx") || strings.Contains(body, "<pre>") {
			t.Errorf("%s: listing leaked: %q", path, body)
		}
	}
}

func writeFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}
