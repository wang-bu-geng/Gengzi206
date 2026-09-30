package security

import (
	"os"
	"strings"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	os.Setenv(EnvKeyName, "test-secret-key")
	defer os.Unsetenv(EnvKeyName)

	if !Init() {
		t.Fatal("Init() should succeed with key set")
	}

	plain := "sk-test-12345"
	enc := Encrypt(plain)
	if enc == plain {
		t.Error("Encrypt() should produce different value when enabled")
	}
	if !strings.HasPrefix(enc, "enc:v1:") {
		t.Errorf("Encrypt() should produce enc:v1: prefix, got %q", enc)
	}
	if got := Decrypt(enc); got != plain {
		t.Errorf("Decrypt(Encrypt(x)) = %q, want %q", got, plain)
	}
}

func TestEncryptEmpty(t *testing.T) {
	os.Setenv(EnvKeyName, "test-secret-key")
	defer os.Unsetenv(EnvKeyName)
	Init()

	if got := Encrypt(""); got != "" {
		t.Errorf("Encrypt(\"\") = %q, want empty", got)
	}
}

func TestDecryptLegacyPlaintext(t *testing.T) {
	os.Setenv(EnvKeyName, "test-secret-key")
	defer os.Unsetenv(EnvKeyName)
	Init()

	// 遗留明文（无前缀）应原样返回
	const legacy = "plaintext-key"
	if got := Decrypt(legacy); got != legacy {
		t.Errorf("legacy plaintext should pass through, got %q", got)
	}
}

func TestDisabledPassthrough(t *testing.T) {
	os.Unsetenv(EnvKeyName)
	if Init() {
		t.Error("Init() should return false without key")
	}
	if Enabled() {
		t.Error("Enabled() should be false without key")
	}

	plain := "sk-plain"
	if got := Encrypt(plain); got != plain {
		t.Errorf("Encrypt() should pass through when disabled, got %q", got)
	}
	if got := Decrypt(plain); got != plain {
		t.Errorf("Decrypt() should pass through when disabled, got %q", got)
	}
}
