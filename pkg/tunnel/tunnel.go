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

// Tunnel represents an SSH tunnel for forwarding network traffic between a local and remote address over a secure connection.
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

// NewTunnel creates a new SSH Tunnel to forward traffic between a local port and a remote host with the given configuration.
// config specifies the SSH connection parameters, remoteHost is the target host, remotePort is the target port on the remote host,
// and localPort specifies the local port to listen on (0 to auto-assign). Returns a Tunnel instance.
func NewTunnel(config *SSHConfig, remoteHost string, remotePort, localPort int) *Tunnel {
	return &Tunnel{
		config:     config,
		remoteHost: remoteHost,
		remotePort: remotePort,
		localPort:  localPort,
		status:     StatusStopped,
	}
}

// Validate checks the Tunnel configuration for required fields and ensures values meet expected constraints.
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

// setError sets the tunnel status to an error state and records the provided error.
func (t *Tunnel) setError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.status = StatusError
	t.lastError = err
}

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
		Auth:            []ssh.AuthMethod{t.config.AuthMethod},
		HostKeyCallback: t.config.HostKeyCallback,
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

// Stop gracefully shuts down the tunnel by closing connections, releasing resources, and updating the tunnel's status.
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

// Restart stops the tunnel if running, then attempts to start it again, returning an error if either operation fails.
func (t *Tunnel) Restart() error {
	if err := t.Stop(); err != nil {
		return fmt.Errorf("failed to stop: %w", err)
	}

	return t.Start()
}

// UpdateConfig updates the SSH configuration of the tunnel safely with locking to ensure thread safety.
func (t *Tunnel) UpdateConfig(config *SSHConfig) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.config = config
}

// Status returns the current operational status of the tunnel, such as running, stopped, or in error state.
func (t *Tunnel) Status() Status {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status
}

// LastError returns the last error encountered by the tunnel, or nil if no error has occurred.
func (t *Tunnel) LastError() error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.lastError
}

// LocalPort returns the local port that the tunnel is using for forwarding traffic, or the assigned port if set to 0.
func (t *Tunnel) LocalPort() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.actualPort > 0 {
		return t.actualPort
	}
	return t.localPort
}

// LocalAddr returns the local address that the tunnel is listening on, formatted as a string in the "host:port" format.
func (t *Tunnel) LocalAddr() string {
	return fmt.Sprintf("127.0.0.1:%d", t.LocalPort())
}

// RemoteAddr returns the remote address the tunnel is connected to, formatted as "host:port".
func (t *Tunnel) RemoteAddr() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return fmt.Sprintf("%s:%d", t.remoteHost, t.remotePort)
}

// forward establishes bidirectional traffic forwarding between local and remote connections over an SSH tunnel.
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

// pipe transfers bidirectional data between a local and remote connection until one side closes or encounters an error.
func (t *Tunnel) pipe(local, remote net.Conn) {
	defer func() {
		_ = local.Close()
		_ = remote.Close()
		t.mu.Lock()
		t.stats.ActiveConnections--
		t.mu.Unlock()
	}()

	done := make(chan struct{}, 2)

	// Local -> Remote (out)
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

	// Remote -> Local (in)
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

// Stats return a snapshot of the tunnel's statistical data, including connection counts and traffic metrics.
func (t *Tunnel) Stats() Stats {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.stats
}

// Close gracefully shuts down the tunnel by invoking the Stop method to release resources and terminate connections.
func (t *Tunnel) Close() error {
	return t.Stop()
}
