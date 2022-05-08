package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/tests/util"
	"github.com/segmentio/ksuid"
	"go.uber.org/multierr"
	pb "google.golang.org/protobuf/proto"
	"testing"
	"time"
)

// TestTaskStreaming runs the Task Streaming test suite.
func TestTaskStreaming(t *testing.T) {
	ts := &taskStreams{}
	runTestSuite(t, func(sc *godog.ScenarioContext) {
		sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			return ctx, ts.setupSuite()
		})
		sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
			return ctx, ts.teardownSuite()
		})

		sc.Step(`^after (\d+) seconds I should receive the same task$`, ts.afterSecondsIShouldReceiveTheSameTask)
		sc.Step(`^after (\d+) seconds I should receive (\d+) tasks$`, ts.afterSecondsIShouldReceiveTasks)
		sc.Step(`^I have created the task:$`, ts.iHaveCreatedTheTask)
		sc.Step(`^I subscribe to tasks in the "([^"]*)" namespace$`, ts.iSubscribeToTasksInTheNamespace)
		sc.Step(`^I subscribe to tasks in the "([^"]*)" namespace after a (\d+) second delay$`, ts.iSubscribeToTasksInTheNamespaceAfterASecondDelay)
	})

}

type taskStreams struct {
	server        *util.LocalServer
	serverPort    int
	client        *util.Client
	taskRequest   *proto.PushTaskRequest
	createdTask   *proto.Task
	receivedTasks []*proto.Task
}

func (ts *taskStreams) setupSuite() error {
	ts.serverPort = 6355
	ts.server = util.NewLocalServer(ts.serverPort)
	ts.taskRequest = nil
	ts.createdTask = nil
	ts.receivedTasks = nil

	if err := ts.server.Start(); err != nil {
		return err
	}

	c, err := util.NewClient("localhost", ts.serverPort)
	if err == nil {
		ts.client = c
	}
	return err
}

func (ts *taskStreams) teardownSuite() error {
	err1 := ts.client.Close()
	err2 := ts.server.Stop()
	return multierr.Combine(err1, err2)
}

func (ts *taskStreams) afterSecondsIShouldReceiveTheSameTask(arg1 int) error {
	time.Sleep(time.Duration(arg1) * time.Second)
	if len(ts.receivedTasks) != 1 {
		return fmt.Errorf("expected to receive the input task back; got %d tasks", len(ts.receivedTasks))
	}
	ok := pb.Equal(ts.receivedTasks[0], ts.createdTask)
	if !ok {
		return fmt.Errorf("tasks not equal; expected %v; got %v", ts.createdTask, ts.receivedTasks[0])
	}
	return nil
}

func (ts *taskStreams) afterSecondsIShouldReceiveTasks(arg1, arg2 int) error {
	time.Sleep(time.Duration(arg1) * time.Second)
	if len(ts.receivedTasks) != arg2 {
		return fmt.Errorf("expected %d tasks; got %d", arg2, len(ts.receivedTasks))
	}
	return nil
}

func (ts *taskStreams) iBeginStreamingAfterASecondDelay(arg1 int) error {
	time.Sleep(time.Duration(arg1) * time.Second)
	return ts.consumeTaskStream("")
}

func (ts *taskStreams) consumeTaskStream(namespace string) error {
	stream, err := ts.client.PullTasks(ksuid.New().String(), namespace)
	if err != nil {
		return err
	}
	responses, err := stream.Listen()
	if err != nil {
		return err
	}

	go func() {
		for res := range responses {
			if res.Error != nil {
				fmt.Println(res.Error)
				break
			}

			ts.receivedTasks = append(ts.receivedTasks, res.Task)
			res.Ack()
		}
	}()
	return nil
}

func (ts *taskStreams) iHaveCreatedTheTask(arg1 *godog.DocString) error {
	var taskDef map[string]interface{}
	err := json.Unmarshal([]byte(arg1.Content), &taskDef)
	if err != nil {
		return err
	}

	ts.taskRequest = &proto.PushTaskRequest{
		RequestId: ksuid.New().String(),
		Namespace: taskDef["namespace"].(string),
		Schedule: &proto.PushTaskRequest_TtsSeconds{
			TtsSeconds: int64(taskDef["tts_seconds"].(float64)),
		},
	}
	task, err := ts.client.Push(ts.taskRequest)
	ts.createdTask = task.GetTask()
	return err
}

func (ts *taskStreams) iSubscribeToTasksInTheNamespace(arg1 string) error {
	return ts.consumeTaskStream(arg1)
}

func (ts *taskStreams) iSubscribeToTasksInTheNamespaceAfterASecondDelay(arg1 string, arg2 int) error {
	time.Sleep(time.Duration(arg2) * time.Second)
	return ts.consumeTaskStream(arg1)
}
