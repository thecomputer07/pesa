package imagecache

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStore_PutHappyPath(t *testing.T) {
	dir := t.TempDir()
	s, err := NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	body := []byte("fake-jpeg-bytes")
	sha, ext, err := s.Put("off", body, "image/jpeg", SidecarMeta{
		Source: "off", License: "ODbL-1.0", LicenseURL: "https://opendatacommons.org/licenses/odbl/1-0/",
		Attribution: "Open Food Facts contributors — ODbL", OriginalURL: "https://example.invalid/a.jpg",
	})
	if err != nil {
		t.Fatal(err)
	}
	if ext != ".jpg" || len(sha) != 64 {
		t.Errorf("ext=%q sha=%q", ext, sha)
	}
	if _, err := os.Stat(filepath.Join(dir, "off", sha+".jpg")); err != nil {
		t.Errorf("binary missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "off", sha+".license.json")); err != nil {
		t.Errorf("sidecar missing: %v", err)
	}
	if !s.Exists("off", sha+".jpg") {
		t.Errorf("Exists should be true")
	}
}

func TestStore_PutIdempotent(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	body := []byte("x")
	sha1, _, err := s.Put("off", body, "image/png", SidecarMeta{Source: "off"})
	if err != nil {
		t.Fatal(err)
	}
	sha2, _, err := s.Put("off", body, "image/png", SidecarMeta{Source: "off"})
	if err != nil {
		t.Fatal(err)
	}
	if sha1 != sha2 {
		t.Errorf("sha mismatch: %s vs %s", sha1, sha2)
	}
}

func TestStore_UnsupportedContentType(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	_, _, err := s.Put("off", []byte("x"), "application/pdf", SidecarMeta{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStore_UnknownSource(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	_, _, err := s.Put("evil", []byte("x"), "image/jpeg", SidecarMeta{})
	if err != ErrUnknownSource {
		t.Fatalf("got %v", err)
	}
}

func TestStore_ReadMeta(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	sha, ext, err := s.Put("off", []byte("x"), "image/jpeg", SidecarMeta{
		Source: "off", License: "ODbL-1.0", Attribution: "OFF",
	})
	if err != nil {
		t.Fatal(err)
	}
	m, err := s.ReadMeta("off", sha+ext)
	if err != nil {
		t.Fatal(err)
	}
	if m.License != "ODbL-1.0" || m.Attribution != "OFF" {
		t.Errorf("meta mismatch: %+v", m)
	}
	if m.Bytes != 1 || !strings.HasPrefix(m.ContentType, "image/") {
		t.Errorf("bytes/ct: %+v", m)
	}
}

func TestStore_PathTraversalRejected(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)
	if _, err := s.Path("off", "../secret"); err == nil {
		t.Errorf("traversal should be rejected")
	}
}
