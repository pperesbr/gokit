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

// TestNewTunnel verifies the creation of a new Tunnel instance and its initial state, ensuring proper configuration and status.
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

// TestNewTunnel_DoesNotConnect verifies that a newly created tunnel is not connected and remains in the 'stopped' status.
func TestNewTunnel_DoesNotConnect(t *testing.T) {
	cfg, _ := NewSSHConfig("user", "pass", "", "localhost", "", 22)

	tun := NewTunnel(cfg, "remote-host", 1521, 0)

	if tun.Status() != StatusStopped {
		t.Errorf("expected status %s, got %s", StatusStopped, tun.Status())
	}

	if tun.client != nil {
		t.Error("expected client to be nil")
	}
}

// TestValidate_Success verifies that a tunnel with valid configuration passes the validation without errors.
func TestValidate_Success(t *testing.T) {
	cfg, _ := NewSSHConfig("user", "pass", "", "localhost", "", 22)
	tun := NewTunnel(cfg, "remote-host", 1521, 0)

	err := tun.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestValidate_NilConfig verifies that the Validate method returns an error when the Tunnel is initialized with a nil SSHConfig.
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

// TestValidate_EmptyRemoteHost verifies that the validation fails when the remoteHost is empty, returning the expected error.
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

// TestValidate_InvalidRemotePort tests the validation of invalid remotePort values in Tunnel configuration.
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

// TestValidate_InvalidLocalPort verifies that Tunnel.Validate() returns an error when the localPort is set to a negative value.
func TestValidate_InvalidLocalPort(t *testing.T) {
	cfg, _ := NewSSHConfig("user", "pass", "", "localhost", "", 22)
	tun := NewTunnel(cfg, "remote-host", 1521, -1)

	err := tun.Validate()
	if err == nil {
		t.Fatal("expected error for negative localPort")
	}
}

// TestStart_Success verifies that the tunnel starts successfully, achieves a running status, and assigns a valid port.
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

// TestStart_AlreadyRunning validates that attempting to start an already running tunnel returns an appropriate error.
func TestStart_AlreadyRunning(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	tun := NewTunnel(cfg, "127.0.0.1", 1521, 0)

	err := tun.Start()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer tun.Close()

	err = tun.Start()
	if err == nil {
		t.Fatal("expected error for already running tunnel")
	}
}

// TestStart_ValidationError verifies that the Tunnel's Start method correctly handles invalid configurations and updates its status.
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

// TestStart_SSHConnectionFailed verifies that the tunnel's Start method handles a failed SSH connection gracefully.
// It checks if an error is returned and ensures the tunnel's status is set to StatusError.
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

// TestStart_FixedLocalPort verifies that the tunnel starts successfully with a fixed local port and matches the expected port.
func TestStart_FixedLocalPort(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

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

// TestStop_Success verifies that the Tunnel's Stop method successfully terminates the connection and updates the status.
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

// TestStop_AlreadyStopped verifies that calling Stop on a tunnel that hasn't been started does not return an error.
func TestStop_AlreadyStopped(t *testing.T) {
	cfg, _ := NewSSHConfig("user", "pass", "", "localhost", "", 22)
	tun := NewTunnel(cfg, "remote-host", 1521, 0)

	err := tun.Stop()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestStop_ReleasesPort verifies that the tunnel releases the local port and prevents new connections after stopping.
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

	time.Sleep(100 * time.Millisecond)

	_, err = net.Dial("tcp", localAddr)
	if err == nil {
		t.Error("expected connection to fail after stop")
	}
}

// TestRestart_Success verifies that a tunnel can successfully restart and maintains the expected StatusRunning state afterward.
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

// TestRestart_FromStopped verifies the behavior of the Tunnel's Restart method when invoked on a tunnel in a stopped state.
func TestRestart_FromStopped(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	tun := NewTunnel(cfg, "127.0.0.1", 1521, 0)

	err := tun.Restart()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if tun.Status() != StatusRunning {
		t.Errorf("expected status %s, got %s", StatusRunning, tun.Status())
	}

	defer tun.Close()
}

// TestUpdateConfig verifies the behavior of the Tunnel's UpdateConfig method by ensuring the SSH configuration is updated correctly.
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

// TestLocalAddr verifies that the LocalAddr method of the Tunnel returns the correct formatted local address and port.
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

// TestRemoteAddr verifies that the Tunnel's RemoteAddr method returns the expected remote address in the correct format.
func TestRemoteAddr(t *testing.T) {
	cfg, _ := NewSSHConfig("user", "pass", "", "localhost", "", 22)
	tun := NewTunnel(cfg, "oracle-server", 1521, 0)

	expected := "oracle-server:1521"
	if tun.RemoteAddr() != expected {
		t.Errorf("expected '%s', got '%s'", expected, tun.RemoteAddr())
	}
}

// TestLastError_NilWhenNoError ensures that calling LastError on a tunnel returns nil when no error has occurred.
func TestLastError_NilWhenNoError(t *testing.T) {
	cfg, _ := NewSSHConfig("user", "pass", "", "localhost", "", 22)
	tun := NewTunnel(cfg, "remote-host", 1521, 0)

	if tun.LastError() != nil {
		t.Errorf("expected nil error, got %v", tun.LastError())
	}
}

// TestForwardData verifies the functionality of tunneling data through an SSH server to a destination server.
func TestForwardData(t *testing.T) {
	sshServer, cfg := setupTestSSHServer(t)
	defer sshServer.Close()

	destServer := setupTestDestinationServer(t, "hello from oracle")
	defer destServer.Close()

	destPort := destServer.Addr().(*net.TCPAddr).Port

	tun := NewTunnel(cfg, "127.0.0.1", destPort, 0)

	err := tun.Start()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer tun.Close()

	conn, err := net.Dial("tcp", tun.LocalAddr())
	if err != nil {
		t.Fatalf("failed to connect to tunnel: %v", err)
	}
	defer conn.Close()

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

// TestMultipleConnections verifies if multiple sequential connections to the tunnel are handled correctly without errors.
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

// setupTestSSHServer creates and starts an SSH server for testing purposes and returns the listener and SSH config.
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

// handleTestSSHConnection manages an incoming SSH connection and handles direct-tcpip channel requests for forwarding.
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

// setupTestDestinationServer creates a test TCP server that sends a fixed response to incoming connections.
func setupTestDestinationServer(t *testing.T, response string) net.Listener {
	t.Helper()
	return setupTestDestinationServerFunc(t, func(conn net.Conn) {
		conn.Write([]byte(response))
		conn.Close()
	})
}

// setupTestDestinationServerFunc creates a test TCP server, uses the provided handler for incoming connections, and returns a listener.
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
