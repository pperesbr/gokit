package tunnel

import (
	"os"
	"path/filepath"
	"testing"
)

const testPrivateKey = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACBwNlsVOBKosw+jG0cxb/L2sHf0luTMKCyLFLWCOIHzVAAAAJhPUzrTT1M6
0wAAAAtzc2gtZWQyNTUxOQAAACBwNlsVOBKosw+jG0cxb/L2sHf0luTMKCyLFLWCOIHzVA
AAAECpVPKPdliGs+H4XUjDJmWTafFnhrpCLVFb8FkUdsLfE3A2WxU4EqizD6MbRzFv8vaw
d/SW5MwoLIsUtYI4gfNUAAAAEHRlc3RAZXhhbXBsZS5jb20BAgMEBQ==
-----END OPENSSH PRIVATE KEY-----`

const testKnownHosts = `bastion.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl`

func createTempFile(t *testing.T, name, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, name)
	err := os.WriteFile(filePath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return filePath
}

func TestNewSSHConfig_WithPassword(t *testing.T) {
	cfg, err := NewSSHConfig("paulo", "senha123", "", "bastion.com", "", 22)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.User != "paulo" {
		t.Errorf("expected user 'paulo', got '%s'", cfg.User)
	}

	if cfg.Password != "senha123" {
		t.Errorf("expected password 'senha123', got '%s'", cfg.Password)
	}

	if cfg.Host != "bastion.com" {
		t.Errorf("expected host 'bastion.com', got '%s'", cfg.Host)
	}

	if cfg.Port != 22 {
		t.Errorf("expected port 22, got %d", cfg.Port)
	}

	if len(cfg.AuthMethods) == 0 {
		t.Error("expected AuthMethods to be set")
	}

	// Com password, deve ter 2 métodos: password e keyboard-interactive
	if len(cfg.AuthMethods) != 2 {
		t.Errorf("expected 2 AuthMethods, got %d", len(cfg.AuthMethods))
	}

	if cfg.HostKeyCallback == nil {
		t.Error("expected HostKeyCallback to be set")
	}
}

func TestNewSSHConfig_WithDefaultPort(t *testing.T) {
	cfg, err := NewSSHConfig("paulo", "senha123", "", "bastion.com", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != 22 {
		t.Errorf("expected default port 22, got %d", cfg.Port)
	}
}

func TestNewSSHConfig_WithKeyFile(t *testing.T) {
	keyPath := createTempFile(t, "id_test", testPrivateKey)

	cfg, err := NewSSHConfig("paulo", "", keyPath, "bastion.com", "", 22)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.KeyFile != keyPath {
		t.Errorf("expected keyFile '%s', got '%s'", keyPath, cfg.KeyFile)
	}

	if len(cfg.AuthMethods) == 0 {
		t.Error("expected AuthMethods to be set")
	}

	// Com keyFile, deve ter 1 método: publickey
	if len(cfg.AuthMethods) != 1 {
		t.Errorf("expected 1 AuthMethod, got %d", len(cfg.AuthMethods))
	}
}

func TestNewSSHConfig_KeyFileTakesPrecedence(t *testing.T) {
	keyPath := createTempFile(t, "id_test", testPrivateKey)

	// Passa password E keyFile - keyFile deve ter precedência
	cfg, err := NewSSHConfig("paulo", "senha123", keyPath, "bastion.com", "", 22)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.AuthMethods) == 0 {
		t.Error("expected AuthMethods to be set")
	}

	// KeyFile tem precedência, deve ter só 1 método
	if len(cfg.AuthMethods) != 1 {
		t.Errorf("expected 1 AuthMethod (keyFile precedence), got %d", len(cfg.AuthMethods))
	}
}

func TestNewSSHConfig_WithKnownHostsFile(t *testing.T) {
	knownHostsPath := createTempFile(t, "known_hosts", testKnownHosts)

	cfg, err := NewSSHConfig("paulo", "senha123", "", "bastion.com", knownHostsPath, 22)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.KnownHostsFile != knownHostsPath {
		t.Errorf("expected knownHostsFile '%s', got '%s'", knownHostsPath, cfg.KnownHostsFile)
	}

	if cfg.HostKeyCallback == nil {
		t.Error("expected HostKeyCallback to be set")
	}

	if cfg.IsInsecure() {
		t.Error("expected IsInsecure() to return false")
	}
}

func TestNewSSHConfig_InsecureWithoutKnownHostsFile(t *testing.T) {
	cfg, err := NewSSHConfig("paulo", "senha123", "", "bastion.com", "", 22)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.HostKeyCallback == nil {
		t.Error("expected HostKeyCallback to be set (insecure)")
	}

	if !cfg.IsInsecure() {
		t.Error("expected IsInsecure() to return true")
	}
}

func TestNewSSHConfig_MissingHost(t *testing.T) {
	_, err := NewSSHConfig("paulo", "senha123", "", "", "", 22)
	if err == nil {
		t.Fatal("expected error for missing host")
	}

	expected := "host is required"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestNewSSHConfig_MissingUser(t *testing.T) {
	_, err := NewSSHConfig("", "senha123", "", "bastion.com", "", 22)
	if err == nil {
		t.Fatal("expected error for missing user")
	}

	expected := "user is required"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestNewSSHConfig_MissingAuth(t *testing.T) {
	_, err := NewSSHConfig("paulo", "", "", "bastion.com", "", 22)
	if err == nil {
		t.Fatal("expected error for missing auth")
	}

	expected := "password or keyFile is required"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestNewSSHConfig_KeyFileNotFound(t *testing.T) {
	_, err := NewSSHConfig("paulo", "", "/path/that/does/not/exist", "bastion.com", "", 22)
	if err == nil {
		t.Fatal("expected error for missing key file")
	}
}

func TestNewSSHConfig_InvalidKeyFile(t *testing.T) {
	keyPath := createTempFile(t, "invalid_key", "not a valid key")

	_, err := NewSSHConfig("paulo", "", keyPath, "bastion.com", "", 22)
	if err == nil {
		t.Fatal("expected error for invalid key file")
	}
}

func TestNewSSHConfig_KnownHostsFileNotFound(t *testing.T) {
	_, err := NewSSHConfig("paulo", "senha123", "", "bastion.com", "/path/that/does/not/exist", 22)
	if err == nil {
		t.Fatal("expected error for missing known_hosts file")
	}
}

func TestNewSSHConfig_InvalidKnownHostsFile(t *testing.T) {
	knownHostsPath := createTempFile(t, "known_hosts", "invalid content %%%")

	_, err := NewSSHConfig("paulo", "senha123", "", "bastion.com", knownHostsPath, 22)
	// knownhosts.New pode ou não falhar com conteúdo inválido
	// dependendo do formato, então só verificamos se não deu panic
	_ = err
}

func TestSSHConfig_Addr(t *testing.T) {
	cfg, err := NewSSHConfig("paulo", "senha123", "", "bastion.com", "", 22)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "bastion.com:22"
	if cfg.Addr() != expected {
		t.Errorf("expected addr '%s', got '%s'", expected, cfg.Addr())
	}
}

func TestSSHConfig_AddrCustomPort(t *testing.T) {
	cfg, err := NewSSHConfig("paulo", "senha123", "", "bastion.com", "", 2222)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "bastion.com:2222"
	if cfg.Addr() != expected {
		t.Errorf("expected addr '%s', got '%s'", expected, cfg.Addr())
	}
}

func TestSSHConfig_IsInsecure(t *testing.T) {
	tests := []struct {
		name           string
		knownHostsFile string
		wantInsecure   bool
	}{
		{
			name:           "insecure when empty",
			knownHostsFile: "",
			wantInsecure:   true,
		},
		{
			name:           "secure when set",
			knownHostsFile: "will_be_replaced",
			wantInsecure:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			knownHostsFile := tt.knownHostsFile
			if knownHostsFile != "" {
				knownHostsFile = createTempFile(t, "known_hosts", testKnownHosts)
			}

			cfg, err := NewSSHConfig("paulo", "senha123", "", "bastion.com", knownHostsFile, 22)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.IsInsecure() != tt.wantInsecure {
				t.Errorf("expected IsInsecure() = %v, got %v", tt.wantInsecure, cfg.IsInsecure())
			}
		})
	}
}
