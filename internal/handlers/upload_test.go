package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newUploadRouter() (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	h := New(nil)
	r := gin.New()
	r.POST("/upload", h.UploadFile)
	r.GET("/files/*filepath", h.ServeFile)
	return r, h
}

func buildMultipart(t *testing.T, fieldname, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile(fieldname, filename)
	if err != nil {
		t.Fatal(err)
	}
	fw.Write(content)
	w.Close()
	return &buf, w.FormDataContentType()
}

func TestUploadFile_PDF_Valid(t *testing.T) {
	r, _ := newUploadRouter()
	body, ct := buildMultipart(t, "file", "test.pdf", []byte("%PDF-1.4 fake pdf content"))
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "fileUrl") {
		t.Fatalf("expected fileUrl in response, got: %s", w.Body.String())
	}
}

func TestUploadFile_BlockedExtension(t *testing.T) {
	r, _ := newUploadRouter()
	body, ct := buildMultipart(t, "file", "evil.sh", []byte("#!/bin/bash\nrm -rf /"))
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for .sh, got %d", w.Code)
	}
}

func TestUploadFile_BlockedExtension_JSON(t *testing.T) {
	r, _ := newUploadRouter()
	body, ct := buildMultipart(t, "file", "data.json", []byte(`{"key":"value"}`))
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for .json, got %d", w.Code)
	}
}

func TestUploadFile_FakePDF(t *testing.T) {
	r, _ := newUploadRouter()
	// .pdf extension but content is not a PDF
	body, ct := buildMultipart(t, "file", "evil.pdf", []byte("<script>alert(1)</script>"))
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for fake PDF, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUploadFile_SizeLimit(t *testing.T) {
	r, _ := newUploadRouter()
	large := make([]byte, maxUploadSize+1)
	body, ct := buildMultipart(t, "file", "big.pdf", large)
	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", ct)
	req.ContentLength = int64(body.Len())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", w.Code)
	}
}

func TestServeFile_NotFound(t *testing.T) {
	r, _ := newUploadRouter()
	req := httptest.NewRequest(http.MethodGet, "/files/nonexistent.pdf", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	// Must NOT trigger a file download (no Content-Disposition on error)
	if w.Header().Get("Content-Disposition") != "" {
		t.Fatal("Content-Disposition should not be set on 404")
	}
}

func TestServeFile_PathTraversal(t *testing.T) {
	r, _ := newUploadRouter()
	req := httptest.NewRequest(http.MethodGet, "/files/../../etc/passwd", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code == http.StatusOK {
		t.Fatal("path traversal should not return 200")
	}
}

func TestServeFile_ExistingFile(t *testing.T) {
	r, _ := newUploadRouter()
	// Place a real file in uploadDir
	name := "testserve.pdf"
	path := filepath.Join(uploadDir, name)
	os.WriteFile(path, []byte("%PDF-1.4 test"), 0644)
	defer os.Remove(path)

	req := httptest.NewRequest(http.MethodGet, "/files/"+name, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
