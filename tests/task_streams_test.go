package tests

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/mlposey/z4/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

type taskStreams struct {
	dataDir    string
	server     *os.Process
	serverPort int
	client     proto.CollectionClient
}

func (ts *taskStreams) setupSuite() error {
	err := new(error)
	ts.doIfOK(err, ts.createNewDBFolder)
	ts.doIfOK(err, ts.startServer)
	ts.doIfOK(err, ts.createClient)
	return *err
}

func (ts *taskStreams) doIfOK(err *error, do func() error) {
	if *err == nil {
		*err = do()
	}
}

func (ts *taskStreams) createNewDBFolder() error {
	ts.dataDir = "test_task_streams_datadir"
	err := os.RemoveAll(ts.dataDir)
	if err != nil {
		return err
	}
	return os.MkdirAll(ts.dataDir, os.ModePerm)
}

func (ts *taskStreams) startServer() error {
	os.Setenv("Z4_DB_DATA_DIR", ts.dataDir)
	os.Setenv("Z4_PORT", fmt.Sprint(ts.serverPort))
	os.Setenv("Z4_DEBUG_LOGGING_ENABLED", "true")

	cmd := exec.Command("bash", "-c", "go run ../cmd/server/*.go")

	// used get print logs later on in this method
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		return err
	}

	// helpful for killing the server / child process
	// essentially assigns all processes we spawn to this group; we will later
	// kill the entire group at one time
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err = cmd.Start()
	if err != nil {
		return err
	}
	ready := make(chan bool)
	go func() {
		for {
			tmp := make([]byte, 1024)
			_, e := stdout.Read(tmp)
			fmt.Print("[server]: " + string(tmp))
			if e != nil {
				break
			}

			if strings.Contains(string(tmp), "listening for connections") {
				ready <- true
			}
		}
	}()

	select {
	case <-ready:
		time.Sleep(time.Millisecond * 100)
	case <-time.NewTimer(time.Second * 10).C:
		return errors.New("server not started before deadline")
	}

	ts.server = cmd.Process
	return nil
}

func (ts *taskStreams) createClient() error {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", ts.serverPort), opts...)
	if err != nil {
		return err
	}
	ts.client = proto.NewCollectionClient(conn)
	return nil
}

func (ts *taskStreams) teardownSuite() error {
	if ts.server != nil {
		err := syscall.Kill(-ts.server.Pid, syscall.SIGKILL)
		if err != nil {
			return err
		}
		_, err = ts.server.Wait()
		return err
	}
	return nil
}

func (ts *taskStreams) afterSecondIShouldReceiveTheSameTask(arg1 int) error {
	return nil
}

func (ts *taskStreams) afterSecondsIShouldReceiveTasks(arg1, arg2 int) error {
	return nil
}

func (ts *taskStreams) iBeginStreamingAfterASecondDelay(arg1 int) error {
	return nil
}

func (ts *taskStreams) iHaveCreatedTheTask(arg1 *godog.DocString) error {
	var taskDef map[string]interface{}
	err := json.Unmarshal([]byte(arg1.Content), &taskDef)
	if err != nil {
		return err
	}

	req := &proto.CreateTaskRequest{
		RequestId:  taskDef["request_id"].(string),
		Namespace:  taskDef["namespace"].(string),
		TtsSeconds: int64(taskDef["tts_seconds"].(float64)),
	}
	_, err = ts.client.CreateTask(context.Background(), req)
	return err
}

func InitializeScenario(sc *godog.ScenarioContext) {
	ts := &taskStreams{serverPort: 6355}

	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		return ctx, ts.setupSuite()
	})
	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		return ctx, ts.teardownSuite()
	})

	sc.Step(`^after (\d+) second I should receive the same task$`, ts.afterSecondIShouldReceiveTheSameTask)
	sc.Step(`^after (\d+) seconds I should receive (\d+) tasks$`, ts.afterSecondsIShouldReceiveTasks)
	sc.Step(`^I begin streaming after a (\d+) second delay$`, ts.iBeginStreamingAfterASecondDelay)
	sc.Step(`^I have created the task:$`, ts.iHaveCreatedTheTask)
}
