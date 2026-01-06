package tunnel

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

// =============================================================================
// Testes do NewTunnel
// =============================================================================

func TestNewTunnel(t *testing.T) {
	cfg, _ := NewSSHConfig("user", "pass", "", "localhost", "", 22)

	tun := NewTunnel(cfg, "remote-host", 1521, 0)

	if tun == nil {
		t.Fatal("expected tunnel to be created")
	}

	if tun.Status() != StatusStopped {
		t.Errorf("expected status %s, got %s", StatusStopped, tun.Status())
	}
}

func TestNewTunnel_DoesNotConnect(t *testing.T) {
	cfg, _ := NewSSHConfig("user", "pass", "", "localhost", "", 22)

	tun := NewTunnel(cfg, "remote-host", 1521, 0)

	// Deve estar parado, não conectado
	if tun.Status() != StatusStopped {
		t.Errorf("expected status %s, got %s", StatusStopped, tun.Status())
	}

	// Client deve ser nil (não conectou)
	if tun.client != nil {
		t.Error("expected client to be nil")
	}
}

// =============================================================================
// Testes do Validate
// =============================================================================

func TestValidate_Success(t *testing.T) {
	cfg, _ := NewSSHConfig("user", "pass", "", "localhost", "", 22)
	tun := NewTunnel(cfg, "remote-host", 1521, 0)

	err := tun.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_NilConfig(t *testing.T) {
	tun := NewTunnel(nil, "remote-host", 1521, 0)

	err := tun.Validate()
	if err == nil {
		t.Fatal("expected error for nil config")
	}

	expected := "config is required"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestValidate_EmptyRemoteHost(t *testing.T) {
	cfg, _ := NewSSHConfig("user", "pass", "", "localhost", "", 22)
	tun := NewTunnel(cfg, "", 1521, 0)

	err := tun.Validate()
	if err == nil {
		t.Fatal("expected error for empty remoteHost")
	}

	expected := "remoteHost is required"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestValidate_InvalidRemotePort(t *testing.T) {
	cfg, _ := NewSSHConfig("user", "pass", "", "localhost", "", 22)

	tests := []struct {
		name string
		port int
	}{
		{"zero", 0},
		{"negative", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tun := NewTunnel(cfg, "remote-host", tt.port, 0)

			err := tun.Validate()
			if err == nil {
				t.Fatal("expected error for invalid remotePort")
			}
		})
	}
}

func TestValidate_InvalidLocalPort(t *testing.T) {
	cfg, _ := NewSSHConfig("user", "pass", "", "localhost", "", 22)
	tun := NewTunnel(cfg, "remote-host", 1521, -1)

	err := tun.Validate()
	if err == nil {
		t.Fatal("expected error for negative localPort")
	}
}

// =============================================================================
// Testes do Start
// =============================================================================

func TestStart_Success(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	tun := NewTunnel(cfg, "127.0.0.1", 1521, 0)

	err := tun.Start()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer tun.Close()

	if tun.Status() != StatusRunning {
		t.Errorf("expected status %s, got %s", StatusRunning, tun.Status())
	}

	if tun.LocalPort() <= 0 {
		t.Errorf("expected positive local port, got %d", tun.LocalPort())
	}
}

func TestStart_AlreadyRunning(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	tun := NewTunnel(cfg, "127.0.0.1", 1521, 0)

	err := tun.Start()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer tun.Close()

	// Tenta iniciar novamente
	err = tun.Start()
	if err == nil {
		t.Fatal("expected error for already running tunnel")
	}
}

func TestStart_ValidationError(t *testing.T) {
	tun := NewTunnel(nil, "remote-host", 1521, 0)

	err := tun.Start()
	if err == nil {
		t.Fatal("expected error for invalid config")
	}

	if tun.Status() != StatusError {
		t.Errorf("expected status %s, got %s", StatusError, tun.Status())
	}

	if tun.LastError() == nil {
		t.Error("expected LastError to be set")
	}
}

func TestStart_SSHConnectionFailed(t *testing.T) {
	cfg, _ := NewSSHConfig("user", "pass", "", "localhost", "", 59999)
	tun := NewTunnel(cfg, "remote-host", 1521, 0)

	err := tun.Start()
	if err == nil {
		t.Fatal("expected error for failed SSH connection")
	}

	if tun.Status() != StatusError {
		t.Errorf("expected status %s, got %s", StatusError, tun.Status())
	}
}

func TestStart_FixedLocalPort(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	// Encontra porta livre
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	freePort := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	tun := NewTunnel(cfg, "127.0.0.1", 1521, freePort)

	err = tun.Start()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer tun.Close()

	if tun.LocalPort() != freePort {
		t.Errorf("expected local port %d, got %d", freePort, tun.LocalPort())
	}
}

// =============================================================================
// Testes do Stop
// =============================================================================

func TestStop_Success(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	tun := NewTunnel(cfg, "127.0.0.1", 1521, 0)

	err := tun.Start()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = tun.Stop()
	if err != nil {
		t.Errorf("unexpected error on stop: %v", err)
	}

	if tun.Status() != StatusStopped {
		t.Errorf("expected status %s, got %s", StatusStopped, tun.Status())
	}
}

func TestStop_AlreadyStopped(t *testing.T) {
	cfg, _ := NewSSHConfig("user", "pass", "", "localhost", "", 22)
	tun := NewTunnel(cfg, "remote-host", 1521, 0)

	// Nunca foi iniciado, já está parado
	err := tun.Stop()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStop_ReleasesPort(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	tun := NewTunnel(cfg, "127.0.0.1", 1521, 0)

	err := tun.Start()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	localAddr := tun.LocalAddr()

	err = tun.Stop()
	if err != nil {
		t.Fatalf("unexpected error on stop: %v", err)
	}

	// Aguarda um pouco para liberar a porta
	time.Sleep(100 * time.Millisecond)

	// Tenta conectar - deve falhar
	_, err = net.Dial("tcp", localAddr)
	if err == nil {
		t.Error("expected connection to fail after stop")
	}
}

// =============================================================================
// Testes do Restart
// =============================================================================

func TestRestart_Success(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	tun := NewTunnel(cfg, "127.0.0.1", 1521, 0)

	err := tun.Start()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer tun.Close()

	err = tun.Restart()
	if err != nil {
		t.Errorf("unexpected error on restart: %v", err)
	}

	if tun.Status() != StatusRunning {
		t.Errorf("expected status %s, got %s", StatusRunning, tun.Status())
	}
}

func TestRestart_FromStopped(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	tun := NewTunnel(cfg, "127.0.0.1", 1521, 0)

	// Nunca foi iniciado
	err := tun.Restart()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if tun.Status() != StatusRunning {
		t.Errorf("expected status %s, got %s", StatusRunning, tun.Status())
	}

	defer tun.Close()
}

// =============================================================================
// Testes do UpdateConfig
// =============================================================================

func TestUpdateConfig(t *testing.T) {
	cfg1, _ := NewSSHConfig("user1", "pass1", "", "host1", "", 22)
	cfg2, _ := NewSSHConfig("user2", "pass2", "", "host2", "", 22)

	tun := NewTunnel(cfg1, "remote-host", 1521, 0)

	if tun.config.User != "user1" {
		t.Errorf("expected user 'user1', got '%s'", tun.config.User)
	}

	tun.UpdateConfig(cfg2)

	if tun.config.User != "user2" {
		t.Errorf("expected user 'user2', got '%s'", tun.config.User)
	}
}

// =============================================================================
// Testes dos Getters
// =============================================================================

func TestLocalAddr(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	tun := NewTunnel(cfg, "127.0.0.1", 1521, 0)

	err := tun.Start()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer tun.Close()

	expected := fmt.Sprintf("127.0.0.1:%d", tun.LocalPort())
	if tun.LocalAddr() != expected {
		t.Errorf("expected '%s', got '%s'", expected, tun.LocalAddr())
	}
}

func TestRemoteAddr(t *testing.T) {
	cfg, _ := NewSSHConfig("user", "pass", "", "localhost", "", 22)
	tun := NewTunnel(cfg, "oracle-server", 1521, 0)

	expected := "oracle-server:1521"
	if tun.RemoteAddr() != expected {
		t.Errorf("expected '%s', got '%s'", expected, tun.RemoteAddr())
	}
}

func TestLastError_NilWhenNoError(t *testing.T) {
	cfg, _ := NewSSHConfig("user", "pass", "", "localhost", "", 22)
	tun := NewTunnel(cfg, "remote-host", 1521, 0)

	if tun.LastError() != nil {
		t.Errorf("expected nil error, got %v", tun.LastError())
	}
}

// =============================================================================
// Testes de Forward de Dados
// =============================================================================

func TestForwardData(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	// Servidor de destino fake
	destServer := setupTestDestinationServer(t, "hello from oracle")
	defer destServer.Close()

	destPort := destServer.Addr().(*net.TCPAddr).Port

	tun := NewTunnel(cfg, "127.0.0.1", destPort, 0)

	err := tun.Start()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer tun.Close()

	// Conecta no tunnel
	conn, err := net.Dial("tcp", tun.LocalAddr())
	if err != nil {
		t.Fatalf("failed to connect to tunnel: %v", err)
	}
	defer conn.Close()

	// Lê resposta
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("failed to read: %v", err)
	}

	response := string(buf[:n])
	if response != "hello from oracle" {
		t.Errorf("expected 'hello from oracle', got '%s'", response)
	}
}

func TestMultipleConnections(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	connCount := 0
	destServer := setupTestDestinationServerFunc(t, func(conn net.Conn) {
		connCount++
		fmt.Fprintf(conn, "connection %d", connCount)
		conn.Close()
	})
	defer destServer.Close()

	destPort := destServer.Addr().(*net.TCPAddr).Port

	tun := NewTunnel(cfg, "127.0.0.1", destPort, 0)

	err := tun.Start()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer tun.Close()

	// Faz 3 conexões
	for i := 1; i <= 3; i++ {
		conn, err := net.Dial("tcp", tun.LocalAddr())
		if err != nil {
			t.Fatalf("failed to connect: %v", err)
		}

		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		buf := make([]byte, 1024)
		n, _ := conn.Read(buf)
		conn.Close()

		expected := fmt.Sprintf("connection %d", i)
		if string(buf[:n]) != expected {
			t.Errorf("expected '%s', got '%s'", expected, string(buf[:n]))
		}
	}
}

// =============================================================================
// Test Helpers
// =============================================================================

func setupTestSSHServer(t *testing.T) (net.Listener, *SSHConfig) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}

	serverConfig := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == "testuser" && string(pass) == "testpass" {
				return nil, nil
			}
			return nil, fmt.Errorf("invalid credentials")
		},
	}
	serverConfig.AddHostKey(signer)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleTestSSHConnection(conn, serverConfig)
		}
	}()

	port := listener.Addr().(*net.TCPAddr).Port
	cfg, err := NewSSHConfig("testuser", "testpass", "", "127.0.0.1", "", port)
	if err != nil {
		listener.Close()
		t.Fatalf("failed to create ssh config: %v", err)
	}

	return listener, cfg
}

func handleTestSSHConnection(conn net.Conn, config *ssh.ServerConfig) {
	defer conn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		return
	}
	defer sshConn.Close()

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() == "direct-tcpip" {
			channel, requests, err := newChannel.Accept()
			if err != nil {
				continue
			}
			go ssh.DiscardRequests(requests)

			var payload struct {
				DestHost   string
				DestPort   uint32
				OriginHost string
				OriginPort uint32
			}
			ssh.Unmarshal(newChannel.ExtraData(), &payload)

			destAddr := fmt.Sprintf("%s:%d", payload.DestHost, payload.DestPort)
			destConn, err := net.Dial("tcp", destAddr)
			if err != nil {
				channel.Close()
				continue
			}

			go func() {
				defer channel.Close()
				defer destConn.Close()
				io.Copy(channel, destConn)
			}()
			go func() {
				defer channel.Close()
				defer destConn.Close()
				io.Copy(destConn, channel)
			}()
		}
	}
}

func setupTestDestinationServer(t *testing.T, response string) net.Listener {
	t.Helper()
	return setupTestDestinationServerFunc(t, func(conn net.Conn) {
		conn.Write([]byte(response))
		conn.Close()
	})
}

func setupTestDestinationServerFunc(t *testing.T, handler func(net.Conn)) net.Listener {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handler(conn)
		}
	}()

	return listener
}
