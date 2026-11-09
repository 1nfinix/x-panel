package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUploadStaticSecurityHeaders(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "old.html"), []byte("<script>alert(1)</script>"), 0644); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	registerUploadStaticRoutes(router, "/conf/uploads", dir)
	for _, route := range []string{"/conf/uploads/old.html", "/uploads/old.html"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, route, nil))
		if w.Code != 200 || w.Header().Get("X-Content-Type-Options") != "nosniff" || !strings.Contains(w.Header().Get("Content-Security-Policy"), "sandbox") {
			t.Fatalf("%s: status %d, headers %v", route, w.Code, w.Header())
		}
	}
}
