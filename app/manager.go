// +build windows

package app

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kardianos/osext"
)

type (
	Manager struct {
		TLSConfig      *tls.Config
		BindEndpoint   string
		connections    chan net.Conn
		maxConnections int
		close          chan bool
	}
	API struct {
		ScriptsPath string
		PSOpts      string
		PSPath      string
	}
	APIRequest struct {
		File   string
		Params map[string]string
	}
	APIResponse struct {
		Data string
	}
)

func (a *API) Execute(req *APIRequest, res *APIResponse) error {
	// Do the things
	filePath := filepath.Join(a.ScriptsPath, req.File)

	var args []string

	for _, v := range strings.Split(a.PSOpts, " ") {
		args = append(args, v)
	}

	args = append(args, "-File")
	args = append(args, filePath)

	for k, v := range req.Params {
		args = append(args, fmt.Sprintf("-%s", k))
		args = append(args, v)
	}

	fmt.Printf("exec.Command(%s, %q)\n", a.PSPath, args)
	data, err := exec.Command(a.PSPath, args...).Output()
	if err != nil {
		return fmt.Errorf("Failed to execute %s -> %s", req.File, err)
	}
	fmt.Printf("API Response: %s\n", string(data))
	*res = APIResponse{Data: string(data)}
	return nil
}

func New(c *Config) (*Manager, error) {
	m := new(Manager)
	a := new(API)

	if filepath.IsAbs(c.ScriptsPath) == false {
		cwd, err := osext.ExecutableFolder()
		if err != nil {
			return nil, fmt.Errorf("Failed to get current working directory -> %s", err)
		}
		a.ScriptsPath = filepath.Join(cwd, c.ScriptsPath)
	} else {
		a.ScriptsPath = c.ScriptsPath
	}

	psPath, err := interpolateEnvironmentVariables(c.PSPath)
	if err != nil {
		return nil, err
	}
	a.PSPath = psPath
	a.PSOpts = c.PSOpts

	rpc.Register(a)

	if err := m.newTLSConfig(c); err != nil {
		return nil, err
	}

	if err := m.setBindEndpoint(c); err != nil {
		return nil, err
	}

	m.maxConnections = c.MaxConnections

	return m, nil
}

func interpolateEnvironmentVariables(data string) (string, error) {
	b := bytes.NewBufferString(data)
	l := len(data)

	var result []byte
	for i := 0; i < l; i++ {
		curByte, err := b.ReadByte()
		if err != nil {
			return "", err
		}
		if curByte == '%' {
			var name []byte
			for x := i + 1; x < l; x++ {
				innerByte, err := b.ReadByte()
				if err != nil {
					return "", err
				}
				if innerByte == '%' {
					i = x
					break
				}
				name = append(name, innerByte)
			}
			result = append(result, []byte(os.Getenv(string(name)))...)
			continue
		}
		result = append(result, curByte)
	}
	return string(result), nil
}

func (m *Manager) Run() error {
	m.connections = make(chan net.Conn, m.maxConnections)
	m.close = make(chan bool, 1)

	go m.listen()

	for {
		select {
		case c := <-m.connections:
			go func(c net.Conn) {
				fmt.Printf("Processing RPC request for %s\n", c.RemoteAddr().String())
				rpc.ServeConn(c)
				c.Close()
			}(c)
		case <-m.close:
			// Close all pending connections
			fmt.Printf("Closing all existing connections..\n")
			for c := range m.connections {
				c.Close()
			}
			break
		}
	}
}

func (m *Manager) listen() error {
	fmt.Printf("Starting to listen on %s\n", m.BindEndpoint)
	l, err := tls.Listen("tcp", m.BindEndpoint, m.TLSConfig)
	if err != nil {
		return fmt.Errorf("Failed to start listening on %s -> %s", m.BindEndpoint, err)
	}
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Printf("Error while accepting connection -> %s\n", err)
		}
		fmt.Printf("Accepted connection from %s\n", c.RemoteAddr().String())
		m.connections <- c
	}
}

func (m *Manager) Close() {
	m.close <- true
}

func (m *Manager) setBindEndpoint(config *Config) error {
	if len(config.BindAddr) == 0 ||
		len(config.BindPort) == 0 {
		return fmt.Errorf("bind_addr and bind_port must be specified in the configuration yaml")
	}

	var b bytes.Buffer
	b.WriteString(config.BindAddr)
	b.WriteString(":")
	b.WriteString(config.BindPort)

	m.BindEndpoint = b.String()
	return nil
}

func (m *Manager) newTLSConfig(config *Config) error {
	cert, err := tls.LoadX509KeyPair(config.CrtPath, config.KeyPath)
	if err != nil {
		return err
	}
	if len(cert.Certificate) != 2 {
		return fmt.Errorf("CRT file should contain 2 certificates, Server and CA certificate")
	}
	ca, err := x509.ParseCertificate(cert.Certificate[1])
	if err != nil {
		return fmt.Errorf("Unable to parse CA certificate -> %s", err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(ca)

	for i, certificate := range cert.Certificate {
		pCert, err := x509.ParseCertificate(certificate)
		if err != nil {
			fmt.Printf("Failed to parse certificate[%d] -> %s\n", i, err)
			continue
		}
		fmt.Printf("[%d] CN: %s Org: %s Serial: %s\n", i, pCert.Subject.CommonName, pCert.Subject.Organization[0], pCert.Subject.SerialNumber)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    pool,
	}
	m.TLSConfig = tlsConfig
	return nil
}
