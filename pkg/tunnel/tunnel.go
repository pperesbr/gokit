package tunnel

import (
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/ssh"
)

// Tunnel represents an SSH tunnel that forwards traffic between a local port and a remote address.
// It encapsulates the SSH configuration, client connection, and listener used for port forwarding.
type Tunnel struct {
	config     *SSHConfig
	client     *ssh.Client
	listener   net.Listener
	remoteAddr string
	localPort  int
	done       chan struct{}
}

// NewTunnel creates and initializes a new SSH tunnel for port forwarding using the given configuration and addresses.
func NewTunnel(config *SSHConfig, remoteHost string, remotePort, localPort int) (*Tunnel, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	if remoteHost == "" {
		return nil, fmt.Errorf("remoteHost is required")
	}

	if remotePort <= 0 {
		return nil, fmt.Errorf("remotePort must be greater than 0")
	}

	sshClientConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            []ssh.AuthMethod{config.AuthMethod},
		HostKeyCallback: config.HostKeyCallback,
	}

	client, err := ssh.Dial("tcp", config.Addr(), sshClientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ssh server: %w", err)
	}

	listenAddr := fmt.Sprintf("127.0.0.1:%d", localPort)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to create local listener: %w", err)
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port

	t := &Tunnel{
		config:     config,
		client:     client,
		listener:   listener,
		remoteAddr: fmt.Sprintf("%s:%d", remoteHost, remotePort),
		localPort:  actualPort,
		done:       make(chan struct{}),
	}

	go t.forward()

	return t, nil
}

// forward manages the SSH tunnel by accepting incoming local connections and forwarding them to the remote address.
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

		remoteConn, err := t.client.Dial("tcp", t.remoteAddr)
		if err != nil {
			_ = localConn.Close()
			continue
		}

		go t.pipe(localConn, remoteConn)
	}
}

// pipe copies data bidirectionally between two network connections and closes them when finished.
func (t *Tunnel) pipe(local, remote net.Conn) {
	defer func() { _ = local.Close() }()
	defer func() { _ = remote.Close() }()

	done := make(chan struct{}, 2)

	go func() {
		_, _ = io.Copy(remote, local)
		done <- struct{}{}
	}()

	go func() {
		_, _ = io.Copy(local, remote)
		done <- struct{}{}
	}()

	<-done
}

// LocalPort returns the local port number where the tunnel is actively listening for incoming connections.
func (t *Tunnel) LocalPort() int {
	return t.localPort
}

// LocalAddr returns the local address of the tunnel in the format "127.0.0.1:<localPort>".
func (t *Tunnel) LocalAddr() string {
	return fmt.Sprintf("127.0.0.1:%d", t.localPort)
}

func (t *Tunnel) RemoteAddr() string {
	return t.remoteAddr
}

// Close gracefully shuts down the SSH tunnel by closing the listener, SSH client, and stopping any active forwarding.
func (t *Tunnel) Close() error {
	close(t.done)

	var errs []error

	if err := t.listener.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close listener: %w", err))
	}

	if err := t.client.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close ssh client: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing tunnel: %v", errs)
	}

	return nil
}
