package tempredis

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const (
	// ready is the string redis-server prints to stdout after starting
	// successfully.
	ready = "The server is now ready to accept connections"
)

// Server encapsulates the configuration, starting, and stopping of a single
// redis-server process that is reachable via a local Unix socket.
type Server struct {
	dir       string
	config    Config
	cmd       *exec.Cmd
	stdout    io.Reader
	stdoutBuf bytes.Buffer
	stderr    io.Reader
}

// Start initiates a new redis-server process configured with the given
// configuration. redis-server will listen on a temporary local Unix socket. An
// error is returned if redis-server is unable to successfully start for any
// reason.
func Start(config Config) (server *Server, err error) {
	if config == nil {
		config = Config{}
	}

	dir, err := ioutil.TempDir(os.TempDir(), "tempredis")
	if err != nil {
		return nil, err
	}

	if _, ok := config["unixsocket"]; !ok {
		config["unixsocket"] = fmt.Sprintf("%s/%s", dir, "redis.sock")
	}
	if _, ok := config["port"]; !ok {
		config["port"] = "0"
	}

	server = &Server{
		dir:    dir,
		config: config,
	}
	err = server.start()
	if err != nil {
		return server, err
	}

	fmt.Println("!!!")

	// Block until Redis is ready to accept connections.
	err = server.waitFor(ready)

	fmt.Println("@@@")

	return server, err
}

func (s *Server) start() (err error) {
	if s.cmd != nil {
		return fmt.Errorf("redis-server has already been started")
	}

	s.cmd = exec.Command("redis-server", "-")

	stdin, _ := s.cmd.StdinPipe()
	s.stdout, _ = s.cmd.StdoutPipe()
	s.stderr, _ = s.cmd.StderrPipe()

	err = s.cmd.Start()
	if err == nil {
		err = writeConfig(s.config, stdin)
	}

	return err
}

func writeConfig(config Config, w io.WriteCloser) (err error) {
	for key, value := range config {
		if value == "" {
			value = "\"\""
		}
		_, err = fmt.Fprintf(w, "%s %s\n", key, value)
		if err != nil {
			return err
		}
	}
	return w.Close()
}

// waitFor blocks until redis-server prints the given string to stdout.
func (s *Server) waitFor(search string) (err error) {
	var line string

	scanner := bufio.NewScanner(s.stdout)
	for scanner.Scan() {
		line = scanner.Text()
		fmt.Println(line)
		fmt.Fprintf(&s.stdoutBuf, "%s\n", line)
		if strings.Contains(line, search) {
			return nil
		}
	}
	err = scanner.Err()
	if err == nil {
		err = io.EOF
	}
	return err
}

// Socket returns the full path to the local redis-server Unix socket.
func (s *Server) Socket() string {
	return s.config.Socket()
}

// Stdout blocks until redis-server returns and then returns the full stdout
// output.
func (s *Server) Stdout() string {
	io.Copy(&s.stdoutBuf, s.stdout)
	return s.stdoutBuf.String()
}

// Stderr blocks until redis-server returns and then returns the full stdout
// output.
func (s *Server) Stderr() string {
	bytes, _ := ioutil.ReadAll(s.stderr)
	return string(bytes)
}

// Term gracefully shuts down redis-server. It returns an error if redis-server
// fails to terminate.
func (s *Server) Term() (err error) {
	return s.signalAndCleanup(syscall.SIGTERM)
}

// Kill forcefully shuts down redis-server. It returns an error if redis-server
// fails to die.
func (s *Server) Kill() (err error) {
	return s.signalAndCleanup(syscall.SIGKILL)
}

func (s *Server) signalAndCleanup(sig syscall.Signal) error {
	s.cmd.Process.Signal(sig)
	_, err := s.cmd.Process.Wait()
	os.RemoveAll(s.dir)
	return err
}
