package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/mlposey/z4/proto"
	"github.com/mlposey/z4/tests/util"
	"github.com/segmentio/ksuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "google.golang.org/protobuf/proto"
	"io"
	"log"
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
		sc.Step(`^I begin streaming after a (\d+) second delay$`, ts.iBeginStreamingAfterASecondDelay)
		sc.Step(`^I have created the task:$`, ts.iHaveCreatedTheTask)
	})

}

type taskStreams struct {
	server        *util.LocalServer
	serverPort    int
	client        proto.QueueClient
	taskRequest   *proto.PushTaskRequest
	createdTask   *proto.Task
	receivedTasks []*proto.Task
}

func (ts *taskStreams) setupSuite() error {
	ts.serverPort = 6355
	ts.server = util.NewLocalServer(ts.serverPort)
	ts.client = nil
	ts.taskRequest = nil
	ts.createdTask = nil
	ts.receivedTasks = nil

	err := new(error)
	ts.doIfOK(err, ts.server.Start)
	ts.doIfOK(err, ts.createClient)
	return *err
}

func (ts *taskStreams) doIfOK(err *error, do func() error) {
	if *err == nil {
		*err = do()
	}
}

func (ts *taskStreams) createClient() error {
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", ts.serverPort), opts...)
	if err != nil {
		return err
	}
	ts.client = proto.NewQueueClient(conn)
	return nil
}

func (ts *taskStreams) teardownSuite() error {
	return ts.server.Stop()
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
	return ts.consumeTaskStream()
}

func (ts *taskStreams) consumeTaskStream() error {
	stream, err := ts.client.Pull(context.Background())
	if err != nil {
		return err
	}

	err = stream.Send(&proto.PullRequest{
		Request: &proto.PullRequest_StartReq{
			StartReq: &proto.StartStreamRequest{
				// TODO: Generate this or take it from the gherkin.
				RequestId: ksuid.New().String(),
				// TODO: Supply namespace in gherkin so we can test failure scenarios.
				Namespace: ts.taskRequest.GetNamespace(),
			},
		},
	})
	if err != nil {
		return err
	}

	go func() {
		for {
			task, err := stream.Recv()
			if err == io.EOF {
				fmt.Println("stream closed by server")
				break
			}
			if err != nil {
				// this is expected when we send the kill signal to the server
				fmt.Printf("stream error: %v\n", err)
				break
			}

			ts.receivedTasks = append(ts.receivedTasks, task)

			err = stream.Send(&proto.PullRequest{
				Request: &proto.PullRequest_Ack{
					Ack: &proto.Ack{
						TaskId:    task.GetId(),
						Namespace: task.GetNamespace(),
					},
				},
			})
			if err != nil {
				log.Fatalf("ack failed: %v", err)
			}
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
		RequestId:  taskDef["request_id"].(string),
		Namespace:  taskDef["namespace"].(string),
		TtsSeconds: int64(taskDef["tts_seconds"].(float64)),
	}
	task, err := ts.client.Push(context.Background(), ts.taskRequest)
	ts.createdTask = task.GetTask()
	return err
}
