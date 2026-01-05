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

func TestNewTunnel_MissingConfig(t *testing.T) {
	_, err := NewTunnel(nil, "remote-host", 1521, 0)
	if err == nil {
		t.Fatal("expected error for missing config")
	}

	expected := "config is required"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestNewTunnel_MissingRemoteHost(t *testing.T) {
	cfg, err := NewSSHConfig("paulo", "senha123", "", "bastion.com", "", 22)
	if err != nil {
		t.Fatalf("failed to create ssh config: %v", err)
	}

	_, err = NewTunnel(cfg, "", 1521, 0)
	if err == nil {
		t.Fatal("expected error for missing remoteHost")
	}

	expected := "remoteHost is required"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestNewTunnel_InvalidRemotePort(t *testing.T) {
	cfg, err := NewSSHConfig("paulo", "senha123", "", "bastion.com", "", 22)
	if err != nil {
		t.Fatalf("failed to create ssh config: %v", err)
	}

	tests := []struct {
		name string
		port int
	}{
		{"zero port", 0},
		{"negative port", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err = NewTunnel(cfg, "remote-host", tt.port, 0)
			if err == nil {
				t.Fatal("expected error for invalid remotePort")
			}

			expected := "remotePort must be greater than 0"
			if err.Error() != expected {
				t.Errorf("expected error '%s', got '%s'", expected, err.Error())
			}
		})
	}
}

func TestNewTunnel_SSHConnectionFailed(t *testing.T) {
	cfg, err := NewSSHConfig("paulo", "senha123", "", "localhost", "", 22222)
	if err != nil {
		t.Fatalf("failed to create ssh config: %v", err)
	}

	_, err = NewTunnel(cfg, "remote-host", 1521, 0)
	if err == nil {
		t.Fatal("expected error for failed ssh connection")
	}
}

func TestNewTunnel_AutomaticLocalPort(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	tun, err := NewTunnel(cfg, "127.0.0.1", 1521, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer tun.Close()

	if tun.LocalPort() <= 0 {
		t.Errorf("expected positive local port, got %d", tun.LocalPort())
	}
}

func TestNewTunnel_FixedLocalPort(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	// Encontra uma porta livre
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	freePort := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	tun, err := NewTunnel(cfg, "127.0.0.1", 1521, freePort)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer tun.Close()

	if tun.LocalPort() != freePort {
		t.Errorf("expected local port %d, got %d", freePort, tun.LocalPort())
	}
}

func TestTunnel_LocalAddr(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	tun, err := NewTunnel(cfg, "127.0.0.1", 1521, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer tun.Close()

	expected := fmt.Sprintf("127.0.0.1:%d", tun.LocalPort())
	if tun.LocalAddr() != expected {
		t.Errorf("expected LocalAddr '%s', got '%s'", expected, tun.LocalAddr())
	}
}

func TestTunnel_RemoteAddr(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	tun, err := NewTunnel(cfg, "remote-host", 1521, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer tun.Close()

	expected := "remote-host:1521"
	if tun.RemoteAddr() != expected {
		t.Errorf("expected RemoteAddr '%s', got '%s'", expected, tun.RemoteAddr())
	}
}

func TestTunnel_Close(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	tun, err := NewTunnel(cfg, "127.0.0.1", 1521, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	localAddr := tun.LocalAddr()

	err = tun.Close()
	if err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}

	// Verifica se a porta local foi liberada
	time.Sleep(100 * time.Millisecond)
	_, err = net.Dial("tcp", localAddr)
	if err == nil {
		t.Error("expected connection to fail after tunnel close")
	}
}

func TestTunnel_ForwardData(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	// Cria servidor de destino fake
	destServer := setupTestDestinationServer(t, "hello from oracle")
	defer destServer.Close()

	destPort := destServer.Addr().(*net.TCPAddr).Port

	tun, err := NewTunnel(cfg, "127.0.0.1", destPort, 0)
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

	// Lê a resposta
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("failed to read from tunnel: %v", err)
	}

	response := string(buf[:n])
	if response != "hello from oracle" {
		t.Errorf("expected 'hello from oracle', got '%s'", response)
	}
}

func TestTunnel_MultipleConnections(t *testing.T) {
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

	tun, err := NewTunnel(cfg, "127.0.0.1", destPort, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer tun.Close()

	for i := 1; i <= 3; i++ {
		conn, err := net.Dial("tcp", tun.LocalAddr())
		if err != nil {
			t.Fatalf("failed to connect to tunnel: %v", err)
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

	// Gera chave RSA para o servidor
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

	// Cria config do cliente (sem knownHostsFile = insecure)
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
		t.Fatalf("failed to create destination server: %v", err)
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
