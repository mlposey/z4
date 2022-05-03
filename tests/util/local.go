package util

import (
	"errors"
	"fmt"
	"github.com/segmentio/ksuid"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// LocalServer manages a local instance of a z4 server.
type LocalServer struct {
	dbDataDir   string
	peerDataDir string
	server      *os.Process
	serverPort  int
	appOutput   io.ReadCloser
	appHandle   *exec.Cmd
}

func NewLocalServer(port int) *LocalServer {
	return &LocalServer{
		dbDataDir:   "/tmp/" + ksuid.New().String(),
		peerDataDir: "/tmp/" + ksuid.New().String(),
		serverPort:  port,
	}
}

// Start starts the application in a background process.
func (ls *LocalServer) Start() error {
	if err := ls.startApp(); err != nil {
		return err
	}
	return ls.waitForInit()
}

// starts the application in a new process
func (ls *LocalServer) startApp() error {
	_ = os.Setenv("Z4_DB_DATA_DIR", ls.dbDataDir)
	_ = os.Setenv("Z4_SERVICE_PORT", fmt.Sprint(ls.serverPort))
	_ = os.Setenv("Z4_DEBUG_LOGGING_ENABLED", "true")
	_ = os.Setenv("Z4_PEER_ID", "godog")
	_ = os.Setenv("Z4_BOOTSTRAP_CLUSTER", "true")
	_ = os.Setenv("Z4_PEER_DATA_DIR", ls.peerDataDir)

	ls.appHandle = exec.Command("bash", "-c", "go run ../cmd/server/*.go")

	// used get print logs later on in this method
	stdout, err := ls.appHandle.StdoutPipe()
	ls.appHandle.Stderr = ls.appHandle.Stdout
	if err != nil {
		return err
	}
	ls.appOutput = stdout

	// helpful for killing the server / child process
	// essentially assigns all processes we spawn to this group; we will later
	// kill the entire group at one time
	ls.appHandle.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return ls.appHandle.Start()
}

// waits for the application to finish startup
func (ls *LocalServer) waitForInit() error {
	ready := make(chan bool)
	go func() {
		// Read the app output until we think it is ready for connections.
		// TODO: Use the health check endpoint instead.
		for {
			tmp := make([]byte, 1024)
			_, e := ls.appOutput.Read(tmp)
			fmt.Print("[server]: " + string(tmp))
			if e != nil {
				break
			}

			if strings.Contains(string(tmp), "entering leader state") {
				ready <- true
			}
		}
	}()

	select {
	case <-ready:
		time.Sleep(time.Millisecond * 100)
	case <-time.NewTimer(time.Second * 30).C:
		return errors.New("server not started before deadline")
	}

	ls.server = ls.appHandle.Process
	return nil
}

// Stop tears down the application.
func (ls *LocalServer) Stop() error {
	if ls.server == nil {
		fmt.Println("cannot stop server: not started")
		return nil
	}

	err := syscall.Kill(-ls.server.Pid, syscall.SIGKILL)
	if err != nil {
		return err
	}
	_, err = ls.server.Wait()
	return err
}
