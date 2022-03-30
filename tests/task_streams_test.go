package tests

import (
	"context"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/mlposey/z4/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"os/exec"
	"time"
)

type taskStreams struct {
	dataDir    string
	server     *os.Process
	serverPort int
	client     proto.CollectionClient
}

func (ts *taskStreams) reset() error {
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

	cmd := exec.Command("go", "run", "../cmd/server/*.go")
	err := cmd.Start()
	if err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 100)
	ts.server = cmd.Process
	return nil
}

func (ts *taskStreams) stopServer() error {
	if ts.server != nil {
		return ts.server.Kill()
	}
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
	return nil
}

func InitializeScenario(sc *godog.ScenarioContext) {
	ts := &taskStreams{serverPort: 6355}

	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		return ctx, ts.reset()
	})
	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		return ctx, ts.stopServer()
	})

	sc.Step(`^after (\d+) second I should receive the same task$`, ts.afterSecondIShouldReceiveTheSameTask)
	sc.Step(`^after (\d+) seconds I should receive (\d+) tasks$`, ts.afterSecondsIShouldReceiveTasks)
	sc.Step(`^I begin streaming after a (\d+) second delay$`, ts.iBeginStreamingAfterASecondDelay)
	sc.Step(`^I have created the task:$`, ts.iHaveCreatedTheTask)
}
