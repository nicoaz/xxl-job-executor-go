package xxl

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

type testLogger struct{}

func (l *testLogger) Info(format string, a ...interface{})  {}
func (l *testLogger) Error(format string, a ...interface{}) {}

func newTestExecutor() *executor {
	e := newExecutor()
	e.log = &testLogger{}
	e.regList = &taskList{data: make(map[string]*Task)}
	e.runList = &taskList{data: make(map[string]*Task)}
	return e
}

func TestRunTaskDoesNotMutateRegisteredTemplate(t *testing.T) {
	t.Parallel()

	started := make(chan struct{}, 1)
	release := make(chan struct{})

	e := newTestExecutor()
	e.RegTask("tenantJobHandler", func(ctx context.Context, req *RunReq) (int64, string) {
		started <- struct{}{}
		<-release
		return SuccessCode, "ok"
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/run", strings.NewReader(`{"jobId":21,"executorHandler":"tenantJobHandler"}`))

	e.runTask(recorder, request)
	<-started

	registered := e.regList.Get("tenantJobHandler")
	if registered == nil {
		t.Fatalf("expected registered task")
	}
	if registered.Param != nil {
		t.Fatalf("expected registered task template param to stay nil, got %#v", registered.Param)
	}
	if registered.Id != 0 {
		t.Fatalf("expected registered task template id to stay zero, got %d", registered.Id)
	}

	close(release)
}

func TestRunTaskReturnsObjectForUnregisteredHandler(t *testing.T) {
	t.Parallel()

	e := newTestExecutor()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/run", strings.NewReader(`{"jobId":21,"executorHandler":"missing"}`))

	e.runTask(recorder, request)

	body := recorder.Body.String()
	if strings.HasPrefix(body, "[") {
		t.Fatalf("expected object response, got array: %s", body)
	}

	var got res
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("unmarshal response failed: %v, body=%s", err, body)
	}
	if got.Code != FailureCode {
		t.Fatalf("expected failure code, got %d", got.Code)
	}
}

func TestRunTaskReturnsObjectWhenJobAlreadyRunning(t *testing.T) {
	t.Parallel()

	e := newTestExecutor()
	e.RegTask("tenantJobHandler", func(ctx context.Context, req *RunReq) (int64, string) { return SuccessCode, "ok" })
	e.runList.Set("21", &Task{Id: 21})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest("POST", "/run", strings.NewReader(`{"jobId":21,"executorHandler":"tenantJobHandler"}`))

	e.runTask(recorder, request)

	body := recorder.Body.String()
	if strings.HasPrefix(body, "[") {
		t.Fatalf("expected object response, got array: %s", body)
	}

	var got res
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("unmarshal response failed: %v, body=%s", err, body)
	}
	if got.Code != FailureCode {
		t.Fatalf("expected failure code, got %d", got.Code)
	}
}

func TestTaskRunUsesReturnedCode(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		fn       TaskFunc
		wantCode int64
		wantMsg  string
	}{
		{
			name: "success",
			fn: func(ctx context.Context, req *RunReq) (int64, string) {
				return SuccessCode, "ok"
			},
			wantCode: SuccessCode,
			wantMsg:  "ok",
		},
		{
			name: "failure",
			fn: func(ctx context.Context, req *RunReq) (int64, string) {
				return FailureCode, "failed"
			},
			wantCode: FailureCode,
			wantMsg:  "failed",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			task := (&Task{fn: tc.fn, log: &testLogger{}}).Clone()
			task.Ext, task.Cancel = context.WithCancel(context.Background())
			defer task.Cancel()

			var (
				gotCode int64
				gotMsg  string
			)
			task.Run(func(code int64, msg string) {
				gotCode = code
				gotMsg = msg
			})

			if gotCode != tc.wantCode {
				t.Fatalf("expected code %d, got %d", tc.wantCode, gotCode)
			}
			if gotMsg != tc.wantMsg {
				t.Fatalf("expected msg %q, got %q", tc.wantMsg, gotMsg)
			}
		})
	}
}
