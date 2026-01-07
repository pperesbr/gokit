package tunnel

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// Status defines a string-based enumeration for representing the operational state of a process or system.
type Status string

const (
	StatusStopped  Status = "stopped"
	StatusStarting Status = "starting"
	StatusRunning  Status = "running"
	StatusError    Status = "error"
)

// Stats represent statistical data related to network connections and activity over a specific period of time.
type Stats struct {
	BytesIn           int64
	BytesOut          int64
	Connections       int64
	ActiveConnections int64
	LastActivity      time.Time
	StartedAt         time.Time
}

// Tunnel represents a secure SSH-based port forwarding connection between a local and a remote host.
type Tunnel struct {
	config     *SSHConfig
	remoteHost string
	remotePort int
	localPort  int

	client     *ssh.Client
	listener   net.Listener
	actualPort int

	status    Status
	lastError error
	stats     Stats

	done chan struct{}
	mu   sync.RWMutex
}

// NewTunnel initializes a Tunnel with the provided SSHConfig, remote host, remote port, and local port settings.
func NewTunnel(config *SSHConfig, remoteHost string, remotePort, localPort int) *Tunnel {
	return &Tunnel{
		config:     config,
		remoteHost: remoteHost,
		remotePort: remotePort,
		localPort:  localPort,
		status:     StatusStopped,
	}
}

// Validate checks if the Tunnel's configuration and parameters are valid, returning an error if any validation fails.
func (t *Tunnel) Validate() error {
	if t.config == nil {
		return fmt.Errorf("config is required")
	}

	if t.remoteHost == "" {
		return fmt.Errorf("remoteHost is required")
	}

	if t.remotePort <= 0 {
		return fmt.Errorf("remotePort must be greater than 0")
	}

	if t.localPort < 0 {
		return fmt.Errorf("localPort must be 0 or greater")
	}

	return nil
}

// setError updates the tunnel's status to error and records the provided error as the last encountered error.
func (t *Tunnel) setError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.status = StatusError
	t.lastError = err
}

// Start initializes and starts the tunnel, setting up the SSH connection and local listener. Returns an error if it fails.
func (t *Tunnel) Start() error {
	t.mu.Lock()

	if t.status == StatusRunning {
		t.mu.Unlock()
		return fmt.Errorf("tunnel is already running")
	}

	t.status = StatusStarting
	t.lastError = nil
	t.mu.Unlock()

	if err := t.Validate(); err != nil {
		t.setError(err)
		return err
	}

	sshClientConfig := &ssh.ClientConfig{
		User:            t.config.User,
		Auth:            t.config.AuthMethods,
		HostKeyCallback: t.config.HostKeyCallback,
		Config: ssh.Config{
			KeyExchanges: []string{
				"diffie-hellman-group-exchange-sha256",
				"diffie-hellman-group14-sha256",
				"diffie-hellman-group14-sha1",
				"curve25519-sha256",
				"curve25519-sha256@libssh.org",
				"ecdh-sha2-nistp256",
				"ecdh-sha2-nistp384",
				"ecdh-sha2-nistp521",
			},
		},
	}

	client, err := ssh.Dial("tcp", t.config.Addr(), sshClientConfig)
	if err != nil {
		err = fmt.Errorf("failed to connect to ssh server: %w", err)
		t.setError(err)
		return err
	}

	listenAddr := fmt.Sprintf("127.0.0.1:%d", t.localPort)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		_ = client.Close()
		err = fmt.Errorf("failed to create local listener: %w", err)
		t.setError(err)
		return err
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port

	t.mu.Lock()
	t.client = client
	t.listener = listener
	t.actualPort = actualPort
	t.status = StatusRunning
	t.done = make(chan struct{})
	t.stats = Stats{StartedAt: time.Now()}
	t.mu.Unlock()

	go t.forward()

	return nil
}

// Stop terminates the tunnel by closing any active connections, freeing resources, and updating the tunnel's status.
func (t *Tunnel) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.status == StatusStopped {
		return nil
	}

	if t.done != nil {
		close(t.done)
	}

	var errs []error
	if t.listener != nil {
		if err := t.listener.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close listener: %w", err))
		}
		t.listener = nil
	}

	if t.client != nil {
		if err := t.client.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close ssh client: %w", err))
		}
		t.client = nil
	}

	t.status = StatusStopped
	t.actualPort = 0
	t.stats = Stats{}

	if len(errs) > 0 {
		return fmt.Errorf("errors stopping tunnel: %v", errs)
	}

	return nil
}

// Restart stops the tunnel if running and then starts it again, returning an error if either operation fails.
func (t *Tunnel) Restart() error {
	if err := t.Stop(); err != nil {
		return fmt.Errorf("failed to stop: %w", err)
	}

	return t.Start()
}

// UpdateConfig updates the tunnel's SSH configuration with the provided config, ensuring thread-safe access.
func (t *Tunnel) UpdateConfig(config *SSHConfig) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.config = config
}

// Status returns the current operational state of the tunnel in a thread-safe manner.
func (t *Tunnel) Status() Status {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

// LastError retrieves the last recorded error encountered by the tunnel in a thread-safe manner.
func (t *Tunnel) LastError() error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.lastError
}

// LocalPort returns the port number being used by the tunnel for local connections, ensuring thread-safe access.
func (t *Tunnel) LocalPort() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.actualPort > 0 {
		return t.actualPort
	}
	return t.localPort
}

// LocalAddr returns the local address and port as a string in the format "127.0.0.1:<port>".
func (t *Tunnel) LocalAddr() string {
	return fmt.Sprintf("127.0.0.1:%d", t.LocalPort())
}

// RemoteAddr retorna o endereço remoto.
func (t *Tunnel) RemoteAddr() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return fmt.Sprintf("%s:%d", t.remoteHost, t.remotePort)
}

// Stats retrieves the statistical data related to network activity for the tunnel in a thread-safe manner.
func (t *Tunnel) Stats() Stats {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.stats
}

// Close gracefully shuts down the tunnel by stopping all active connections and releasing resources.
func (t *Tunnel) Close() error {
	return t.Stop()
}

// forward establishes and manages a connection between a local endpoint and a remote endpoint through the tunnel.
func (t *Tunnel) forward() {
	for {
		select {
		case <-t.done:
			return
		default:
		}

		localConn, err := t.listener.Accept()
		if err != nil {
			select {
			case <-t.done:
				return
			default:
				continue
			}
		}

		t.mu.Lock()
		t.stats.Connections++
		t.stats.ActiveConnections++
		t.mu.Unlock()

		t.mu.RLock()
		remoteAddr := fmt.Sprintf("%s:%d", t.remoteHost, t.remotePort)
		client := t.client
		t.mu.RUnlock()

		remoteConn, err := client.Dial("tcp", remoteAddr)
		if err != nil {
			_ = localConn.Close()
			t.mu.Lock()
			t.stats.ActiveConnections--
			t.mu.Unlock()
			continue
		}

		go t.pipe(localConn, remoteConn)
	}
}

// pipe establishes bidirectional data transfer between local and remote connections and manages connection lifecycle.
func (t *Tunnel) pipe(local, remote net.Conn) {
	defer func() {
		_ = local.Close()
		_ = remote.Close()
		t.mu.Lock()
		t.stats.ActiveConnections--
		t.mu.Unlock()
	}()

	done := make(chan struct{}, 2)

	// Local -> Remote
	go func() {
		n, err := io.Copy(remote, local)
		t.mu.Lock()
		t.stats.BytesOut += n
		t.stats.LastActivity = time.Now()
		if err != nil {
			t.lastError = fmt.Errorf("local->remote copy failed: %w", err)
		}
		t.mu.Unlock()
		done <- struct{}{}
	}()

	// Remote -> Local
	go func() {
		n, err := io.Copy(local, remote)
		t.mu.Lock()
		t.stats.BytesIn += n
		t.stats.LastActivity = time.Now()
		if err != nil {
			t.lastError = fmt.Errorf("remote->local copy failed: %w", err)
		}
		t.mu.Unlock()
		done <- struct{}{}
	}()

	<-done
}
