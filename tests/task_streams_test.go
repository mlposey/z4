package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/google/uuid"
	"github.com/mlposey/z4/pkg/z4"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/tests/util"
	"go.uber.org/multierr"
	pb "google.golang.org/protobuf/proto"
	"sync"
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
		sc.Step(`^I subscribe to tasks in the "([^"]*)" queue`, ts.iSubscribeToTasksInTheQueue)
		sc.Step(`^I subscribe to tasks in the "([^"]*)" queue after a (\d+) second delay$`, ts.iSubscribeToTasksInTheQueueAfterASecondDelay)
	})

}

type taskStreams struct {
	server        *util.LocalServer
	serverPort    int
	producer      *z4.UnaryProducer
	consumer      *z4.Consumer
	taskRequest   *proto.PushTaskRequest
	createdTask   *proto.Task
	receivedTasks []*proto.Task
	client        *z4.Client
	taskMu        sync.Mutex
	ctx           context.Context
	ctxCancel     context.CancelFunc
}

func (ts *taskStreams) setupSuite() error {
	ts.serverPort = 6355
	ts.server = util.NewLocalServer(ts.serverPort)
	ts.taskRequest = nil
	ts.createdTask = nil
	ts.receivedTasks = nil
	ts.consumer = nil

	err := ts.server.Start()
	if err != nil {
		return err
	}

	ts.ctx, ts.ctxCancel = context.WithCancel(context.Background())

	target := fmt.Sprintf("localhost:%d", ts.serverPort)
	ts.client, err = z4.NewClient(ts.ctx, z4.ClientOptions{
		Addrs: []string{target},
	})
	if err != nil {
		return err
	}

	ts.producer = ts.client.UnaryProducer()
	return err
}

func (ts *taskStreams) teardownSuite() error {
	var err1 error
	if ts.client != nil {
		err1 = ts.client.Close()
		ts.ctxCancel()
	}
	err2 := ts.server.Stop()
	return multierr.Combine(err1, err2)
}

func (ts *taskStreams) afterSecondsIShouldReceiveTheSameTask(arg1 int) error {
	time.Sleep(time.Duration(arg1) * time.Second)

	ts.taskMu.Lock()
	defer ts.taskMu.Unlock()
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

	ts.taskMu.Lock()
	defer ts.taskMu.Unlock()
	if len(ts.receivedTasks) != arg2 {
		return fmt.Errorf("expected %d tasks; got %d", arg2, len(ts.receivedTasks))
	}
	return nil
}

func (ts *taskStreams) consumeTaskStream(queue string) error {
	var err error
	ts.consumer, err = ts.client.Consumer(ts.ctx, queue)
	if err != nil {
		return err
	}

	go func() {
		_ = ts.consumer.Consume(func(m z4.Message) error {

			ts.taskMu.Lock()
			ts.receivedTasks = append(ts.receivedTasks, m.Task())
			ts.taskMu.Unlock()

			m.Ack()
			return nil
		})
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
		RequestId: uuid.New().String(),
		Queue:     taskDef["queue"].(string),
		Schedule: &proto.PushTaskRequest_TtsSeconds{
			TtsSeconds: int64(taskDef["tts_seconds"].(float64)),
		},
	}
	task, err := ts.producer.CreateTask(ts.ctx, ts.taskRequest)
	ts.createdTask = task
	return err
}

func (ts *taskStreams) iSubscribeToTasksInTheQueue(arg1 string) error {
	return ts.consumeTaskStream(arg1)
}

func (ts *taskStreams) iSubscribeToTasksInTheQueueAfterASecondDelay(arg1 string, arg2 int) error {
	time.Sleep(time.Duration(arg2) * time.Second)
	return ts.consumeTaskStream(arg1)
}
