package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDetectImageMimeType 验证图片 MIME 类型识别（基于魔数）
func TestDetectImageMimeType(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"png", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0, 0, 0, 0}, "image/png"},
		{"jpeg", []byte{0xFF, 0xD8, 0xFF, 0xE0, 0, 0, 0, 0, 0, 0, 0, 0}, "image/jpeg"},
		{"gif", []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0, 0, 0, 0, 0, 0}, "image/gif"},
		{"webp", []byte{0x52, 0x49, 0x46, 0x46, 0, 0, 0, 0, 0x57, 0x45, 0x42, 0x50}, "image/webp"},
		{"too short defaults to jpeg", []byte{0xFF, 0xD8}, "image/jpeg"},
		{"unknown defaults to jpeg", []byte("hello world data"), "image/jpeg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectImageMimeType(tt.data); got != tt.want {
				t.Errorf("detectImageMimeType() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestImageToBase64LocalFile 验证本地图片转 base64 data URI
func TestImageToBase64LocalFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.png")
	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 1, 2, 3, 4}
	if err := os.WriteFile(path, pngData, 0o644); err != nil {
		t.Fatalf("create temp file: %v", err)
	}

	got, err := ImageToBase64(path)
	if err != nil {
		t.Fatalf("ImageToBase64() error: %v", err)
	}
	if !strings.HasPrefix(got, "data:image/png;base64,") {
		t.Errorf("ImageToBase64() should start with data:image/png;base64,, got: %s", got[:40])
	}
}

// TestImageToBase64MissingFile 验证文件不存在时返回错误
func TestImageToBase64MissingFile(t *testing.T) {
	if _, err := ImageToBase64("/nonexistent/path/abc.png"); err == nil {
		t.Error("ImageToBase64() expected error for missing file, got nil")
	}
}
